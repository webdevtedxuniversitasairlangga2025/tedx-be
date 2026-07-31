package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Bundle struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string          `gorm:"size:255;not null"`
	Description string          `gorm:"type:text;not null"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	IsActive    bool            `gorm:"default:true"`

	Timestamp
}

func (b *Bundle) BeforeCreate(_ *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
