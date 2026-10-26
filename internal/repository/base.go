package repository

import (
	"context"

	"gorm.io/gorm"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	Updates(ctx context.Context, entity *T, id any) error
	Delete(ctx context.Context, id any) error
	GetByID(ctx context.Context, id any) (*T, error)
	GetByColumn(ctx context.Context, column string, value any) (*T, error)
	Exists(ctx context.Context, column string, value any) (bool, error)
	GetAll(ctx context.Context) ([]T, error)
}

type RepositoryImpl[T any] struct {
	db *gorm.DB
}

// This function is usually not to be called directly, this is just for type checking of implementation of Repository interface
func repositoryImpl[T any](db *gorm.DB) Repository[T] {
	return &RepositoryImpl[T]{db: db}
}

func (r *RepositoryImpl[T]) Create(ctx context.Context, entity *T) (*T, error) {
	return entity, r.db.WithContext(ctx).Create(entity).Error
}

// Fill id with either string or uuid.UUID
func (r *RepositoryImpl[T]) Updates(ctx context.Context, entity *T, id any) error {
	return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(entity).Error
}

// Fill id with either string or uuid.UUID
func (r *RepositoryImpl[T]) Delete(ctx context.Context, id any) error {
	return r.db.WithContext(ctx).Delete(new(T), "id = ?", id).Error
}

// Fill id with either string or uuid.UUID
func (r *RepositoryImpl[T]) GetByID(ctx context.Context, id any) (*T, error) {
	var t T
	return &t, r.db.WithContext(ctx).First(&t, "id = ?", id).Error
}

func (r *RepositoryImpl[T]) GetByColumn(ctx context.Context, column string, value any) (*T, error) {
	var t T
	return &t, r.db.WithContext(ctx).Where(map[string]any{column: value}).First(&t).Error
}

// Exists checks whether any row matches the given column/value without fetching it.
func (r *RepositoryImpl[T]) Exists(ctx context.Context, column string, value any) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where(map[string]any{column: value}).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RepositoryImpl[T]) GetAll(ctx context.Context) ([]T, error) {
	var t []T
	if err := r.db.WithContext(ctx).Find(&t).Error; err != nil {
		return nil, err
	}

	if t == nil {
		t = []T{}
	}
	return t, nil
}

func (r *RepositoryImpl[T]) UpdateIsEnable(ctx context.Context, isEnable bool, id any) error {
	return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Update("is_enable", isEnable).Error
}
