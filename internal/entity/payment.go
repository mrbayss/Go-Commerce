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
	OrderID              uuid.UUID     `gorm:"type:uuid;not null;index"`
	Method               PaymentMethod `gorm:"type:varchar(20);not null"` // gateway, wallet, cod
	Gateway              *string       `gorm:"type:varchar(50)"`          // midtrans, xendit, stripe
	GatewayTransactionID *string       `gorm:"type:varchar(255);uniqueIndex"`
	Amount               int64         `gorm:"not null"`
	Currency             string        `gorm:"type:varchar(3);not null;default:'IDR'"`
	Status               PaymentStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	PaidAt               *time.Time
	ExpiredAt            *time.Time
	Metadata             *string `gorm:"type:jsonb"` // Gateway-specific response data

	// Relations
	Order Order `gorm:"foreignKey:OrderID"`
}
