package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendeeTicket struct {
	ID            uuid.UUID  	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OrderID       uuid.UUID  	`gorm:"type:uuid;not null;index"`
	CheckedBy     *uuid.UUID 	`gorm:"type:uuid;index"`
	TicketCode    string     	`gorm:"size:255;uniqueIndex;not null"`
	AttendeeName  string     	`gorm:"size:255;not null"`
	AttendeeEmail string     	`gorm:"size:255;not null"`
	AttendeePhone *string    	`gorm:"size:20"`
	AudienceType  string     	`gorm:"size:50;not null"`
	Institution   *string    	`gorm:"size:255"`
	IsSent        bool       	`gorm:"default:false"`
	SentAt        *time.Time 	`gorm:"type:timestamp with time zone"`
	IsUsed        bool       	`gorm:"default:false"`
	UsedAt        *time.Time 	`gorm:"type:timestamp with time zone"`
	Order         Order      	`gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CheckedByUser *User      	`gorm:"foreignKey:CheckedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	CreatedAt time.Time `gorm:"type:timestamp with time zone"`
}

func (a *AttendeeTicket) BeforeCreate(_ *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
