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
	// Size & Material nullable agar baris lama tetap valid; wajib diisi saat create baru.
	Size        *string         `gorm:"size:100"`
	Material    *string         `gorm:"size:255"`
	// GformURL link pembelian per bundle; kosong → pakai link GForm global di FE.
	GformURL    *string         `gorm:"size:500"`
	IsActive    bool            `gorm:"default:true"`

	BundleImages []BundleImage `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}

func (b *Bundle) BeforeCreate(_ *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
