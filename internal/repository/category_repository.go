package repository

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrCategoryCycle signals that a reparent would create (or has found) a
// cycle in the category hierarchy.
var ErrCategoryCycle = errors.New("category cycle detected")

// maxAncestorHops guards hierarchy walks against corrupt parent chains.
const maxAncestorHops = 100

type CategoryRepository interface {
	Repository[entity.Category]
	GetActiveByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	List(ctx context.Context, activeOnly bool, page, limit int) ([]entity.Category, int64, error)
	CountChildren(ctx context.Context, id uuid.UUID) (int64, error)
	CountActiveProducts(ctx context.Context, id uuid.UUID) (int64, error)
	UpdateHierarchy(ctx context.Context, id uuid.UUID, fields map[string]any, newParentID *uuid.UUID, newLevel int) error
}

type categoryRepository struct {
	RepositoryImpl[entity.Category]
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		RepositoryImpl: RepositoryImpl[entity.Category]{db: db},
		db:             db,
	}
}

func (r *categoryRepository) GetActiveByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var cat entity.Category
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&cat).Error
	return &cat, err
}

func (r *categoryRepository) List(ctx context.Context, activeOnly bool, page, limit int) ([]entity.Category, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.Category{})
	if activeOnly {
		tx = tx.Where("is_active = ?", true)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	cats := make([]entity.Category, 0)
	if err := tx.Order("sort_order ASC, created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&cats).Error; err != nil {
		return nil, 0, err
	}
	return cats, total, nil
}

func (r *categoryRepository) CountChildren(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Category{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

// CountActiveProducts counts products that are live: status active AND
// is_active true (GORM adds deleted_at IS NULL automatically).
func (r *categoryRepository) CountActiveProducts(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("category_id = ? AND status = ? AND is_active = ?", id, entity.ProductStatusActive, true).
		Count(&count).Error
	return count, err
}

// UpdateHierarchy applies fields and, when the parent changes, re-validates
// the hierarchy and recomputes descendant levels — all in one transaction.
// The moved category and its new parent are row-locked (in a stable order to
// avoid deadlocks) so concurrent reparents serialize and the cycle re-check
// sees committed state, closing the read-then-write race.
func (r *categoryRepository) UpdateHierarchy(ctx context.Context, id uuid.UUID, fields map[string]any, newParentID *uuid.UUID, newLevel int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockIDs := []uuid.UUID{id}
		if newParentID != nil {
			lockIDs = append(lockIDs, *newParentID)
		}
		sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })
		var locked []entity.Category
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", lockIDs).Find(&locked).Error; err != nil {
			return err
		}

		// Re-check for cycles under the lock: the new parent must not be a
		// descendant of the category being moved.
		if newParentID != nil {
			cur := *newParentID
			for hop := 0; hop < maxAncestorHops; hop++ {
				if cur == id {
					return ErrCategoryCycle
				}
				var p entity.Category
				if err := tx.First(&p, "id = ?", cur).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						break
					}
					return err
				}
				if p.ParentID == nil {
					break
				}
				cur = *p.ParentID
			}
		}

		if err := tx.Model(&entity.Category{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return err
		}
		return updateSubtreeLevelsTx(tx, id, newLevel)
	})
}

// updateSubtreeLevelsTx recomputes the level of every descendant of rootID
// (root itself is updated by the caller). A visited set aborts the walk if a
// node ever repeats — a cycle can never loop forever inside the transaction.
func updateSubtreeLevelsTx(tx *gorm.DB, rootID uuid.UUID, rootLevel int) error {
	seen := map[uuid.UUID]bool{rootID: true}
	level := rootLevel + 1
	frontier := []uuid.UUID{rootID}
	for len(frontier) > 0 {
		var ids []uuid.UUID
		if err := tx.Model(&entity.Category{}).
			Where("parent_id IN ?", frontier).
			Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, cid := range ids {
			if seen[cid] {
				return ErrCategoryCycle
			}
			seen[cid] = true
		}
		if err := tx.Model(&entity.Category{}).
			Where("id IN ?", ids).
			Update("level", level).Error; err != nil {
			return err
		}
		frontier = ids
		level++
	}
	return nil
}
