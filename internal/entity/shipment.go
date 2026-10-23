package entity

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	ShipmentStatusPending   ShipmentStatus = "pending"
	ShipmentStatusPickedUp  ShipmentStatus = "picked_up"
	ShipmentStatusInTransit ShipmentStatus = "in_transit"
	ShipmentStatusDelivered ShipmentStatus = "delivered"
	ShipmentStatusReturned  ShipmentStatus = "returned"
	ShipmentStatusCancelled ShipmentStatus = "cancelled"
)

// Shipment represents a shipping/delivery record for a store order.
// Tracks courier, tracking number, and delivery status.
type Shipment struct {
	BaseModel
	StoreOrderID      uuid.UUID      `json:"store_order_id" gorm:"type:uuid;not null;index"`
	Courier           string         `json:"courier" gorm:"type:varchar(50);not null"`          // e.g., "jne", "jnt", "sicepat", "gosend"
	CourierService    *string        `json:"courier_service,omitempty" gorm:"type:varchar(50)"` // e.g., "REG", "YES", "OKE"
	TrackingNumber    *string        `json:"tracking_number,omitempty" gorm:"type:varchar(100);index"`
	TrackingURL       *string        `json:"tracking_url,omitempty" gorm:"type:varchar(500)"`
	Status            ShipmentStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	ShippingCost      int64          `json:"shipping_cost" gorm:"not null;default:0"`
	ShippedAt         *time.Time     `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time     `json:"delivered_at,omitempty"`
	EstimatedDelivery *time.Time     `json:"estimated_delivery,omitempty"`

	// Relations
	StoreOrder StoreOrder `json:"-" gorm:"foreignKey:StoreOrderID"`
}
