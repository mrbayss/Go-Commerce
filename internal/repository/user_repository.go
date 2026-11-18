package repository

import (
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	Repository[entity.User]
}

type userRepository struct {
	RepositoryImpl[entity.User]
	db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		RepositoryImpl: RepositoryImpl[entity.User]{db: db},
		db:             db,
	}
}
