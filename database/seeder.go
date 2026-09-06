package database

import (
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

// SeedCategories membuat kategori bawaan bila belum ada.
// Idempotent: aman dijalankan ulang.
func SeedCategories(db *gorm.DB) error {
	names := []string{"t-shirt", "cap", "sticker", "other"}

	for _, name := range names {
		var category entities.Category
		if err := db.Where("name = ?", name).FirstOrCreate(&category, entities.Category{Name: name}).Error; err != nil {
			return err
		}
	}

	return nil
}
