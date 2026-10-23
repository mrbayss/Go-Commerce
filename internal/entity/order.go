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
	UserID                  uuid.UUID   `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderNumber             string      `json:"order_number" gorm:"type:varchar(50);uniqueIndex;not null"` // e.g., "ORD-20261023-XXXX"
	Status                  OrderStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Subtotal                int64       `json:"subtotal" gorm:"not null;default:0"`
	ShippingTotal           int64       `json:"shipping_total" gorm:"not null;default:0"`
	DiscountTotal           int64       `json:"discount_total" gorm:"not null;default:0"`
	Total                   int64       `json:"total" gorm:"not null;default:0"`
	Currency                string      `json:"currency" gorm:"type:varchar(3);not null;default:'IDR'"`
	ShippingAddressSnapshot string      `json:"shipping_address_snapshot" gorm:"type:jsonb;not null"` // JSON snapshot of Address at purchase time
	BillingAddressSnapshot  *string     `json:"billing_address_snapshot,omitempty" gorm:"type:jsonb"`
	VoucherID               *uuid.UUID  `json:"voucher_id,omitempty" gorm:"type:uuid;index"`
	Notes                   *string     `json:"notes,omitempty" gorm:"type:text"`
	PaidAt                  *time.Time  `json:"paid_at,omitempty"`

	// Relations
	User        User         `json:"-" gorm:"foreignKey:UserID"`
	Voucher     *Voucher     `json:"voucher,omitempty" gorm:"foreignKey:VoucherID"`
	Items       []OrderItem  `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	StoreOrders []StoreOrder `json:"store_orders,omitempty" gorm:"foreignKey:OrderID"`
	Payments    []Payment    `json:"payments,omitempty" gorm:"foreignKey:OrderID"`
}

// OrderItem represents a line item within an order.
// Contains price snapshots captured at the time of purchase — these never change
// even if the product's current price changes later.
type OrderItem struct {
	BaseModel
	OrderID      uuid.UUID  `json:"order_id" gorm:"type:uuid;not null;index"`
	StoreOrderID *uuid.UUID `json:"store_order_id,omitempty" gorm:"type:uuid;index"`
	VariantID    uuid.UUID  `json:"variant_id" gorm:"type:uuid;not null;index"`

	// Snapshot fields — captured at purchase time, immutable
	ProductName string `json:"product_name" gorm:"type:varchar(255);not null"`
	VariantName string `json:"variant_name" gorm:"type:varchar(255);not null"`
	SKU         string `json:"sku" gorm:"type:varchar(100);not null"`
	UnitPrice   int64  `json:"unit_price" gorm:"not null"`
	Quantity    int    `json:"quantity" gorm:"not null;default:1"`
	Total       int64  `json:"total" gorm:"not null"` // UnitPrice * Quantity

	// Relations
	Order      Order          `json:"-" gorm:"foreignKey:OrderID"`
	StoreOrder *StoreOrder    `json:"-" gorm:"foreignKey:StoreOrderID"`
	Variant    ProductVariant `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
}

// StoreOrder represents a sub-order for a specific store within a master order.
// When a buyer purchases items from multiple sellers, the order is split into
// separate StoreOrders so each seller can manage fulfillment independently.
type StoreOrder struct {
	BaseModel
	OrderID        uuid.UUID   `json:"order_id" gorm:"type:uuid;not null;index"`
	StoreID        uuid.UUID   `json:"store_id" gorm:"type:uuid;not null;index"`
	Status         OrderStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Subtotal       int64       `json:"subtotal" gorm:"not null;default:0"`
	ShippingTotal  int64       `json:"shipping_total" gorm:"not null;default:0"`
	Total          int64       `json:"total" gorm:"not null;default:0"`
	CommissionRate float64     `json:"commission_rate" gorm:"type:decimal(5,2);not null;default:0"` // Snapshot of store's rate at purchase time
	CommissionFee  int64       `json:"commission_fee" gorm:"not null;default:0"`                    // Platform fee amount

	// Relations
	Order     Order       `json:"-" gorm:"foreignKey:OrderID"`
	Store     Store       `json:"store,omitempty" gorm:"foreignKey:StoreID"`
	Items     []OrderItem `json:"items,omitempty" gorm:"foreignKey:StoreOrderID"`
	Shipments []Shipment  `json:"shipments,omitempty" gorm:"foreignKey:StoreOrderID"`
}
