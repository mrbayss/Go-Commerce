package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/cache"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfileRes, error)
}

type userService struct {
	UserRepo repository.UserRepository
	Log      *logrus.Logger
	Redis    *redis.Client
}

func NewUserService(userRepo repository.UserRepository, log *logrus.Logger, redis *redis.Client) UserService {
	return &userService{
		UserRepo: userRepo,
		Log:      log,
		Redis:    redis,
	}
}

func (us *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfileRes, error) {
	log := us.logger(ctx, "GetProfile").WithField("user_id", userID.String())
	cacheKey := "user:profile:" + userID.String()

	cachedUser, err := cache.Get[model.UserProfileRes](ctx, us.Redis, cacheKey)
	if err == nil {
		return cachedUser, nil
	}
	if !errors.Is(err, cache.ErrNotFound) {
		log.WithError(err).Warn("failed to retrieve user profile from cache")
	}

	user, err := us.UserRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("user not found")
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "user not found")
		}
		log.WithError(err).Error("failed to retrieve user profile")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve profile")
	}

	if !user.IsActive {
		log.Warn("user account is inactive")
		return nil, apperrors.Wrap(nil, http.StatusForbidden, "account is inactive")
	}

	res := &model.UserProfileRes{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if err := cache.Set(ctx, us.Redis, cacheKey, res, 15*time.Minute); err != nil {
		log.WithError(err).Warn("failed to set user profile to cache")
	}

	return res, nil
}

func (us *userService) logger(ctx context.Context, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return us.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}
