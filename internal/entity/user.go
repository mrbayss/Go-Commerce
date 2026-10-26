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
	Email        string   `gorm:"type:varchar(255);uniqueIndex:idx_users_email,where:deleted_at IS NULL;not null"`
	PasswordHash *string  `gorm:"type:varchar(255)"`
	FirstName    string   `gorm:"type:varchar(100);not null"`
	LastName     string   `gorm:"type:varchar(100)"`
	Phone        *string  `gorm:"type:varchar(20);uniqueIndex:idx_users_phone,where:deleted_at IS NULL"`
	AvatarURL    *string  `gorm:"type:varchar(500)"`
	Role         UserRole `gorm:"type:varchar(20);not null;default:'buyer'"`
	IsActive     bool     `gorm:"not null;default:false"`
	IsVerified   bool     `gorm:"not null;default:false"`

	// Relations
	Addresses     []Address      `gorm:"foreignKey:UserID"`
	Store         *Store         `gorm:"foreignKey:UserID"`
	Cart          *Cart          `gorm:"foreignKey:UserID"`
	Wallet        *Wallet        `gorm:"foreignKey:UserID"`
	Orders        []Order        `gorm:"foreignKey:UserID"`
	Reviews       []Review       `gorm:"foreignKey:UserID"`
	Notifications []Notification `gorm:"foreignKey:UserID"`
}

// Address represents a shipping or billing address for a user.
// A user can have multiple addresses with one default.
type Address struct {
	BaseModel
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Label         string    `gorm:"type:varchar(50)"` // e.g., "Rumah", "Kantor"
	RecipientName string    `gorm:"type:varchar(100);not null"`
	Phone         string    `gorm:"type:varchar(20);not null"`
	AddressLine1  string    `gorm:"type:varchar(255);not null"`
	AddressLine2  *string   `gorm:"type:varchar(255)"`
	City          string    `gorm:"type:varchar(100);not null"`
	State         string    `gorm:"type:varchar(100);not null"`
	PostalCode    string    `gorm:"type:varchar(20);not null"`
	Country       string    `gorm:"type:varchar(2);not null;default:'ID'"` // ISO 3166-1 alpha-2
	IsDefault     bool      `gorm:"not null;default:false"`

	// Relations
	User User `gorm:"foreignKey:UserID"`
}
