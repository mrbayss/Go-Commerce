package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueryOption modifies a GORM query chain.
type QueryOption func(*gorm.DB) *gorm.DB

// WithFilter applies WHERE condition(s) to the query.
func WithFilter(query string, args ...any) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

// WithOrder applies ORDER BY clause to the query.
func WithOrder(value any) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	}
}

// WithPreload applies Preload eagerly loading associations.
func WithPreload(relation string, args ...any) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(relation, args...)
	}
}

func applyOpts(db *gorm.DB, opts []QueryOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// Repository defines common CRUD operations for an entity type T.
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uuid.UUID, opts ...QueryOption) (*T, error)
	FindAll(ctx context.Context, opts ...QueryOption) ([]T, error)
	FindWithPagination(ctx context.Context, page, limit int, opts ...QueryOption) ([]T, int64, error)
	Updates(ctx context.Context, id uuid.UUID, data map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GormRepository is the concrete implementation of Repository[T] using GORM.
type GormRepository[T any] struct {
	db *gorm.DB
}

// NewGormRepository creates a new generic repository instance.
func NewGormRepository[T any](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: db}
}

// Create inserts a new record into the database.
func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID retrieves a single record by its UUID primary key.
func (r *GormRepository[T]) FindByID(ctx context.Context, id uuid.UUID, opts ...QueryOption) (*T, error) {
	var entity T
	q := applyOpts(r.db.WithContext(ctx), opts)

	if err := q.First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records matching optional query modifiers.
func (r *GormRepository[T]) FindAll(ctx context.Context, opts ...QueryOption) ([]T, error) {
	var items []T
	q := applyOpts(r.db.WithContext(ctx).Model(new(T)), opts)

	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindWithPagination retrieves records with pagination (page & limit) and returns items along with total count.
func (r *GormRepository[T]) FindWithPagination(ctx context.Context, page, limit int, opts ...QueryOption) ([]T, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var items []T
	var total int64

	// Count phase: apply filters but omit pagination
	countQ := applyOpts(r.db.WithContext(ctx).Model(new(T)), opts)
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data phase: apply all query options + offset and limit
	dataQ := applyOpts(r.db.WithContext(ctx).Model(new(T)), opts)
	if err := dataQ.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Updates updates specific fields of an entity by ID using a map to prevent zero-value omitting.
func (r *GormRepository[T]) Updates(ctx context.Context, id uuid.UUID, data map[string]any) error {
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete performs a soft delete on the record by ID (if BaseModel with DeletedAt is embedded).
func (r *GormRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
