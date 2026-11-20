package entity

import "github.com/google/uuid"

type StoreStatus string

const (
	StoreStatusPending   StoreStatus = "pending"
	StoreStatusActive    StoreStatus = "active"
	StoreStatusRejected  StoreStatus = "rejected"
	StoreStatusSuspended StoreStatus = "suspended"
)

// Store is a seller's shop. One user can own at most one store.
// Registration starts as pending and is approved by an admin.
type Store struct {
	BaseModel
	UserID         uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_stores_user_id,where:deleted_at IS NULL"`
	Name           string      `gorm:"type:varchar(150);not null"`
	Slug           string      `gorm:"type:varchar(150);not null;uniqueIndex:idx_stores_slug,where:deleted_at IS NULL"`
	Description    *string     `gorm:"type:text"`
	LogoURL        *string     `gorm:"type:varchar(500)"`
	BannerURL      *string     `gorm:"type:varchar(500)"`
	Status         StoreStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	CommissionRate float64     `gorm:"type:decimal(5,2);not null;default:0"`

	User     User      `gorm:"foreignKey:UserID"`
	Products []Product `gorm:"foreignKey:StoreID"`
}
