package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// Order represents a master order placed by a buyer.
// In a marketplace, one order can be split into multiple StoreOrders.
// All monetary values are in the smallest currency unit (e.g., 1 = Rp 1).
// Address fields store a JSON snapshot to preserve the address at time of purchase.
type Order struct {
	BaseModel
	UserID                  uuid.UUID   `gorm:"type:uuid;not null;index"`
	OrderNumber             string      `gorm:"type:varchar(50);uniqueIndex;not null"` // e.g., "ORD-20261023-XXXX"
	Status                  OrderStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	Subtotal                int64       `gorm:"not null;default:0"`
	ShippingTotal           int64       `gorm:"not null;default:0"`
	DiscountTotal           int64       `gorm:"not null;default:0"`
	Total                   int64       `gorm:"not null;default:0"`
	Currency                string      `gorm:"type:varchar(3);not null;default:'IDR'"`
	ShippingAddressSnapshot string      `gorm:"type:jsonb;not null"` // JSON snapshot of Address at purchase time
	BillingAddressSnapshot  *string     `gorm:"type:jsonb"`
	VoucherID               *uuid.UUID  `gorm:"type:uuid;index"`
	Notes                   *string     `gorm:"type:text"`
	PaidAt                  *time.Time

	// Relations
	User        User         `gorm:"foreignKey:UserID"`
	Voucher     *Voucher     `gorm:"foreignKey:VoucherID"`
	Items       []OrderItem  `gorm:"foreignKey:OrderID"`
	StoreOrders []StoreOrder `gorm:"foreignKey:OrderID"`
	Payments    []Payment    `gorm:"foreignKey:OrderID"`
}

// OrderItem represents a line item within an order.
// Contains price snapshots captured at the time of purchase — these never change
// even if the product's current price changes later.
type OrderItem struct {
	BaseModel
	OrderID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	StoreOrderID *uuid.UUID `gorm:"type:uuid;index"`
	VariantID    uuid.UUID  `gorm:"type:uuid;not null;index"`

	// Snapshot fields — captured at purchase time, immutable
	ProductName string `gorm:"type:varchar(255);not null"`
	VariantName string `gorm:"type:varchar(255);not null"`
	SKU         string `gorm:"type:varchar(100);not null"`
	UnitPrice   int64  `gorm:"not null"`
	Quantity    int    `gorm:"not null;default:1"`
	Total       int64  `gorm:"not null"` // UnitPrice * Quantity

	// Relations
	Order      Order          `gorm:"foreignKey:OrderID"`
	StoreOrder *StoreOrder    `gorm:"foreignKey:StoreOrderID"`
	Variant    ProductVariant `gorm:"foreignKey:VariantID"`
}

// StoreOrder represents a sub-order for a specific store within a master order.
// When a buyer purchases items from multiple sellers, the order is split into
// separate StoreOrders so each seller can manage fulfillment independently.
type StoreOrder struct {
	BaseModel
	OrderID        uuid.UUID   `gorm:"type:uuid;not null;index"`
	StoreID        uuid.UUID   `gorm:"type:uuid;not null;index"`
	Status         OrderStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	Subtotal       int64       `gorm:"not null;default:0"`
	ShippingTotal  int64       `gorm:"not null;default:0"`
	Total          int64       `gorm:"not null;default:0"`
	CommissionRate float64     `gorm:"type:decimal(5,2);not null;default:0"` // Snapshot of store's rate at purchase time
	CommissionFee  int64       `gorm:"not null;default:0"`                   // Platform fee amount

	// Relations
	Order     Order       `gorm:"foreignKey:OrderID"`
	Store     Store       `gorm:"foreignKey:StoreID"`
	Items     []OrderItem `gorm:"foreignKey:StoreOrderID"`
	Shipments []Shipment  `gorm:"foreignKey:StoreOrderID"`
}
