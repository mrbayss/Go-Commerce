package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
)

type StoreRepository interface {
	Repository[entity.Store]
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Store, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Store, error)
	ListByStatus(ctx context.Context, status string, page, limit int) ([]entity.Store, int64, error)
	Approve(ctx context.Context, storeID, userID uuid.UUID, promoteToSeller bool) error
	Reject(ctx context.Context, storeID, userID uuid.UUID, targetStatus entity.StoreStatus, demoteToBuyer bool) error
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

// Approve sets the store active and, when promoteToSeller is true, raises
// the owner's role to seller — atomically.
func (r *storeRepository) Approve(ctx context.Context, storeID, userID uuid.UUID, promoteToSeller bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Store{}).Where("id = ?", storeID).Update("status", entity.StoreStatusActive).Error; err != nil {
			return err
		}
		if promoteToSeller {
			if err := tx.Model(&entity.User{}).Where("id = ?", userID).Update("role", entity.UserRoleSeller).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Reject sets the store to targetStatus (rejected or suspended) and, when
// demoteToBuyer is true, lowers the owner's role to buyer — atomically.
func (r *storeRepository) Reject(ctx context.Context, storeID, userID uuid.UUID, targetStatus entity.StoreStatus, demoteToBuyer bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Store{}).Where("id = ?", storeID).Update("status", targetStatus).Error; err != nil {
			return err
		}
		if demoteToBuyer {
			if err := tx.Model(&entity.User{}).Where("id = ?", userID).Update("role", entity.UserRoleBuyer).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
