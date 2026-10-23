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
	ProductID     *uuid.UUID `json:"product_id,omitempty" gorm:"type:uuid;index"`
	OrderID       *uuid.UUID `json:"order_id,omitempty" gorm:"type:uuid;index"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty" gorm:"index"`

	// Relations
	Product      *Product                  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Order        *Order                    `json:"-" gorm:"foreignKey:OrderID"`
	Participants []ConversationParticipant `json:"participants,omitempty" gorm:"foreignKey:ConversationID"`
	Messages     []Message                 `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// ConversationParticipant tracks users participating in a conversation
// and their read status.
type ConversationParticipant struct {
	BaseModel
	ConversationID    uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null;uniqueIndex:idx_conv_user"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_conv_user"`
	LastReadMessageID *uuid.UUID `json:"last_read_message_id,omitempty" gorm:"type:uuid"`

	// Relations
	Conversation    Conversation `json:"-" gorm:"foreignKey:ConversationID"`
	User            User         `json:"-" gorm:"foreignKey:UserID"`
	LastReadMessage *Message     `json:"-" gorm:"foreignKey:LastReadMessageID"`
}

// Message represents a single chat message in a conversation.
type Message struct {
	BaseModel
	ConversationID uuid.UUID   `json:"conversation_id" gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID   `json:"sender_id" gorm:"type:uuid;not null;index"`
	Type           MessageType `json:"type" gorm:"type:varchar(20);not null;default:'text'"` // text, image, product_card, system
	Body           string      `json:"body" gorm:"type:text;not null"`
	Metadata       *string     `json:"metadata,omitempty" gorm:"type:jsonb"` // e.g., image URL, product card data

	// Relations
	Conversation Conversation `json:"-" gorm:"foreignKey:ConversationID"`
	Sender       User         `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}
