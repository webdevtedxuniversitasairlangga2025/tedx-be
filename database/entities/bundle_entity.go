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

	// Tag constraint harus ada di sisi has-many ini: GORM membuat foreign key
	// bundle_images.bundle_id dari relasi ini (fk_bundles_bundle_images), bukan
	// dari tag di BundleImage.Bundle. Tanpa tag ini constraint-nya NO ACTION,
	// sehingga menghapus bundle yang punya gambar akan gagal — padahal
	// docs/database-schema.md mensyaratkan ON DELETE CASCADE.
	BundleImages []BundleImage `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}

func (b *Bundle) BeforeCreate(_ *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
