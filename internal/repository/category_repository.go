package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Repository[entity.Category]
	GetActiveByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	List(ctx context.Context, activeOnly bool, page, limit int) ([]entity.Category, int64, error)
	CountChildren(ctx context.Context, id uuid.UUID) (int64, error)
	CountActiveProducts(ctx context.Context, id uuid.UUID) (int64, error)
	UpdateSubtreeLevels(ctx context.Context, rootID uuid.UUID, rootLevel int) error
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

// UpdateSubtreeLevels recomputes the level of every descendant of rootID
// (root itself is updated by the caller) inside a single transaction.
func (r *categoryRepository) UpdateSubtreeLevels(ctx context.Context, rootID uuid.UUID, rootLevel int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			if err := tx.Model(&entity.Category{}).
				Where("id IN ?", ids).
				Update("level", level).Error; err != nil {
				return err
			}
			frontier = ids
			level++
		}
		return nil
	})
}
