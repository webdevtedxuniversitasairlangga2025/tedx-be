package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchImage struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	MerchandiseID uuid.UUID   `gorm:"type:uuid;not null;index"`
	ImageURL      string      `gorm:"size:255;not null"`
	Merchandise   Merchandise `gorm:"foreignKey:MerchandiseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (m *MerchImage) BeforeCreate(_ *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
