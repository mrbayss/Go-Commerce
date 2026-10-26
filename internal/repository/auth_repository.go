package repository

import (
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
)

// AuthRepository defines auth-specific data access operations.
type AuthRepository interface {
	Repository[entity.User]
}

// GormAuthRepository implements AuthRepository using GORM.
type authRepository struct {
	RepositoryImpl[entity.User]
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository.
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		RepositoryImpl: RepositoryImpl[entity.User]{db: db},
		db:             db,
	}
}
