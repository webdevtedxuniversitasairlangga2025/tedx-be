package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Merchandise struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string          `gorm:"size:255;not null"`
	Description string          `gorm:"type:text;not null"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	Category    string          `gorm:"size:50;not null"`
	IsActive    bool            `gorm:"default:true"`

	MerchImages []MerchImage

	Timestamp
}

func (Merchandise) TableName() string {
	return "merchandise"
}

func (m *Merchandise) BeforeCreate(_ *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}