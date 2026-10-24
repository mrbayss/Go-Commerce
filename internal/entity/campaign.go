package entity

import (
	"time"

	"github.com/google/uuid"
)

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusEnded     CampaignStatus = "ended"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type FlashSaleItemStatus string

const (
	FlashSaleItemStatusActive   FlashSaleItemStatus = "active"
	FlashSaleItemStatusSoldOut  FlashSaleItemStatus = "sold_out"
	FlashSaleItemStatusInactive FlashSaleItemStatus = "inactive"
)

// Campaign represents a promotional campaign or flash sale event.
// A campaign groups multiple flash sale items under a shared time window.
type Campaign struct {
	BaseModel
	Title       string         `gorm:"type:varchar(255);not null"`
	Description *string        `gorm:"type:text"`
	BannerURL   *string        `gorm:"type:varchar(500)"`
	StartTime   time.Time      `gorm:"not null;index"`
	EndTime     time.Time      `gorm:"not null;index"`
	Status      CampaignStatus `gorm:"type:varchar(20);not null;default:'draft';index"`

	// Relations
	FlashSaleItems []FlashSaleItem `gorm:"foreignKey:CampaignID"`
}

// FlashSaleItem represents a specific product variant offered in a campaign
// with a special flash price and limited stock.
type FlashSaleItem struct {
	BaseModel
	CampaignID     uuid.UUID           `gorm:"type:uuid;not null;index"`
	VariantID      uuid.UUID           `gorm:"type:uuid;not null;index"`
	OriginalPrice  int64               `gorm:"not null"`
	FlashPrice     int64               `gorm:"not null"`
	TotalStock     int                 `gorm:"not null"`
	AvailableStock int                 `gorm:"not null"`
	MaxPerUser     int                 `gorm:"not null;default:1"`
	Status         FlashSaleItemStatus `gorm:"type:varchar(20);not null;default:'active'"`

	// Relations
	Campaign Campaign       `gorm:"foreignKey:CampaignID"`
	Variant  ProductVariant `gorm:"foreignKey:VariantID"`
}
