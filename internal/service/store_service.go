package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/slug"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StoreService interface {
	Create(ctx context.Context, userID uuid.UUID, request model.CreateStoreReq) (*model.StoreRes, error)
	GetMine(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error)
	UpdateMine(ctx context.Context, userID uuid.UUID, request model.UpdateStoreReq) (*model.StoreRes, error)
	List(ctx context.Context, status string, page, limit int) ([]model.StoreRes, int64, error)
	Approve(ctx context.Context, storeID uuid.UUID) (*model.StoreRes, error)
	Reject(ctx context.Context, storeID uuid.UUID) (*model.StoreRes, error)
}

type storeService struct {
	StoreRepo      repository.StoreRepository
	UserRepo       repository.UserRepository
	Log            *logrus.Logger
	Validation     *validator.Validate
	CommissionRate float64
}

func NewStoreService(
	storeRepo repository.StoreRepository,
	userRepo repository.UserRepository,
	log *logrus.Logger,
	validation *validator.Validate,
	commissionRate float64,
) StoreService {
	return &storeService{
		StoreRepo:      storeRepo,
		UserRepo:       userRepo,
		Log:            log,
		Validation:     validation,
		CommissionRate: commissionRate,
	}
}

func (ss *storeService) Create(ctx context.Context, userID uuid.UUID, request model.CreateStoreReq) (*model.StoreRes, error) {
	log := ss.logger(ctx, "Create").WithField("user_id", userID.String())

	if err := ss.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	_, err := ss.StoreRepo.GetByUserID(ctx, userID)
	if err == nil {
		return nil, apperrors.Wrap(nil, http.StatusConflict, "you already have a store")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create store")
	}

	s, err := slug.Unique(ctx, request.Name, func(ctx context.Context, candidate string) (bool, error) {
		return ss.StoreRepo.Exists(ctx, "slug", candidate)
	})
	if err != nil {
		if errors.Is(err, slug.ErrConflict) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "store slug already taken")
		}
		log.WithError(err).Error("failed to check slug availability")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create store")
	}

	store := &entity.Store{
		UserID:         userID,
		Name:           strings.TrimSpace(request.Name),
		Slug:           s,
		Description:    request.Description,
		Status:         entity.StoreStatusPending,
		CommissionRate: ss.CommissionRate,
	}

	created, err := ss.StoreRepo.Create(ctx, store)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "you already have a store")
		}
		log.WithError(err).Error("failed to create store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create store")
	}

	log.Info("store registered")
	return storeResponse(created), nil
}

func (ss *storeService) GetMine(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error) {
	log := ss.logger(ctx, "GetMine").WithField("user_id", userID.String())

	store, err := ss.StoreRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve store")
	}
	return storeResponse(store), nil
}

func (ss *storeService) UpdateMine(ctx context.Context, userID uuid.UUID, request model.UpdateStoreReq) (*model.StoreRes, error) {
	log := ss.logger(ctx, "UpdateMine").WithField("user_id", userID.String())

	fields := make(map[string]any, 2)
	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		request.Name = &name
		fields["name"] = name
	}
	if request.Description != nil {
		fields["description"] = *request.Description
	}

	if len(fields) == 0 {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "no fields to update")
	}

	if err := ss.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	store, err := ss.StoreRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update store")
	}

	// Note: slug is never regenerated on rename — URLs stay stable.
	if err := ss.StoreRepo.UpdateFields(ctx, store.ID, fields); err != nil {
		log.WithError(err).Error("failed to update store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update store")
	}

	updated, err := ss.StoreRepo.GetByID(ctx, store.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve updated store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update store")
	}

	log.Info("store updated")
	return storeResponse(updated), nil
}

func (ss *storeService) List(ctx context.Context, status string, page, limit int) ([]model.StoreRes, int64, error) {
	log := ss.logger(ctx, "List")

	stores, total, err := ss.StoreRepo.ListByStatus(ctx, status, page, limit)
	if err != nil {
		log.WithError(err).Error("failed to list stores")
		return nil, 0, apperrors.Wrap(err, http.StatusInternalServerError, "failed to list stores")
	}

	res := make([]model.StoreRes, 0, len(stores))
	for i := range stores {
		res = append(res, *storeResponse(&stores[i]))
	}
	return res, total, nil
}

func (ss *storeService) Approve(ctx context.Context, storeID uuid.UUID) (*model.StoreRes, error) {
	log := ss.logger(ctx, "Approve").WithField("store_id", storeID.String())

	store, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to approve store")
	}

	user, err := ss.UserRepo.GetByID(ctx, store.UserID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve store owner")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to approve store")
	}

	// Never demote an admin: only a plain buyer is promoted to seller.
	promote := user.Role == entity.UserRoleBuyer

	if err := ss.StoreRepo.Approve(ctx, storeID, store.UserID, promote); err != nil {
		log.WithError(err).Error("failed to approve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to approve store")
	}

	updated, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve approved store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to approve store")
	}

	log.Info("store approved")
	return storeResponse(updated), nil
}

func (ss *storeService) Reject(ctx context.Context, storeID uuid.UUID) (*model.StoreRes, error) {
	log := ss.logger(ctx, "Reject").WithField("store_id", storeID.String())

	store, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to reject store")
	}

	user, err := ss.UserRepo.GetByID(ctx, store.UserID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve store owner")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to reject store")
	}

	// Never-approved stores become rejected; once-active stores become suspended.
	target := entity.StoreStatusRejected
	if store.Status == entity.StoreStatusActive || store.Status == entity.StoreStatusSuspended {
		target = entity.StoreStatusSuspended
	}
	demote := user.Role == entity.UserRoleSeller

	if err := ss.StoreRepo.Reject(ctx, storeID, store.UserID, target, demote); err != nil {
		log.WithError(err).Error("failed to reject store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to reject store")
	}

	updated, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve rejected store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to reject store")
	}

	log.Info("store rejected")
	return storeResponse(updated), nil
}

// requireActiveStore loads the caller's store and enforces the
// active-store guard used by all seller endpoints. Checking the DB (not
// JWT claims) means a freshly approved seller can sell immediately.
func requireActiveStore(ctx context.Context, repo repository.StoreRepository, userID uuid.UUID) (*entity.Store, error) {
	store, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve store")
	}
	if store.Status != entity.StoreStatusActive {
		return nil, apperrors.Wrap(nil, http.StatusForbidden, "store is not active")
	}
	return store, nil
}

func storeResponse(store *entity.Store) *model.StoreRes {
	return &model.StoreRes{
		ID:             store.ID,
		UserID:         store.UserID,
		Name:           store.Name,
		Slug:           store.Slug,
		Description:    store.Description,
		LogoURL:        store.LogoURL,
		BannerURL:      store.BannerURL,
		Status:         string(store.Status),
		CommissionRate: store.CommissionRate,
		CreatedAt:      store.CreatedAt,
		UpdatedAt:      store.UpdatedAt,
	}
}

func (ss *storeService) logger(ctx context.Context, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return ss.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}
