package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BundleImage struct {
	ID        uuid.UUID 	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	BundleID  uuid.UUID 	`gorm:"type:uuid;not null;index"`
	ImageURL  string    	`gorm:"size:255;not null"`
	Bundle    Bundle    	`gorm:"foreignKey:BundleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (b *BundleImage) BeforeCreate(_ *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
