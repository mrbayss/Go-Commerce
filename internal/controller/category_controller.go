package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type CategoryController struct {
	CategoryService service.CategoryService
}

func NewCategoryController(categoryService service.CategoryService) *CategoryController {
	return &CategoryController{
		CategoryService: categoryService,
	}
}

// List returns active categories (Bearer — sellers pick one when creating products).
func (cc *CategoryController) List(c fiber.Ctx) error {
	page, limit := pagination(c)
	items, total, err := cc.CategoryService.List(c.Context(), true, page, limit)
	if err != nil {
		return err
	}

	return response.PaginatedResponse(c, "categories retrieved successfully", items, page, limit, int(total))
}

// ListAll returns every category including inactive ones (admin).
func (cc *CategoryController) ListAll(c fiber.Ctx) error {
	page, limit := pagination(c)
	items, total, err := cc.CategoryService.List(c.Context(), false, page, limit)
	if err != nil {
		return err
	}

	return response.PaginatedResponse(c, "categories retrieved successfully", items, page, limit, int(total))
}

func (cc *CategoryController) Get(c fiber.Ctx) error {
	id, err := parseCategoryID(c)
	if err != nil {
		return err
	}

	res, err := cc.CategoryService.Get(c.Context(), id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "category retrieved successfully", res)
}

func (cc *CategoryController) Create(c fiber.Ctx) error {
	var request model.CreateCategoryReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	res, err := cc.CategoryService.Create(c.Context(), request)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, "category created successfully", res)
}

func (cc *CategoryController) Update(c fiber.Ctx) error {
	var request model.UpdateCategoryReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	id, err := parseCategoryID(c)
	if err != nil {
		return err
	}

	res, err := cc.CategoryService.Update(c.Context(), id, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "category updated successfully", res)
}

func (cc *CategoryController) Delete(c fiber.Ctx) error {
	id, err := parseCategoryID(c)
	if err != nil {
		return err
	}

	if err := cc.CategoryService.Delete(c.Context(), id); err != nil {
		return err
	}

	return response.SuccessResponse(c, "category deleted successfully", nil)
}

func parseCategoryID(c fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid category id")
	}
	return id, nil
}
