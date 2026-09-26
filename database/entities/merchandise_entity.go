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
	// Size & Material nullable agar baris lama tetap valid; wajib diisi saat create baru.
	Size        *string         `gorm:"size:100"`
	Material    *string         `gorm:"size:255"`
	// GformURL link pembelian per produk; kosong → pakai link GForm global di FE.
	GformURL    *string         `gorm:"size:500"`
	CategoryID  uuid.UUID       `gorm:"type:uuid;not null;index"`
	IsActive    bool            `gorm:"default:true"`

	Category    Category           `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
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