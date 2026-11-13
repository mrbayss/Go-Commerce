package entity

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
}
