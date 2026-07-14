package database

import (
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
		&entities.RefreshToken{},
		&entities.Todo{},
	)
	
	if err != nil {
		return err
	}

	return nil
}