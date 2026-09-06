package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name string    `gorm:"size:50;uniqueIndex;not null"`

	Timestamp
}

func (Category) TableName() string {
	return "categories"
}

func (c *Category) BeforeCreate(_ *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
