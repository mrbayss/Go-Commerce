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
	Title       string         `json:"title" gorm:"type:varchar(255);not null"`
	Description *string        `json:"description,omitempty" gorm:"type:text"`
	BannerURL   *string        `json:"banner_url,omitempty" gorm:"type:varchar(500)"`
	StartTime   time.Time      `json:"start_time" gorm:"not null;index"`
	EndTime     time.Time      `json:"end_time" gorm:"not null;index"`
	Status      CampaignStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft';index"`

	// Relations
	FlashSaleItems []FlashSaleItem `json:"flash_sale_items,omitempty" gorm:"foreignKey:CampaignID"`
}

// FlashSaleItem represents a specific product variant offered in a campaign
// with a special flash price and limited stock.
type FlashSaleItem struct {
	BaseModel
	CampaignID     uuid.UUID           `json:"campaign_id" gorm:"type:uuid;not null;index"`
	VariantID      uuid.UUID           `json:"variant_id" gorm:"type:uuid;not null;index"`
	OriginalPrice  int64               `json:"original_price" gorm:"not null"`
	FlashPrice     int64               `json:"flash_price" gorm:"not null"`
	TotalStock     int                 `json:"total_stock" gorm:"not null"`
	AvailableStock int                 `json:"available_stock" gorm:"not null"`
	MaxPerUser     int                 `json:"max_per_user" gorm:"not null;default:1"`
	Status         FlashSaleItemStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`

	// Relations
	Campaign Campaign       `json:"-" gorm:"foreignKey:CampaignID"`
	Variant  ProductVariant `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
}
