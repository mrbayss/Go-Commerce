package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRead    NotificationStatus = "read"
)

// Notification represents a notification sent to a user.
// Supports multiple channels: in-app, email, SMS, push.
type Notification struct {
	BaseModel
	UserID  uuid.UUID           `gorm:"type:uuid;not null;index"`
	Type    string              `gorm:"type:varchar(50);not null;index"` // e.g., "order_paid", "order_shipped", "new_message"
	Title   string              `gorm:"type:varchar(255);not null"`
	Body    string              `gorm:"type:text;not null"`
	Channel NotificationChannel `gorm:"type:varchar(20);not null;default:'in_app'"` // in_app, email, sms, push
	Payload *string             `gorm:"type:jsonb"`                                 // Additional data (e.g., order_id, conversation_id)
	Status  NotificationStatus  `gorm:"type:varchar(20);not null;default:'pending';index"`
	ReadAt  *time.Time

	// Relations
	User User `gorm:"foreignKey:UserID"`
}
