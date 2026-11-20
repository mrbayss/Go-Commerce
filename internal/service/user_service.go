package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/cache"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const profileCacheTTL = 15 * time.Minute

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfileRes, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, request model.UpdateProfileReq) (*model.UserProfileRes, error)
}

type userService struct {
	UserRepo   repository.UserRepository
	Log        *logrus.Logger
	Redis      *redis.Client
	Validation *validator.Validate
}

func NewUserService(
	userRepo repository.UserRepository,
	log *logrus.Logger,
	redis *redis.Client,
	validation *validator.Validate,
) UserService {
	return &userService{
		UserRepo:   userRepo,
		Log:        log,
		Redis:      redis,
		Validation: validation,
	}
}

func (us *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserProfileRes, error) {
	log := us.logger(ctx, "GetProfile").WithField("user_id", userID.String())
	cacheKey := profileCacheKey(userID)

	cachedUser, err := cache.Get[model.UserProfileRes](ctx, us.Redis, cacheKey)
	if err == nil {
		return cachedUser, nil
	}
	if !errors.Is(err, cache.ErrNotFound) {
		log.WithError(err).Warn("failed to retrieve user profile from cache")
	}

	user, err := us.getActiveUser(ctx, log, userID)
	if err != nil {
		return nil, err
	}

	res := profileResponse(user)

	if err := cache.Set(ctx, us.Redis, cacheKey, res, profileCacheTTL); err != nil {
		log.WithError(err).Warn("failed to set user profile to cache")
	}

	return res, nil
}

func (us *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, request model.UpdateProfileReq) (*model.UserProfileRes, error) {
	log := us.logger(ctx, "UpdateProfile").WithField("user_id", userID.String())

	fields := make(map[string]any, 3)

	if request.FirstName != nil {
		name := strings.TrimSpace(*request.FirstName)
		request.FirstName = &name
		fields["first_name"] = name
	}
	if request.LastName != nil {
		name := strings.TrimSpace(*request.LastName)
		request.LastName = &name
		fields["last_name"] = name
	}
	if request.Phone != nil {
		if *request.Phone == "" {
			request.Phone = nil
			fields["phone"] = nil
		} else {
			phone := strings.TrimSpace(*request.Phone)
			if phone != "" {
				request.Phone = &phone
			}
			fields["phone"] = phone
		}
	}

	if len(fields) == 0 {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "no fields to update")
	}

	if err := us.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	if _, err := us.getActiveUser(ctx, log, userID); err != nil {
		return nil, err
	}

	if err := us.UserRepo.UpdateFields(ctx, userID, fields); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Warn("phone number already registered")
			return nil, apperrors.Wrap(nil, http.StatusConflict, "phone number already registered")
		}
		log.WithError(err).Error("failed to update user profile")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update profile")
	}

	user, err := us.getActiveUser(ctx, log, userID)
	if err != nil {
		return nil, err
	}

	us.deleteProfileCache(ctx, log, userID)
	log.Info("user profile updated")
	return profileResponse(user), nil
}

// getActiveUser fetches the user and enforces the active-account check.
func (us *userService) getActiveUser(ctx context.Context, log *logrus.Entry, userID uuid.UUID) (*entity.User, error) {
	user, err := us.UserRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("user not found")
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "user not found")
		}
		log.WithError(err).Error("failed to retrieve user")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve user")
	}
	if !user.IsActive {
		log.Warn("user account is inactive")
		return nil, apperrors.Wrap(nil, http.StatusForbidden, "account is inactive")
	}
	return user, nil
}

func (us *userService) deleteProfileCache(ctx context.Context, log *logrus.Entry, userID uuid.UUID) {
	if err := cache.Del(ctx, us.Redis, profileCacheKey(userID)); err != nil {
		log.WithError(err).Warn("failed to invalidate user profile cache")
	}
}

func profileCacheKey(userID uuid.UUID) string {
	return "user:profile:" + userID.String()
}

func profileResponse(user *entity.User) *model.UserProfileRes {
	return &model.UserProfileRes{
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
}

func (us *userService) logger(ctx context.Context, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return us.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}
