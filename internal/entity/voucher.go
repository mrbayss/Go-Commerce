package entity

import (
	"time"

	"github.com/google/uuid"
)

type VoucherType string

const (
	VoucherTypeEntireOrder     VoucherType = "entire_order"
	VoucherTypeSpecificProduct VoucherType = "specific_product"
	VoucherTypeFreeShipping    VoucherType = "free_shipping"
)

type VoucherValueType string

const (
	VoucherValueTypeFixed      VoucherValueType = "fixed"
	VoucherValueTypePercentage VoucherValueType = "percentage"
)

// Voucher represents a discount coupon code.
// Supports fixed amount or percentage discounts applied to orders, specific products, or shipping.
type Voucher struct {
	BaseModel
	Code        string           `json:"code" gorm:"type:varchar(50);uniqueIndex;not null"`
	Type        VoucherType      `json:"type" gorm:"type:varchar(30);not null"`       // entire_order, specific_product, free_shipping
	ValueType   VoucherValueType `json:"value_type" gorm:"type:varchar(20);not null"` // fixed, percentage
	Value       int64            `json:"value" gorm:"not null"`                       // Amount (fixed) or basis points (percentage, e.g., 1000 = 10%)
	MinPurchase *int64           `json:"min_purchase,omitempty"`                      // Minimum order subtotal to qualify
	MaxDiscount *int64           `json:"max_discount,omitempty"`                      // Cap for percentage discounts
	UsageLimit  *int             `json:"usage_limit,omitempty"`                       // Max total uses (nil = unlimited)
	UsedCount   int              `json:"used_count" gorm:"not null;default:0"`
	StartDate   time.Time        `json:"start_date" gorm:"not null"`
	EndDate     time.Time        `json:"end_date" gorm:"not null"`
	IsActive    bool             `json:"is_active" gorm:"not null;default:true"`
	Description *string          `json:"description,omitempty" gorm:"type:text"`

	// Relations
	Usages []VoucherUsage `json:"usages,omitempty" gorm:"foreignKey:VoucherID"`
}

// VoucherUsage tracks each usage of a voucher by a user for an order.
type VoucherUsage struct {
	BaseModel
	VoucherID uuid.UUID `json:"voucher_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`

	// Relations
	Voucher Voucher `json:"-" gorm:"foreignKey:VoucherID"`
	User    User    `json:"-" gorm:"foreignKey:UserID"`
	Order   Order   `json:"-" gorm:"foreignKey:OrderID"`
}
