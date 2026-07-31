package entities

import (
	"time"

	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         				uuid.UUID 	`gorm:"primaryKey;default:uuid_generate_v4()"`
	Name       				string    	`gorm:"size:150;not null"`
	Email      				string    	`gorm:"size:225;uniqueIndex;not null"`
	Password   				string    	`gorm:"size:225;not null"`
	TelpNumber 				*string   	`gorm:"size:13;index"`
	Role 					string	  	`gorm:"size:50;not null;default:'user'"`
	IsVerified				bool	  	`gorm:"default:false"`
	VerificationCode   		string     	`gorm:"type:varchar(6)"`
	VerificationExpiry 		*time.Time 	`gorm:"type:timestamp with time zone"`

	Todo []Todo

	Timestamp
}

func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	if u.Role == "" {
		u.Role = constants.ENUM_ROLE_USER
	}

	return nil
}