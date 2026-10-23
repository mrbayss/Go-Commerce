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
	UserID     uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID  uuid.UUID    `json:"product_id" gorm:"type:uuid;not null;index"`
	OrderID    *uuid.UUID   `json:"order_id,omitempty" gorm:"type:uuid;index"`
	Rating     int          `json:"rating" gorm:"not null"` // 1-5
	Title      *string      `json:"title,omitempty" gorm:"type:varchar(255)"`
	Comment    *string      `json:"comment,omitempty" gorm:"type:text"`
	IsVerified bool         `json:"is_verified" gorm:"not null;default:false"` // Verified purchase
	Status     ReviewStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`

	// Relations
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product Product `json:"-" gorm:"foreignKey:ProductID"`
	Order   *Order  `json:"-" gorm:"foreignKey:OrderID"`
}
