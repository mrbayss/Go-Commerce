package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/middleware"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type ProductController struct {
	ProductService service.ProductService
}

func NewProductController(productService service.ProductService) *ProductController {
	return &ProductController{
		ProductService: productService,
	}
}

func (pc *ProductController) Create(c fiber.Ctx) error {
	var request model.CreateProductReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := pc.ProductService.Create(c.Context(), userID, request)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, "product created successfully", res)
}

func (pc *ProductController) ListMine(c fiber.Ctx) error {
	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	page, limit := pagination(c)
	items, total, err := pc.ProductService.ListMine(c.Context(), userID, page, limit)
	if err != nil {
		return err
	}

	return response.PaginatedResponse(c, "products retrieved successfully", items, page, limit, int(total))
}

func (pc *ProductController) Get(c fiber.Ctx) error {
	userID, claims, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	isAdmin := claims != nil && claims.Role == string(entity.UserRoleAdmin)

	res, err := pc.ProductService.Get(c.Context(), userID, isAdmin, productID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "product retrieved successfully", res)
}

func (pc *ProductController) Update(c fiber.Ctx) error {
	var request model.UpdateProductReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	res, err := pc.ProductService.Update(c.Context(), userID, productID, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "product updated successfully", res)
}

func (pc *ProductController) Delete(c fiber.Ctx) error {
	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	if err := pc.ProductService.Delete(c.Context(), userID, productID); err != nil {
		return err
	}

	return response.SuccessResponse(c, "product deleted successfully", nil)
}

func (pc *ProductController) AddVariant(c fiber.Ctx) error {
	var request model.VariantReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	res, err := pc.ProductService.AddVariant(c.Context(), userID, productID, request)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, "variant added successfully", res)
}

func (pc *ProductController) UpdateVariant(c fiber.Ctx) error {
	var request model.UpdateVariantReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	variantID, err := parseVariantID(c)
	if err != nil {
		return err
	}

	res, err := pc.ProductService.UpdateVariant(c.Context(), userID, productID, variantID, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "variant updated successfully", res)
}

func (pc *ProductController) DeleteVariant(c fiber.Ctx) error {
	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	productID, err := parseProductID(c)
	if err != nil {
		return err
	}

	variantID, err := parseVariantID(c)
	if err != nil {
		return err
	}

	res, err := pc.ProductService.DeleteVariant(c.Context(), userID, productID, variantID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "variant deleted successfully", res)
}

func parseProductID(c fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid product id")
	}
	return id, nil
}

func parseVariantID(c fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("variantId"))
	if err != nil {
		return uuid.Nil, apperrors.Wrap(nil, http.StatusBadRequest, "invalid variant id")
	}
	return id, nil
}
