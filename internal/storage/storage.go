package storage

import (
	"log"

	"test/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLite(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Link{}, &models.LinkSet{}); err != nil {
		log.Println("migrate error:", err)
	}
	return db, err
}
