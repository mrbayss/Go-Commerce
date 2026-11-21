package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/pkg/slug"
	"github.com/mrbayss/Go-Commerce/internal/pkg/storage"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// maxJSONObjectBytes caps metadata / attributes payloads.
const maxJSONObjectBytes = 4096

type ProductService interface {
	Create(ctx context.Context, userID uuid.UUID, request model.CreateProductReq) (*model.ProductRes, error)
	Get(ctx context.Context, userID uuid.UUID, isAdmin bool, productID uuid.UUID) (*model.ProductRes, error)
	ListMine(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.ProductRes, int64, error)
	Update(ctx context.Context, userID, productID uuid.UUID, request model.UpdateProductReq) (*model.ProductRes, error)
	Delete(ctx context.Context, userID, productID uuid.UUID) error
	AddVariant(ctx context.Context, userID, productID uuid.UUID, request model.VariantReq) (*model.ProductRes, error)
	UpdateVariant(ctx context.Context, userID, productID, variantID uuid.UUID, request model.UpdateVariantReq) (*model.ProductRes, error)
	DeleteVariant(ctx context.Context, userID, productID, variantID uuid.UUID) (*model.ProductRes, error)
	AddImage(ctx context.Context, userID, productID uuid.UUID, file *multipart.FileHeader) (*model.ProductRes, error)
	UpdateImage(ctx context.Context, userID, productID, imageID uuid.UUID, request model.UpdateImageReq) (*model.ProductRes, error)
	DeleteImage(ctx context.Context, userID, productID, imageID uuid.UUID) (*model.ProductRes, error)
}

type productService struct {
	ProductRepo  repository.ProductRepository
	StoreRepo    repository.StoreRepository
	CatRepo      repository.CategoryRepository
	Log          *logrus.Logger
	Validation   *validator.Validate
	UploadDir    string
	MaxImageSize int64
	MaxImages    int
}

func NewProductService(
	productRepo repository.ProductRepository,
	storeRepo repository.StoreRepository,
	catRepo repository.CategoryRepository,
	log *logrus.Logger,
	validation *validator.Validate,
	uploadDir string,
	maxImageSize int64,
	maxImages int,
) ProductService {
	return &productService{
		ProductRepo:  productRepo,
		StoreRepo:    storeRepo,
		CatRepo:      catRepo,
		Log:          log,
		Validation:   validation,
		UploadDir:    uploadDir,
		MaxImageSize: maxImageSize,
		MaxImages:    maxImages,
	}
}

func (ps *productService) Create(ctx context.Context, userID uuid.UUID, request model.CreateProductReq) (*model.ProductRes, error) {
	log := ps.logger(ctx, "Create").WithField("user_id", userID.String())

	if err := ps.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}
	if err := validateJSONObject(request.Metadata, "metadata"); err != nil {
		return nil, err
	}
	for i := range request.Variants {
		if err := validateJSONObject(request.Variants[i].Attributes, "attributes"); err != nil {
			return nil, err
		}
	}

	store, err := requireActiveStore(ctx, ps.StoreRepo, userID)
	if err != nil {
		return nil, err
	}

	if _, err := ps.CatRepo.GetActiveByID(ctx, request.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "category not found or not active")
		}
		log.WithError(err).Error("failed to retrieve category")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create product")
	}

	s, err := slug.Unique(ctx, request.Name, func(ctx context.Context, candidate string) (bool, error) {
		return ps.ProductRepo.Exists(ctx, "slug", candidate)
	})
	if err != nil {
		if errors.Is(err, slug.ErrConflict) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "product slug already taken")
		}
		log.WithError(err).Error("failed to check slug availability")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create product")
	}

	product := &entity.Product{
		StoreID:     store.ID,
		CategoryID:  request.CategoryID,
		Name:        strings.TrimSpace(request.Name),
		Slug:        s,
		Description: request.Description,
		Status:      entity.ProductStatusActive,
		IsActive:    true,
		Metadata:    jsonText(request.Metadata),
	}

	variants := make([]entity.ProductVariant, 0, len(request.Variants))
	for _, vr := range request.Variants {
		variants = append(variants, entity.ProductVariant{
			Name:           strings.TrimSpace(vr.Name),
			SKU:            strings.TrimSpace(vr.SKU),
			Price:          vr.Price,
			CompareAtPrice: vr.CompareAtPrice,
			Stock:          vr.Stock,
			Weight:         vr.Weight,
			IsActive:       true,
			Attributes:     jsonText(vr.Attributes),
		})
	}

	created, err := ps.ProductRepo.CreateWithVariants(ctx, product, variants)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, duplicateProductMessage(err))
		}
		log.WithError(err).Error("failed to create product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create product")
	}

	if err := ps.ProductRepo.RecomputePriceRange(ctx, created.ID); err != nil {
		log.WithError(err).Error("failed to compute price range")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create product")
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, created.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve created product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to create product")
	}

	log.Info("product created")
	return productResponse(detail), nil
}

func (ps *productService) Get(ctx context.Context, userID uuid.UUID, isAdmin bool, productID uuid.UUID) (*model.ProductRes, error) {
	log := ps.logger(ctx, "Get").WithField("product_id", productID.String())

	product, err := ps.ProductRepo.GetDetail(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "product not found")
		}
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve product")
	}

	if !isAdmin {
		// 404 (not 403) for foreign products so their existence stays private.
		store, err := ps.StoreRepo.GetByUserID(ctx, userID)
		if err != nil || product.StoreID != store.ID {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "product not found")
		}
	}

	return productResponse(product), nil
}

func (ps *productService) ListMine(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.ProductRes, int64, error) {
	log := ps.logger(ctx, "ListMine").WithField("user_id", userID.String())

	store, err := requireActiveStore(ctx, ps.StoreRepo, userID)
	if err != nil {
		return nil, 0, err
	}

	products, total, err := ps.ProductRepo.ListByStore(ctx, store.ID, page, limit)
	if err != nil {
		log.WithError(err).Error("failed to list products")
		return nil, 0, apperrors.Wrap(err, http.StatusInternalServerError, "failed to list products")
	}

	res := make([]model.ProductRes, 0, len(products))
	for i := range products {
		res = append(res, *productResponse(&products[i]))
	}
	return res, total, nil
}

func (ps *productService) Update(ctx context.Context, userID, productID uuid.UUID, request model.UpdateProductReq) (*model.ProductRes, error) {
	log := ps.logger(ctx, "Update").WithField("product_id", productID.String())

	if err := ps.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]any, 5)
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
	if request.CategoryID != nil {
		if _, err := ps.CatRepo.GetActiveByID(ctx, *request.CategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "category not found or not active")
			}
			log.WithError(err).Error("failed to retrieve category")
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update product")
		}
		fields["category_id"] = *request.CategoryID
	}
	if request.Metadata != nil {
		if err := validateJSONObject(request.Metadata, "metadata"); err != nil {
			return nil, err
		}
		fields["metadata"] = jsonText(request.Metadata)
	}

	if len(fields) == 0 {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "no fields to update")
	}

	// Note: slug is never regenerated on rename — URLs stay stable.
	if err := ps.ProductRepo.UpdateFields(ctx, product.ID, fields); err != nil {
		log.WithError(err).Error("failed to update product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update product")
	}

	updated, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve updated product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update product")
	}

	log.Info("product updated")
	return productResponse(updated), nil
}

func (ps *productService) Delete(ctx context.Context, userID, productID uuid.UUID) error {
	log := ps.logger(ctx, "Delete").WithField("product_id", productID.String())

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return err
	}

	// Collect image URLs before soft-deleting so the files can be removed.
	imageURLs := make([]string, 0, len(product.Images))
	for _, img := range product.Images {
		imageURLs = append(imageURLs, img.URL)
	}

	if err := ps.ProductRepo.Delete(ctx, product.ID); err != nil {
		log.WithError(err).Error("failed to delete product")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete product")
	}
	if err := ps.ProductRepo.DeleteChildren(ctx, product.ID); err != nil {
		log.WithError(err).Error("failed to delete product children")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete product")
	}

	for _, url := range imageURLs {
		if err := storage.Remove(storage.DiskPath(ps.UploadDir, url)); err != nil {
			log.WithError(err).Warn("failed to remove product image file")
		}
	}

	log.Info("product deleted")
	return nil
}

func (ps *productService) AddVariant(ctx context.Context, userID, productID uuid.UUID, request model.VariantReq) (*model.ProductRes, error) {
	log := ps.logger(ctx, "AddVariant").WithField("product_id", productID.String())

	if err := ps.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}
	if err := validateJSONObject(request.Attributes, "attributes"); err != nil {
		return nil, err
	}

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	variant := &entity.ProductVariant{
		ProductID:      product.ID,
		StoreID:        product.StoreID,
		SKU:            strings.TrimSpace(request.SKU),
		Name:           strings.TrimSpace(request.Name),
		Price:          request.Price,
		CompareAtPrice: request.CompareAtPrice,
		Stock:          request.Stock,
		Weight:         request.Weight,
		IsActive:       true,
		Attributes:     jsonText(request.Attributes),
	}

	if err := ps.ProductRepo.CreateVariant(ctx, variant); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, duplicateProductMessage(err))
		}
		log.WithError(err).Error("failed to create variant")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to add variant")
	}

	if err := ps.refreshProduct(ctx, product.ID); err != nil {
		return nil, err
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to add variant")
	}

	log.Info("variant added")
	return productResponse(detail), nil
}

func (ps *productService) UpdateVariant(ctx context.Context, userID, productID, variantID uuid.UUID, request model.UpdateVariantReq) (*model.ProductRes, error) {
	log := ps.logger(ctx, "UpdateVariant").WithField("variant_id", variantID.String())

	if err := ps.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}
	if request.Attributes != nil {
		if err := validateJSONObject(request.Attributes, "attributes"); err != nil {
			return nil, err
		}
	}

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	variant, err := ps.ProductRepo.GetVariant(ctx, product.ID, variantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "variant not found")
		}
		log.WithError(err).Error("failed to retrieve variant")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update variant")
	}

	fields := make(map[string]any, 8)
	if request.SKU != nil {
		sku := strings.TrimSpace(*request.SKU)
		request.SKU = &sku
		fields["sku"] = sku
	}
	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		request.Name = &name
		fields["name"] = name
	}
	if request.Price != nil {
		fields["price"] = *request.Price
	}
	if request.CompareAtPrice != nil {
		fields["compare_at_price"] = *request.CompareAtPrice
	}
	if request.Stock != nil {
		fields["stock"] = *request.Stock
	}
	if request.Weight != nil {
		fields["weight"] = *request.Weight
	}
	if request.IsActive != nil {
		fields["is_active"] = *request.IsActive
	}
	if request.Attributes != nil {
		fields["attributes"] = jsonText(request.Attributes)
	}

	if len(fields) == 0 {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "no fields to update")
	}

	if err := ps.ProductRepo.UpdateVariantFields(ctx, variant.ID, fields); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, duplicateProductMessage(err))
		}
		log.WithError(err).Error("failed to update variant")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update variant")
	}

	// Price / is_active may have changed — recompute the product cache.
	if err := ps.refreshProduct(ctx, product.ID); err != nil {
		return nil, err
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update variant")
	}

	log.Info("variant updated")
	return productResponse(detail), nil
}

func (ps *productService) DeleteVariant(ctx context.Context, userID, productID, variantID uuid.UUID) (*model.ProductRes, error) {
	log := ps.logger(ctx, "DeleteVariant").WithField("variant_id", variantID.String())

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	// Guards (last variant / last active variant) run inside the repository
	// under a product row lock so concurrent deletes cannot race past them.
	if err := ps.ProductRepo.DeleteVariantChecked(ctx, product.ID, variantID); err != nil {
		if errors.Is(err, repository.ErrLastVariant) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "cannot delete the last variant")
		}
		if errors.Is(err, repository.ErrLastActiveVariant) {
			return nil, apperrors.Wrap(nil, http.StatusConflict, "cannot delete the last active variant")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "variant not found")
		}
		log.WithError(err).Error("failed to delete variant")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete variant")
	}

	if err := ps.refreshProduct(ctx, product.ID); err != nil {
		return nil, err
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete variant")
	}

	log.Info("variant deleted")
	return productResponse(detail), nil
}

func (ps *productService) AddImage(ctx context.Context, userID, productID uuid.UUID, file *multipart.FileHeader) (*model.ProductRes, error) {
	log := ps.logger(ctx, "AddImage").WithField("product_id", productID.String())

	if file == nil {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "image file is required")
	}
	if file.Size > ps.MaxImageSize {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "image file size exceeds the maximum")
	}

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	// Fast pre-check so a full product fails before any file is written;
	// the authoritative check happens inside CreateImageChecked under lock.
	count, err := ps.ProductRepo.CountImages(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to count images")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload image")
	}
	if count >= int64(ps.MaxImages) {
		return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity,
			fmt.Sprintf("maximum %d images per product", ps.MaxImages))
	}

	url, err := storage.SaveUpload(file, ps.UploadDir, "products", storage.MIMEJPEG, storage.MIMEPNG, storage.MIMEWebP)
	if err != nil {
		if errors.Is(err, storage.ErrUnsupportedType) {
			return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity, "unsupported image type")
		}
		log.WithError(err).Error("failed to save image")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload image")
	}

	if err := ps.ProductRepo.CreateImageChecked(ctx, product.ID, url, ps.MaxImages); err != nil {
		if rmErr := storage.Remove(storage.DiskPath(ps.UploadDir, url)); rmErr != nil {
			log.WithError(rmErr).Warn("failed to clean up uploaded image")
		}
		if errors.Is(err, repository.ErrImageLimit) {
			return nil, apperrors.Wrap(nil, http.StatusUnprocessableEntity,
				fmt.Sprintf("maximum %d images per product", ps.MaxImages))
		}
		log.WithError(err).Error("failed to record image")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload image")
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to upload image")
	}

	log.Info("image uploaded")
	return productResponse(detail), nil
}

func (ps *productService) UpdateImage(ctx context.Context, userID, productID, imageID uuid.UUID, request model.UpdateImageReq) (*model.ProductRes, error) {
	log := ps.logger(ctx, "UpdateImage").WithField("image_id", imageID.String())

	if err := ps.Validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return nil, err
	}

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	img, err := ps.ProductRepo.GetImage(ctx, product.ID, imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "image not found")
		}
		log.WithError(err).Error("failed to retrieve image")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update image")
	}

	fields := make(map[string]any, 3)
	if request.AltText != nil {
		fields["alt_text"] = *request.AltText
	}
	if request.SortOrder != nil {
		fields["sort_order"] = *request.SortOrder
	}

	// Primary is handled transactionally so only one image keeps the flag.
	if request.IsPrimary != nil && *request.IsPrimary && !img.IsPrimary {
		if err := ps.ProductRepo.SetPrimaryImage(ctx, product.ID, imageID); err != nil {
			log.WithError(err).Error("failed to set primary image")
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update image")
		}
	}

	if len(fields) > 0 {
		if err := ps.ProductRepo.UpdateImageFields(ctx, imageID, fields); err != nil {
			log.WithError(err).Error("failed to update image")
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update image")
		}
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to update image")
	}

	log.Info("image updated")
	return productResponse(detail), nil
}

func (ps *productService) DeleteImage(ctx context.Context, userID, productID, imageID uuid.UUID) (*model.ProductRes, error) {
	log := ps.logger(ctx, "DeleteImage").WithField("image_id", imageID.String())

	product, _, err := ownedProduct(ctx, ps.ProductRepo, ps.StoreRepo, userID, productID)
	if err != nil {
		return nil, err
	}

	img, err := ps.ProductRepo.GetImage(ctx, product.ID, imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(nil, http.StatusNotFound, "image not found")
		}
		log.WithError(err).Error("failed to retrieve image")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete image")
	}
	wasPrimary := img.IsPrimary

	if err := ps.ProductRepo.DeleteImage(ctx, img.ID); err != nil {
		log.WithError(err).Error("failed to delete image")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete image")
	}

	if err := storage.Remove(storage.DiskPath(ps.UploadDir, img.URL)); err != nil {
		log.WithError(err).Warn("failed to remove image file")
	}

	// Promote the lowest sort_order remaining image to keep one primary.
	if wasPrimary {
		images, err := ps.ProductRepo.ListImages(ctx, product.ID)
		if err != nil {
			log.WithError(err).Error("failed to list images")
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete image")
		}
		if len(images) > 0 {
			if err := ps.ProductRepo.SetPrimaryImage(ctx, product.ID, images[0].ID); err != nil {
				log.WithError(err).Error("failed to reassign primary image")
				return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete image")
			}
		}
	}

	detail, err := ps.ProductRepo.GetDetail(ctx, product.ID)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product")
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to delete image")
	}

	log.Info("image deleted")
	return productResponse(detail), nil
}

// refreshProduct recomputes min/max price or maps a failure to 500.
func (ps *productService) refreshProduct(ctx context.Context, productID uuid.UUID) error {
	if err := ps.ProductRepo.RecomputePriceRange(ctx, productID); err != nil {
		ps.Log.WithContext(ctx).WithError(err).Error("failed to compute price range")
		return apperrors.Wrap(err, http.StatusInternalServerError, "failed to update product price range")
	}
	return nil
}

// ownedProduct loads the product and verifies the caller's active store owns
// it. Foreign products answer 404 so their existence stays private.
func ownedProduct(ctx context.Context, productRepo repository.ProductRepository, storeRepo repository.StoreRepository, userID, productID uuid.UUID) (*entity.Product, *entity.Store, error) {
	product, err := productRepo.GetDetail(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperrors.Wrap(nil, http.StatusNotFound, "product not found")
		}
		return nil, nil, apperrors.Wrap(err, http.StatusInternalServerError, "failed to retrieve product")
	}

	store, err := requireActiveStore(ctx, storeRepo, userID)
	if err != nil {
		return nil, nil, err
	}

	if product.StoreID != store.ID {
		return nil, nil, apperrors.Wrap(nil, http.StatusNotFound, "product not found")
	}
	return product, store, nil
}

// validateJSONObject ensures raw is absent, "null", or a JSON object no
// larger than maxJSONObjectBytes.
func validateJSONObject(raw json.RawMessage, fieldName string) error {
	if len(raw) == 0 {
		return nil
	}
	if string(raw) == "null" {
		return nil
	}
	if len(raw) > maxJSONObjectBytes {
		return apperrors.Wrap(nil, http.StatusUnprocessableEntity,
			fieldName+" must be a JSON object no larger than 4 KB")
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return apperrors.Wrap(nil, http.StatusUnprocessableEntity, fieldName+" must be a JSON object")
	}
	return nil
}

// jsonText converts raw JSON to the string stored in a jsonb column
// ("null" clears it, absent stays nil).
func jsonText(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	s := string(raw)
	return &s
}

// duplicateProductMessage maps a duplicate key to a friendly message.
func duplicateProductMessage(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "idx_product_variants_store_sku":
			return "SKU already used in this store"
		case "idx_products_slug":
			return "product slug already taken"
		}
	}
	return "duplicate entry"
}

func productResponse(p *entity.Product) *model.ProductRes {
	res := &model.ProductRes{
		ID:          p.ID,
		StoreID:     p.StoreID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		MinPrice:    p.MinPrice,
		MaxPrice:    p.MaxPrice,
		Status:      string(p.Status),
		IsActive:    p.IsActive,
		TotalSold:   p.TotalSold,
		Metadata:    json.RawMessage(rawOrNull(p.Metadata)),
		Variants:    make([]model.VariantRes, 0, len(p.Variants)),
		Images:      make([]model.ImageRes, 0, len(p.Images)),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	for _, v := range p.Variants {
		res.Variants = append(res.Variants, model.VariantRes{
			ID:             v.ID,
			ProductID:      v.ProductID,
			SKU:            v.SKU,
			Name:           v.Name,
			Price:          v.Price,
			CompareAtPrice: v.CompareAtPrice,
			Stock:          v.Stock,
			Weight:         v.Weight,
			IsActive:       v.IsActive,
			Attributes:     json.RawMessage(rawOrNull(v.Attributes)),
			CreatedAt:      v.CreatedAt,
			UpdatedAt:      v.UpdatedAt,
		})
	}
	for _, img := range p.Images {
		res.Images = append(res.Images, model.ImageRes{
			ID:        img.ID,
			ProductID: img.ProductID,
			URL:       img.URL,
			AltText:   img.AltText,
			SortOrder: img.SortOrder,
			IsPrimary: img.IsPrimary,
			CreatedAt: img.CreatedAt,
			UpdatedAt: img.UpdatedAt,
		})
	}
	return res
}

func rawOrNull(s *string) []byte {
	if s == nil {
		return nil
	}
	return []byte(*s)
}

func (ps *productService) logger(ctx context.Context, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return ps.Log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}
