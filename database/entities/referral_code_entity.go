package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ReferralCode struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code              string          `gorm:"size:100;uniqueIndex;not null"`
	DiscountPerTicket decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	Quota             int             `gorm:"not null"`
	QuotaHeld         int             `gorm:"not null;default:0"`
	QuotaUsed         int             `gorm:"not null;default:0"`
	IsActive          bool            `gorm:"not null;default:true"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"deleted_at"`
	Claims            []ReferralClaim `gorm:"foreignKey:ReferralCodeID"`
	Timestamp
}

func (c *ReferralCode) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
