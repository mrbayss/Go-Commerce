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
	UserID         uuid.UUID   `gorm:"type:uuid;uniqueIndex;not null"`
	Name           string      `gorm:"type:varchar(150);not null"`
	Slug           string      `gorm:"type:varchar(150);uniqueIndex;not null"`
	Description    *string     `gorm:"type:text"`
	LogoURL        *string     `gorm:"type:varchar(500)"`
	BannerURL      *string     `gorm:"type:varchar(500)"`
	Status         StoreStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	CommissionRate float64     `gorm:"type:decimal(5,2);not null;default:0"` // e.g., 5.00 means 5%

	// Relations
	User        User         `gorm:"foreignKey:UserID"`
	Products    []Product    `gorm:"foreignKey:StoreID"`
	StoreOrders []StoreOrder `gorm:"foreignKey:StoreID"`
}
