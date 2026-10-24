package entity

import (
	"github.com/google/uuid"
)

type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "active"
	WalletStatusFrozen    WalletStatus = "frozen"
	WalletStatusSuspended WalletStatus = "suspended"
)

type WalletTxType string

const (
	WalletTxTypeTopUp    WalletTxType = "top_up"
	WalletTxTypePayment  WalletTxType = "payment"
	WalletTxTypeRefund   WalletTxType = "refund"
	WalletTxTypeTransfer WalletTxType = "transfer"
)

type WalletTxStatus string

const (
	WalletTxStatusPending   WalletTxStatus = "pending"
	WalletTxStatusCompleted WalletTxStatus = "completed"
	WalletTxStatusFailed    WalletTxStatus = "failed"
	WalletTxStatusReversed  WalletTxStatus = "reversed"
)

// Wallet represents a user's digital wallet / internal balance.
// Each user has one wallet.
type Wallet struct {
	BaseModel
	UserID   uuid.UUID    `gorm:"type:uuid;uniqueIndex;not null"`
	Balance  int64        `gorm:"not null;default:0"` // In smallest currency unit
	Currency string       `gorm:"type:varchar(3);not null;default:'IDR'"`
	Status   WalletStatus `gorm:"type:varchar(20);not null;default:'active'"`

	// Relations
	User         User                `gorm:"foreignKey:UserID"`
	Transactions []WalletTransaction `gorm:"foreignKey:WalletID"`
}

// WalletTransaction represents an immutable ledger entry for wallet balance changes.
// Follows the transaction ledger pattern — entries are never updated or deleted.
type WalletTransaction struct {
	BaseModel
	WalletID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	Type           WalletTxType   `gorm:"type:varchar(20);not null"` // top_up, payment, refund, transfer
	Amount         int64          `gorm:"not null"`                  // Positive = credit, Negative = debit
	BalanceAfter   int64          `gorm:"not null"`                  // Wallet balance after this transaction
	ReferenceType  *string        `gorm:"type:varchar(50)"`          // order, refund, top_up, etc.
	ReferenceID    *string        `gorm:"type:varchar(255)"`         // Related entity ID
	Description    string         `gorm:"type:varchar(255);not null"`
	IdempotencyKey string         `gorm:"type:varchar(255);uniqueIndex;not null"` // Prevents duplicate transactions
	Status         WalletTxStatus `gorm:"type:varchar(20);not null;default:'completed'"`

	// Relations
	Wallet Wallet `gorm:"foreignKey:WalletID"`
}
