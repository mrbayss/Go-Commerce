package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mrbayss/Go-Commerce/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrLastVariant signals the product would be left with no variants at all.
	ErrLastVariant = errors.New("cannot delete the last variant")
	// ErrLastActiveVariant signals the product would lose its last active variant.
	ErrLastActiveVariant = errors.New("cannot delete the last active variant")
	// ErrImageLimit signals the product already has the maximum number of images.
	ErrImageLimit = errors.New("image limit reached")
)

type ProductRepository interface {
	Repository[entity.Product]
	CreateWithVariants(ctx context.Context, p *entity.Product, variants []entity.ProductVariant) (*entity.Product, error)
	GetDetail(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	ListByStore(ctx context.Context, storeID uuid.UUID, page, limit int) ([]entity.Product, int64, error)
	CountImages(ctx context.Context, productID uuid.UUID) (int64, error)
	RecomputePriceRange(ctx context.Context, productID uuid.UUID) error
	DeleteChildren(ctx context.Context, productID uuid.UUID) error
	CreateVariant(ctx context.Context, v *entity.ProductVariant) error
	GetVariant(ctx context.Context, productID, variantID uuid.UUID) (*entity.ProductVariant, error)
	UpdateVariantFields(ctx context.Context, variantID uuid.UUID, fields map[string]any) error
	DeleteVariantChecked(ctx context.Context, productID, variantID uuid.UUID) error
	CreateImageChecked(ctx context.Context, productID uuid.UUID, url string, maxImages int) error
	GetImage(ctx context.Context, productID, imageID uuid.UUID) (*entity.ProductImage, error)
	ListImages(ctx context.Context, productID uuid.UUID) ([]entity.ProductImage, error)
	UpdateImageFields(ctx context.Context, imageID uuid.UUID, fields map[string]any) error
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
	SetPrimaryImage(ctx context.Context, productID, imageID uuid.UUID) error
}

type productRepository struct {
	RepositoryImpl[entity.Product]
	db *gorm.DB
}

// NewProductRepository creates a new product repository.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		RepositoryImpl: RepositoryImpl[entity.Product]{db: db},
		db:             db,
	}
}

// CreateWithVariants inserts the product and all its variants atomically.
func (r *productRepository) CreateWithVariants(ctx context.Context, p *entity.Product, variants []entity.ProductVariant) (*entity.Product, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for i := range variants {
			variants[i].ProductID = p.ID
			variants[i].StoreID = p.StoreID
		}
		if len(variants) > 0 {
			if err := tx.Create(&variants).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return p, err
}

// GetDetail loads a product with its variants and images preloaded.
func (r *productRepository) GetDetail(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var p entity.Product
	err := r.db.WithContext(ctx).
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, created_at ASC") }).
		First(&p, "id = ?", id).Error
	return &p, err
}

func (r *productRepository) ListByStore(ctx context.Context, storeID uuid.UUID, page, limit int) ([]entity.Product, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.Product{}).Where("store_id = ?", storeID)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	products := make([]entity.Product, 0)
	if err := tx.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) CountImages(ctx context.Context, productID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ProductImage{}).Where("product_id = ?", productID).Count(&count).Error
	return count, err
}

// RecomputePriceRange refreshes the min/max price cache from active
// variants in a single atomic UPDATE (COALESCE -> 0 when none remain).
func (r *productRepository) RecomputePriceRange(ctx context.Context, productID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", productID).Updates(map[string]any{
		"min_price": gorm.Expr("(SELECT COALESCE(MIN(price), 0) FROM product_variants WHERE product_id = ? AND is_active = ? AND deleted_at IS NULL)", productID, true),
		"max_price": gorm.Expr("(SELECT COALESCE(MAX(price), 0) FROM product_variants WHERE product_id = ? AND is_active = ? AND deleted_at IS NULL)", productID, true),
	}).Error
}

// DeleteChildren soft-deletes every variant and image of a product atomically.
func (r *productRepository) DeleteChildren(ctx context.Context, productID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.ProductVariant{}, "product_id = ?", productID).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.ProductImage{}, "product_id = ?", productID).Error
	})
}

func (r *productRepository) CreateVariant(ctx context.Context, v *entity.ProductVariant) error {
	return r.db.WithContext(ctx).Create(v).Error
}

// GetVariant scopes the lookup to the product so foreign variants stay 404.
func (r *productRepository) GetVariant(ctx context.Context, productID, variantID uuid.UUID) (*entity.ProductVariant, error) {
	var v entity.ProductVariant
	err := r.db.WithContext(ctx).First(&v, "id = ? AND product_id = ?", variantID, productID).Error
	return &v, err
}

func (r *productRepository) UpdateVariantFields(ctx context.Context, variantID uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.ProductVariant{}).Where("id = ?", variantID).Updates(fields).Error
}

// DeleteVariantChecked locks the product row, verifies the variant belongs to
// it, enforces both "last variant" rules, and deletes — all in one
// transaction, so concurrent deletes cannot race past the guard.
//
// Rules (design: a product must always keep at least one variant; all-
// inactive is allowed):
//   - last variant overall  -> ErrLastVariant
//   - active variant that is the last active one -> ErrLastActiveVariant
func (r *productRepository) DeleteVariantChecked(ctx context.Context, productID, variantID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&entity.Product{}, "id = ?", productID).Error; err != nil {
			return err
		}

		var v entity.ProductVariant
		if err := tx.First(&v, "id = ? AND product_id = ?", variantID, productID).Error; err != nil {
			return err
		}

		var total int64
		if err := tx.Model(&entity.ProductVariant{}).Where("product_id = ?", productID).Count(&total).Error; err != nil {
			return err
		}
		if v.IsActive {
			var active int64
			if err := tx.Model(&entity.ProductVariant{}).
				Where("product_id = ? AND is_active = ?", productID, true).
				Count(&active).Error; err != nil {
				return err
			}
			if active <= 1 {
				return ErrLastActiveVariant
			}
		}
		if total <= 1 {
			return ErrLastVariant
		}
		return tx.Delete(&entity.ProductVariant{}, "id = ?", variantID).Error
	})
}

// CreateImageChecked locks the product row, enforces the image limit, and
// inserts the image in one transaction — concurrent uploads cannot exceed
// maxImages. SortOrder and IsPrimary are derived under the lock.
func (r *productRepository) CreateImageChecked(ctx context.Context, productID uuid.UUID, url string, maxImages int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&entity.Product{}, "id = ?", productID).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&entity.ProductImage{}).Where("product_id = ?", productID).Count(&count).Error; err != nil {
			return err
		}
		if count >= int64(maxImages) {
			return ErrImageLimit
		}
		return tx.Create(&entity.ProductImage{
			ProductID: productID,
			URL:       url,
			SortOrder: int(count),
			IsPrimary: count == 0,
		}).Error
	})
}

// GetImage scopes the lookup to the product so foreign images stay 404.
func (r *productRepository) GetImage(ctx context.Context, productID, imageID uuid.UUID) (*entity.ProductImage, error) {
	var img entity.ProductImage
	err := r.db.WithContext(ctx).First(&img, "id = ? AND product_id = ?", imageID, productID).Error
	return &img, err
}

func (r *productRepository) ListImages(ctx context.Context, productID uuid.UUID) ([]entity.ProductImage, error) {
	var imgs []entity.ProductImage
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("sort_order ASC").Find(&imgs).Error
	return imgs, err
}

func (r *productRepository) UpdateImageFields(ctx context.Context, imageID uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.ProductImage{}).Where("id = ?", imageID).Updates(fields).Error
}

func (r *productRepository) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.ProductImage{}, "id = ?", imageID).Error
}

// SetPrimaryImage atomically unsets every primary flag of the product, then
// sets it on imageID — guaranteeing at most one primary per product.
func (r *productRepository) SetPrimaryImage(ctx context.Context, productID, imageID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.ProductImage{}).Where("product_id = ?", productID).Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Model(&entity.ProductImage{}).Where("id = ?", imageID).Update("is_primary", true).Error
	})
}
