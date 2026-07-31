package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TicketTier struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TicketID    uuid.UUID       `gorm:"type:uuid;not null;index"`
	Tier        string          `gorm:"size:50;not null"`
	Price       decimal.Decimal	`gorm:"type:numeric(10,2);not null"`
	Quota       int             `gorm:"not null;default:0"`
	QuotaFilled int             `gorm:"not null;default:0"`
	SaleStart   *time.Time      `gorm:"type:timestamp with time zone"`
	SaleEnd     *time.Time      `gorm:"type:timestamp with time zone"`
	IsActive    bool            `gorm:"default:true"`
	Ticket      Ticket          `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}

func (t *TicketTier) BeforeCreate(_ *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
