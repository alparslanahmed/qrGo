package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name   string `gorm:"size:100" json:"name"`
	Slug   string `gorm:"size:100" json:"slug"`
	UserID uint   `gorm:"not null" json:"user_id"`
	Image  string `gorm:"size:255" json:"image"`
}
