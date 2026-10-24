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
	Code        string           `gorm:"type:varchar(50);uniqueIndex;not null"`
	Type        VoucherType      `gorm:"type:varchar(30);not null"` // entire_order, specific_product, free_shipping
	ValueType   VoucherValueType `gorm:"type:varchar(20);not null"` // fixed, percentage
	Value       int64            `gorm:"not null"`                  // Amount (fixed) or basis points (percentage, e.g., 1000 = 10%)
	MinPurchase *int64           // Minimum order subtotal to qualify
	MaxDiscount *int64           // Cap for percentage discounts
	UsageLimit  *int             // Max total uses (nil = unlimited)
	UsedCount   int              `gorm:"not null;default:0"`
	StartDate   time.Time        `gorm:"not null"`
	EndDate     time.Time        `gorm:"not null"`
	IsActive    bool             `gorm:"not null;default:true"`
	Description *string          `gorm:"type:text"`

	// Relations
	Usages []VoucherUsage `gorm:"foreignKey:VoucherID"`
}

// VoucherUsage tracks each usage of a voucher by a user for an order.
type VoucherUsage struct {
	BaseModel
	VoucherID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index"`

	// Relations
	Voucher Voucher `gorm:"foreignKey:VoucherID"`
	User    User    `gorm:"foreignKey:UserID"`
	Order   Order   `gorm:"foreignKey:OrderID"`
}
