package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/slug"
	"github.com/mrbayss/Go-Commerce/internal/pkg/storage"
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
	UpdateLogo(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*model.StoreRes, error)
	DeleteLogo(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error)
	UpdateBanner(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*model.StoreRes, error)
	DeleteBanner(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error)
}

type storeService struct {
	StoreRepo      repository.StoreRepository
	UserRepo       repository.UserRepository
	Log            *logrus.Logger
	Validation     *validator.Validate
	CommissionRate float64
	UploadDir      string
	MaxLogoSize    int64
	MaxBannerSize  int64
}

func NewStoreService(
	storeRepo repository.StoreRepository,
	userRepo repository.UserRepository,
	log *logrus.Logger,
	validation *validator.Validate,
	commissionRate float64,
	uploadDir string,
	maxLogoSize int64,
	maxBannerSize int64,
) StoreService {
	return &storeService{
		StoreRepo:      storeRepo,
		UserRepo:       userRepo,
		Log:            log,
		Validation:     validation,
		CommissionRate: commissionRate,
		UploadDir:      uploadDir,
		MaxLogoSize:    maxLogoSize,
		MaxBannerSize:  maxBannerSize,
	}
}

func (ss *storeService) Create(ctx context.Context, userID uuid.UUID, request model.CreateStoreReq) (*model.StoreRes, error) {
	log := serviceLogger(ctx, ss.Log, "Create").WithField("user_id", userID.String())

	if err := validateStruct(log, ss.Validation, request); err != nil {
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
	log := serviceLogger(ctx, ss.Log, "GetMine").WithField("user_id", userID.String())

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
	log := serviceLogger(ctx, ss.Log, "UpdateMine").WithField("user_id", userID.String())

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

	if err := validateStruct(log, ss.Validation, request); err != nil {
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
	log := serviceLogger(ctx, ss.Log, "List")

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
	log := serviceLogger(ctx, ss.Log, "Approve").WithField("store_id", storeID.String())

	store, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to approve store")
	}

	// 1. Idempotent check: if the store is already active, return 200 no-op.
	if store.Status == entity.StoreStatusActive {
		return storeResponse(store), nil
	}

	// Role promotion is decided atomically inside the repository tx
	// (only a buyer is promoted; an admin is never demoted).
	if err := ss.StoreRepo.Approve(ctx, storeID, store.UserID); err != nil {
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
	log := serviceLogger(ctx, ss.Log, "Reject").WithField("store_id", storeID.String())

	store, err := ss.StoreRepo.GetByID(ctx, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to reject store")
	}

	// Target status and role demotion are decided atomically inside the
	// repository tx (only a seller is demoted; an admin stays an admin).
	if err := ss.StoreRepo.Reject(ctx, storeID, store.UserID); err != nil {
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

// UpdateLogo stores a new store logo. The pending-store check is skipped on
// purpose: a seller may brand the store before admin approval.
func (ss *storeService) UpdateLogo(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*model.StoreRes, error) {
	return ss.updateStoreImage(ctx, userID, file, "logo_url", ss.MaxLogoSize, "UpdateLogo", "logo")
}

func (ss *storeService) DeleteLogo(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error) {
	return ss.deleteStoreImage(ctx, userID, "logo_url", "DeleteLogo", "logo")
}

func (ss *storeService) UpdateBanner(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*model.StoreRes, error) {
	return ss.updateStoreImage(ctx, userID, file, "banner_url", ss.MaxBannerSize, "UpdateBanner", "banner")
}

func (ss *storeService) DeleteBanner(ctx context.Context, userID uuid.UUID) (*model.StoreRes, error) {
	return ss.deleteStoreImage(ctx, userID, "banner_url", "DeleteBanner", "banner")
}

// updateStoreImage saves the file first, updates the DB column, then removes
// the previous file (best-effort). Mirrors userService.UpdateAvatar.
func (ss *storeService) updateStoreImage(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, column string, maxSize int64, fn, label string) (*model.StoreRes, error) {
	log := serviceLogger(ctx, ss.Log, fn).WithField("user_id", userID.String())

	if file.Size > maxSize {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity,
			fmt.Sprintf("%s exceeds maximum size of %d MB", label, maxSize/(1024*1024)))
	}

	store, err := ss.StoreRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload "+label)
	}

	publicPath, err := storage.SaveUpload(file, ss.UploadDir, "stores", storage.MIMEJPEG, storage.MIMEPNG, storage.MIMEWebP)
	if err != nil {
		if errors.Is(err, storage.ErrUnsupportedType) {
			return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "unsupported format (allowed: JPEG, PNG, WebP)")
		}
		log.WithError(err).Error("failed to save " + label + " file")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload "+label)
	}

	if err := ss.StoreRepo.UpdateFields(ctx, store.ID, map[string]any{column: publicPath}); err != nil {
		if rmErr := storage.Remove(storage.DiskPath(ss.UploadDir, publicPath)); rmErr != nil {
			log.WithError(rmErr).Warn("failed to remove orphaned " + label + " file")
		}
		log.WithError(err).Error("failed to update store " + label)
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload "+label)
	}

	if old := storeField(store, column); old != nil {
		if err := storage.Remove(storage.DiskPath(ss.UploadDir, *old)); err != nil {
			log.WithError(err).Warn("failed to remove previous " + label + " file")
		}
	}

	updated, err := ss.StoreRepo.GetByID(ctx, store.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve updated store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload "+label)
	}

	log.Info("store " + label + " updated")
	return storeResponse(updated), nil
}

func (ss *storeService) deleteStoreImage(ctx context.Context, userID uuid.UUID, column, fn, label string) (*model.StoreRes, error) {
	log := serviceLogger(ctx, ss.Log, fn).WithField("user_id", userID.String())

	store, err := ss.StoreRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "store not found")
		}
		log.WithError(err).Error("failed to retrieve store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete "+label)
	}

	old := storeField(store, column)
	if old == nil {
		return nil, apperrors.Wrap(nil, http.StatusNotFound, label+" not found")
	}

	if err := ss.StoreRepo.UpdateFields(ctx, store.ID, map[string]any{column: nil}); err != nil {
		log.WithError(err).Error("failed to clear store " + label)
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete "+label)
	}

	if err := storage.Remove(storage.DiskPath(ss.UploadDir, *old)); err != nil {
		log.WithError(err).Warn("failed to remove " + label + " file")
	}

	updated, err := ss.StoreRepo.GetByID(ctx, store.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve updated store")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete "+label)
	}

	log.Info("store " + label + " deleted")
	return storeResponse(updated), nil
}

// storeField reads either the logo or banner URL column.
func storeField(store *entity.Store, column string) *string {
	if column == "banner_url" {
		return store.BannerURL
	}
	return store.LogoURL
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
