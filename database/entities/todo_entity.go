package entities

import "github.com/google/uuid"

type Todo struct {
	ID 							uuid.UUID						`gorm:"primaryKey;default:uuid_generate_v4()"`
	UserID					uuid.UUID						`gorm:"not null;index"`
	Name						string							`gorm:"size:100;not null"`
	Category 				string							`gorm:"size:100;not null"`
	IsDone					bool								`gorm:"default:false"`

	User						User								`gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`

	Timestamp
}