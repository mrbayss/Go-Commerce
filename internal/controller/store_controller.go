package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/middleware"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type StoreController struct {
	StoreService service.StoreService
}

func NewStoreController(storeService service.StoreService) *StoreController {
	return &StoreController{
		StoreService: storeService,
	}
}

func (sc *StoreController) Create(c fiber.Ctx) error {
	var request model.CreateStoreReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := sc.StoreService.Create(c.Context(), userID, request)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, "store created successfully", res)
}

func (sc *StoreController) GetMine(c fiber.Ctx) error {
	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := sc.StoreService.GetMine(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "store retrieved successfully", res)
}

func (sc *StoreController) UpdateMine(c fiber.Ctx) error {
	var request model.UpdateStoreReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := sc.StoreService.UpdateMine(c.Context(), userID, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "store updated successfully", res)
}

func (sc *StoreController) List(c fiber.Ctx) error {
	if _, _, err := middleware.AuthIdentity(c); err != nil {
		return err
	}

	status := c.Query("status")
	page, limit := pagination(c)

	items, total, err := sc.StoreService.List(c.Context(), status, page, limit)
	if err != nil {
		return err
	}

	return response.PaginatedResponse(c, "stores retrieved successfully", items, page, limit, int(total))
}

func (sc *StoreController) Approve(c fiber.Ctx) error {
	storeID, err := parseStoreID(c)
	if err != nil {
		return err
	}

	res, err := sc.StoreService.Approve(c.Context(), storeID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "store approved successfully", res)
}

func (sc *StoreController) Reject(c fiber.Ctx) error {
	storeID, err := parseStoreID(c)
	if err != nil {
		return err
	}

	res, err := sc.StoreService.Reject(c.Context(), storeID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "store rejected successfully", res)
}

func parseStoreID(c fiber.Ctx) (uuid.UUID, error) {
	storeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid store id")
	}
	return storeID, nil
}

// pagination reads page/limit with defaults and clamps limit to 100.
func pagination(c fiber.Ctx) (page, limit int) {
	page = fiber.Query[int](c, "page", 1)
	if page < 1 {
		page = 1
	}
	limit = fiber.Query[int](c, "limit", 20)
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
