package entity

import (
	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleSeller UserRole = "seller"
	UserRoleBuyer  UserRole = "buyer"
)

// User represents a registered user in the system.
// A user can be a buyer, seller, or admin based on their role.
type User struct {
	BaseModel
	Email        string   `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string   `json:"-" gorm:"type:varchar(255);not null"`
	FirstName    string   `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string   `json:"last_name" gorm:"type:varchar(100)"`
	Phone        *string  `json:"phone,omitempty" gorm:"type:varchar(20);uniqueIndex"`
	AvatarURL    *string  `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
	Role         UserRole `json:"role" gorm:"type:varchar(20);not null;default:'buyer'"`
	IsActive     bool     `json:"is_active" gorm:"not null;default:true"`

	// Relations
	Addresses     []Address      `json:"addresses,omitempty" gorm:"foreignKey:UserID"`
	Store         *Store         `json:"store,omitempty" gorm:"foreignKey:UserID"`
	Cart          *Cart          `json:"cart,omitempty" gorm:"foreignKey:UserID"`
	Wallet        *Wallet        `json:"wallet,omitempty" gorm:"foreignKey:UserID"`
	Orders        []Order        `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	Reviews       []Review       `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
}

// Address represents a shipping or billing address for a user.
// A user can have multiple addresses with one default.
type Address struct {
	BaseModel
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Label         string    `json:"label" gorm:"type:varchar(50)"` // e.g., "Rumah", "Kantor"
	RecipientName string    `json:"recipient_name" gorm:"type:varchar(100);not null"`
	Phone         string    `json:"phone" gorm:"type:varchar(20);not null"`
	AddressLine1  string    `json:"address_line_1" gorm:"type:varchar(255);not null"`
	AddressLine2  *string   `json:"address_line_2,omitempty" gorm:"type:varchar(255)"`
	City          string    `json:"city" gorm:"type:varchar(100);not null"`
	State         string    `json:"state" gorm:"type:varchar(100);not null"`
	PostalCode    string    `json:"postal_code" gorm:"type:varchar(20);not null"`
	Country       string    `json:"country" gorm:"type:varchar(2);not null;default:'ID'"` // ISO 3166-1 alpha-2
	IsDefault     bool      `json:"is_default" gorm:"not null;default:false"`

	// Relations
	User User `json:"-" gorm:"foreignKey:UserID"`
}
