package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	ReferralClaimHeld     = "held"
	ReferralClaimRedeemed = "redeemed"
)

type ReferralClaim struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ReferralCodeID   uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_referral_claim_code_email"`
	OrderID          uuid.UUID       `gorm:"type:uuid;not null;index"`
	AttendeeEmail    string          `gorm:"size:255;not null;uniqueIndex:idx_referral_claim_code_email"`
	AttendeePosition int             `gorm:"not null"`
	DiscountAmount   decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	Status           string          `gorm:"size:20;not null;index"`
	ReferralCode     ReferralCode    `gorm:"foreignKey:ReferralCodeID;constraint:OnDelete:RESTRICT"`
	Timestamp
}

func (c *ReferralClaim) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
