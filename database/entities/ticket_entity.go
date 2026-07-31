package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text;not null"`
	IsActive    bool      `gorm:"default:true"`

	Timestamp
}

func (t *Ticket) BeforeCreate(_ *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
