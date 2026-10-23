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
	UserID   uuid.UUID    `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	Balance  int64        `json:"balance" gorm:"not null;default:0"` // In smallest currency unit
	Currency string       `json:"currency" gorm:"type:varchar(3);not null;default:'IDR'"`
	Status   WalletStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`

	// Relations
	User         User                `json:"-" gorm:"foreignKey:UserID"`
	Transactions []WalletTransaction `json:"transactions,omitempty" gorm:"foreignKey:WalletID"`
}

// WalletTransaction represents an immutable ledger entry for wallet balance changes.
// Follows the transaction ledger pattern — entries are never updated or deleted.
type WalletTransaction struct {
	BaseModel
	WalletID       uuid.UUID      `json:"wallet_id" gorm:"type:uuid;not null;index"`
	Type           WalletTxType   `json:"type" gorm:"type:varchar(20);not null"`            // top_up, payment, refund, transfer
	Amount         int64          `json:"amount" gorm:"not null"`                           // Positive = credit, Negative = debit
	BalanceAfter   int64          `json:"balance_after" gorm:"not null"`                    // Wallet balance after this transaction
	ReferenceType  *string        `json:"reference_type,omitempty" gorm:"type:varchar(50)"` // order, refund, top_up, etc.
	ReferenceID    *string        `json:"reference_id,omitempty" gorm:"type:varchar(255)"`  // Related entity ID
	Description    string         `json:"description" gorm:"type:varchar(255);not null"`
	IdempotencyKey string         `json:"idempotency_key" gorm:"type:varchar(255);uniqueIndex;not null"` // Prevents duplicate transactions
	Status         WalletTxStatus `json:"status" gorm:"type:varchar(20);not null;default:'completed'"`

	// Relations
	Wallet Wallet `json:"-" gorm:"foreignKey:WalletID"`
}
