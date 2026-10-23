package entity

import (
	"github.com/google/uuid"
)

type StoreStatus string

const (
	StoreStatusPending   StoreStatus = "pending"
	StoreStatusActive    StoreStatus = "active"
	StoreStatusRejected  StoreStatus = "rejected"
	StoreStatusSuspended StoreStatus = "suspended"
)

// Store represents a seller's store in the marketplace.
// Each seller (user with role "seller") can have one store.
type Store struct {
	BaseModel
	UserID         uuid.UUID   `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	Name           string      `json:"name" gorm:"type:varchar(150);not null"`
	Slug           string      `json:"slug" gorm:"type:varchar(150);uniqueIndex;not null"`
	Description    *string     `json:"description,omitempty" gorm:"type:text"`
	LogoURL        *string     `json:"logo_url,omitempty" gorm:"type:varchar(500)"`
	BannerURL      *string     `json:"banner_url,omitempty" gorm:"type:varchar(500)"`
	Status         StoreStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	CommissionRate float64     `json:"commission_rate" gorm:"type:decimal(5,2);not null;default:0"` // e.g., 5.00 means 5%

	// Relations
	User        User         `json:"-" gorm:"foreignKey:UserID"`
	Products    []Product    `json:"products,omitempty" gorm:"foreignKey:StoreID"`
	StoreOrders []StoreOrder `json:"store_orders,omitempty" gorm:"foreignKey:StoreID"`
}
