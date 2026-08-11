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

		&entities.Merchandise{},
		&entities.MerchImage{},
		&entities.Bundle{},
		&entities.BundleImage{},

		&entities.Ticket{},
		&entities.TicketTier{},
		&entities.Order{},
		&entities.AttendeeTicket{},
	)
	
	if err != nil {
		return err
	}

	return nil
}