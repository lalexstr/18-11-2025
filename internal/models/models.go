package models

import "time"

type LinkSet struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	Links     []Link
}

type Link struct {
	ID        uint `gorm:"primary_key"`
	LinkSetID uint
	URL       string `gorm:"size:2048"`
	Status    string `gorm:"size:32"` // pending, ok, fail
	Processed bool
	UpdatedAt time.Time
}
