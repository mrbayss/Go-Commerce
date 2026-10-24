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
	StoreOrderID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	Courier           string         `gorm:"type:varchar(50);not null"` // e.g., "jne", "jnt", "sicepat", "gosend"
	CourierService    *string        `gorm:"type:varchar(50)"`          // e.g., "REG", "YES", "OKE"
	TrackingNumber    *string        `gorm:"type:varchar(100);index"`
	TrackingURL       *string        `gorm:"type:varchar(500)"`
	Status            ShipmentStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	ShippingCost      int64          `gorm:"not null;default:0"`
	ShippedAt         *time.Time
	DeliveredAt       *time.Time
	EstimatedDelivery *time.Time

	// Relations
	StoreOrder StoreOrder `gorm:"foreignKey:StoreOrderID"`
}
