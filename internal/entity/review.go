package entity

import (
	"github.com/google/uuid"
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

// Review represents a product review/rating left by a buyer.
// Only verified purchases (buyer who completed an order) should be marked as verified.
type Review struct {
	BaseModel
	UserID     uuid.UUID    `gorm:"type:uuid;not null;index"`
	ProductID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	OrderID    *uuid.UUID   `gorm:"type:uuid;index"`
	Rating     int          `gorm:"not null"` // 1-5
	Title      *string      `gorm:"type:varchar(255)"`
	Comment    *string      `gorm:"type:text"`
	IsVerified bool         `gorm:"not null;default:false"` // Verified purchase
	Status     ReviewStatus `gorm:"type:varchar(20);not null;default:'pending';index"`

	// Relations
	User    User    `gorm:"foreignKey:UserID"`
	Product Product `gorm:"foreignKey:ProductID"`
	Order   *Order  `gorm:"foreignKey:OrderID"`
}
