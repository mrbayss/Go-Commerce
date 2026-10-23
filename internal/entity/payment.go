package entity

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodGateway PaymentMethod = "gateway" // Midtrans, Xendit, Stripe, etc.
	PaymentMethodWallet  PaymentMethod = "wallet"
	PaymentMethodCOD     PaymentMethod = "cod"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusVoided     PaymentStatus = "voided"
)

// Payment represents a payment attempt or transaction for an order.
// Supports multiple payment methods: gateway (Midtrans/Xendit/Stripe), wallet, COD.
type Payment struct {
	BaseModel
	OrderID              uuid.UUID     `json:"order_id" gorm:"type:uuid;not null;index"`
	Method               PaymentMethod `json:"method" gorm:"type:varchar(20);not null"`   // gateway, wallet, cod
	Gateway              *string       `json:"gateway,omitempty" gorm:"type:varchar(50)"` // midtrans, xendit, stripe
	GatewayTransactionID *string       `json:"gateway_transaction_id,omitempty" gorm:"type:varchar(255);uniqueIndex"`
	Amount               int64         `json:"amount" gorm:"not null"`
	Currency             string        `json:"currency" gorm:"type:varchar(3);not null;default:'IDR'"`
	Status               PaymentStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	PaidAt               *time.Time    `json:"paid_at,omitempty"`
	ExpiredAt            *time.Time    `json:"expired_at,omitempty"`
	Metadata             *string       `json:"metadata,omitempty" gorm:"type:jsonb"` // Gateway-specific response data

	// Relations
	Order Order `json:"-" gorm:"foreignKey:OrderID"`
}
