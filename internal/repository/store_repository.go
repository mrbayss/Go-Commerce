package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StoreRepository interface {
	Repository[entity.Store]
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Store, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Store, error)
	ListByStatus(ctx context.Context, status string, page, limit int) ([]entity.Store, int64, error)
	Approve(ctx context.Context, storeID, userID uuid.UUID) error
	Reject(ctx context.Context, storeID, userID uuid.UUID) error
}

type storeRepository struct {
	RepositoryImpl[entity.Store]
	db *gorm.DB
}

// NewStoreRepository creates a new store repository.
func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{
		RepositoryImpl: RepositoryImpl[entity.Store]{db: db},
		db:             db,
	}
}

func (r *storeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Store, error) {
	var store entity.Store
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&store).Error
	return &store, err
}

func (r *storeRepository) GetBySlug(ctx context.Context, slug string) (*entity.Store, error) {
	var store entity.Store
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&store).Error
	return &store, err
}

func (r *storeRepository) ListByStatus(ctx context.Context, status string, page, limit int) ([]entity.Store, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.Store{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	stores := make([]entity.Store, 0)
	if err := tx.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&stores).Error; err != nil {
		return nil, 0, err
	}
	return stores, total, nil
}

// Approve sets the store active and promotes the owner to seller — atomically.
// It is idempotent (already active -> no-op) and supports reactivation of
// rejected or suspended stores. The store row is locked so role promotion and
// status update occur consistently. Role update is conditional in SQL (only
// a buyer is promoted), so a concurrent role change can never demote an admin.
func (r *storeRepository) Approve(ctx context.Context, storeID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store entity.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&store, "id = ?", storeID).Error; err != nil {
			return err
		}

		if store.Status == entity.StoreStatusActive {
			return nil
		}

		if err := tx.Model(&entity.Store{}).Where("id = ?", storeID).Update("status", entity.StoreStatusActive).Error; err != nil {
			return err
		}
		return tx.Model(&entity.User{}).
			Where("id = ? AND role = ?", userID, entity.UserRoleBuyer).
			Update("role", entity.UserRoleSeller).Error
	})
}

// Reject suspends a once-active store (or rejects a never-active one) and
// demotes the owner to buyer — atomically. The store row is locked so the
// target status is decided from a consistent read, and the role update is
// conditional in SQL (only a seller is demoted; admins stay admins).
func (r *storeRepository) Reject(ctx context.Context, storeID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store entity.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&store, "id = ?", storeID).Error; err != nil {
			return err
		}

		// Never-approved stores become rejected; once-active stores become suspended.
		target := entity.StoreStatusRejected
		if store.Status == entity.StoreStatusActive || store.Status == entity.StoreStatusSuspended {
			target = entity.StoreStatusSuspended
		}
		if err := tx.Model(&entity.Store{}).Where("id = ?", storeID).Update("status", target).Error; err != nil {
			return err
		}
		return tx.Model(&entity.User{}).
			Where("id = ? AND role = ?", userID, entity.UserRoleSeller).
			Update("role", entity.UserRoleBuyer).Error
	})
}
