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

// maxAncestorHops guards the cycle check against corrupt parent chains.
const maxAncestorHops = 100

type CategoryService interface {
	Create(ctx context.Context, request model.CreateCategoryReq) (*model.CategoryRes, error)
	Get(ctx context.Context, id uuid.UUID) (*model.CategoryRes, error)
	List(ctx context.Context, activeOnly bool, page, limit int) ([]model.CategoryRes, int64, error)
	Update(ctx context.Context, id uuid.UUID, request model.UpdateCategoryReq) (*model.CategoryRes, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	CatRepo    repository.CategoryRepository
	Log        *logrus.Logger
	Validation *validator.Validate
}

func NewCategoryService(
	catRepo repository.CategoryRepository,
	log *logrus.Logger,
	validation *validator.Validate,
) CategoryService {
	return &categoryService{
		CatRepo:    catRepo,
		Log:        log,
		Validation: validation,
	}
}

func (cs *categoryService) Create(ctx context.Context, request model.CreateCategoryReq) (*model.CategoryRes, error) {
	log := cs.logger(ctx, "Create")

	if err := cs.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	cat := &entity.Category{
		Name:        strings.TrimSpace(request.Name),
		Description: request.Description,
		ImageURL:    request.ImageURL,
	}

	if request.ParentID != nil {
		parent, err := cs.loadActiveParent(ctx, *request.ParentID)
		if err != nil {
			return nil, err
		}
		cat.ParentID = &parent.ID
		cat.Level = parent.Level + 1
	}

	s, err := slug.Unique(ctx, cat.Name, func(ctx context.Context, candidate string) (bool, error) {
		return cs.CatRepo.Exists(ctx, "slug", candidate)
	})
	if err != nil {
		if errors.Is(err, slug.ErrConflict) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "category slug already taken")
		}
		log.WithError(err).Error("failed to check slug availability")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create category")
	}
	cat.Slug = s

	created, err := cs.CatRepo.Create(ctx, cat)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "category slug already taken")
		}
		log.WithError(err).Error("failed to create category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create category")
	}

	log.Info("category created")
	return categoryResponse(created), nil
}

func (cs *categoryService) Get(ctx context.Context, id uuid.UUID) (*model.CategoryRes, error) {
	log := cs.logger(ctx, "Get")

	cat, err := cs.CatRepo.GetActiveByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "category not found")
		}
		log.WithError(err).Error("failed to retrieve category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve category")
	}
	return categoryResponse(cat), nil
}

func (cs *categoryService) List(ctx context.Context, activeOnly bool, page, limit int) ([]model.CategoryRes, int64, error) {
	log := cs.logger(ctx, "List")

	cats, total, err := cs.CatRepo.List(ctx, activeOnly, page, limit)
	if err != nil {
		log.WithError(err).Error("failed to list categories")
		return nil, 0, apperrors.Wrap(err, http.StatusInternalServerError, "failed to list categories")
	}

	res := make([]model.CategoryRes, 0, len(cats))
	for i := range cats {
		res = append(res, *categoryResponse(&cats[i]))
	}
	return res, total, nil
}

func (cs *categoryService) Update(ctx context.Context, id uuid.UUID, request model.UpdateCategoryReq) (*model.CategoryRes, error) {
	log := cs.logger(ctx, "Update").WithField("category_id", id.String())

	cat, err := cs.CatRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "category not found")
		}
		log.WithError(err).Error("failed to retrieve category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update category")
	}

	fields := make(map[string]any, 5)
	parentChanged := false

	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		request.Name = &name
		fields["name"] = name
	}
	if request.Description != nil {
		fields["description"] = *request.Description
	}
	if request.IsActive != nil {
		fields["is_active"] = *request.IsActive
	}
	if request.SortOrder != nil {
		fields["sort_order"] = *request.SortOrder
	}

	if request.ParentID != nil {
		if *request.ParentID == "" {
			// Re-root: parent becomes NULL, level resets to 0.
			fields["parent_id"] = nil
			fields["level"] = 0
			parentChanged = true
		} else {
			parentID, err := uuid.Parse(*request.ParentID)
			if err != nil {
				return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "invalid parent_id")
			}
			if parentID == id {
				return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "category cannot be its own parent")
			}
			parent, err := cs.loadActiveParent(ctx, parentID)
			if err != nil {
				return nil, err
			}
			// A cycle exists when the current category is an ancestor of the new parent.
			cycle, err := cs.isAncestorOf(ctx, id, parent)
			if err != nil {
				return nil, err
			}
			if cycle {
				return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "category cycle detected")
			}
			fields["parent_id"] = parentID
			fields["level"] = parent.Level + 1
			parentChanged = true
		}
	}

	if len(fields) == 0 {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "no fields to update")
	}

	if err := cs.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	if err := cs.CatRepo.UpdateFields(ctx, cat.ID, fields); err != nil {
		log.WithError(err).Error("failed to update category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update category")
	}

	if parentChanged {
		newLevel := fields["level"].(int)
		if err := cs.CatRepo.UpdateSubtreeLevels(ctx, cat.ID, newLevel); err != nil {
			log.WithError(err).Error("failed to recompute subtree levels")
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update category")
		}
	}

	updated, err := cs.CatRepo.GetByID(ctx, cat.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve updated category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update category")
	}

	log.Info("category updated")
	return categoryResponse(updated), nil
}

func (cs *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	log := cs.logger(ctx, "Delete").WithField("category_id", id.String())

	cat, err := cs.CatRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Wrap(nil, http.StatusNotFound, "category not found")
		}
		log.WithError(err).Error("failed to retrieve category")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete category")
	}

	children, err := cs.CatRepo.CountChildren(ctx, cat.ID)
	if err != nil {
		log.WithError(err).Error("failed to count subcategories")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete category")
	}
	if children > 0 {
		return apperrors.Wrap(nil, http.StatusConflict, "category has subcategories")
	}

	products, err := cs.CatRepo.CountActiveProducts(ctx, cat.ID)
	if err != nil {
		log.WithError(err).Error("failed to count active products")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete category")
	}
	if products > 0 {
		return apperrors.Wrap(nil, http.StatusConflict, "category has active products")
	}

	if err := cs.CatRepo.Delete(ctx, cat.ID); err != nil {
		log.WithError(err).Error("failed to delete category")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete category")
	}

	log.Info("category deleted")
	return nil
}

// loadActiveParent fetches a parent that exists and is active (422 otherwise).
func (cs *categoryService) loadActiveParent(ctx context.Context, parentID uuid.UUID) (*entity.Category, error) {
	parent, err := cs.CatRepo.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "parent category not found")
		}
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve parent category")
	}
	if !parent.IsActive {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "parent category is not active")
	}
	return parent, nil
}

// isAncestorOf walks up from candidate; true when it reaches ancestorID.
func (cs *categoryService) isAncestorOf(ctx context.Context, ancestorID uuid.UUID, candidate *entity.Category) (bool, error) {
	cur := candidate
	for i := 0; i < maxAncestorHops && cur != nil; i++ {
		if cur.ID == ancestorID {
			return true, nil
		}
		if cur.ParentID == nil {
			return false, nil
		}
		next, err := cs.CatRepo.GetByID(ctx, *cur.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, apperrors.Wrap(err, http.StatusInternalServerError, "failed to validate category hierarchy")
		}
		cur = next
	}
	return false, nil
}

func categoryResponse(cat *entity.Category) *model.CategoryRes {
	return &model.CategoryRes{
		ID:          cat.ID,
		ParentID:    cat.ParentID,
		Name:        cat.Name,
		Slug:        cat.Slug,
		Description: cat.Description,
		ImageURL:    cat.ImageURL,
		Level:       cat.Level,
		SortOrder:   cat.SortOrder,
		IsActive:    cat.IsActive,
		CreatedAt:   cat.CreatedAt,
		UpdatedAt:   cat.UpdatedAt,
	}
}

func (cs *categoryService) logger(ctx context.Context, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return cs.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}
