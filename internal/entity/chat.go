package entity

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText        MessageType = "text"
	MessageTypeImage       MessageType = "image"
	MessageTypeProductCard MessageType = "product_card"
	MessageTypeSystem      MessageType = "system"
)

// Conversation represents a chat thread between a buyer and a seller.
// Can be optionally linked to a product or order for context.
type Conversation struct {
	BaseModel
	ProductID     *uuid.UUID `gorm:"type:uuid;index"`
	OrderID       *uuid.UUID `gorm:"type:uuid;index"`
	LastMessageAt *time.Time `gorm:"index"`

	// Relations
	Product      *Product                  `gorm:"foreignKey:ProductID"`
	Order        *Order                    `gorm:"foreignKey:OrderID"`
	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID"`
}

// ConversationParticipant tracks users participating in a conversation
// and their read status.
type ConversationParticipant struct {
	BaseModel
	ConversationID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user"`
	LastReadMessageID *uuid.UUID `gorm:"type:uuid"`

	// Relations
	Conversation    Conversation `gorm:"foreignKey:ConversationID"`
	User            User         `gorm:"foreignKey:UserID"`
	LastReadMessage *Message     `gorm:"foreignKey:LastReadMessageID"`
}

// Message represents a single chat message in a conversation.
type Message struct {
	BaseModel
	ConversationID uuid.UUID   `gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID   `gorm:"type:uuid;not null;index"`
	Type           MessageType `gorm:"type:varchar(20);not null;default:'text'"` // text, image, product_card, system
	Body           string      `gorm:"type:text;not null"`
	Metadata       *string     `gorm:"type:jsonb"` // e.g., image URL, product card data

	// Relations
	Conversation Conversation `gorm:"foreignKey:ConversationID"`
	Sender       User         `gorm:"foreignKey:SenderID"`
}
