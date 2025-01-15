package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string  `json:"name" gorm:"size:100"`
	Description string  `json:"description" gorm:"size:255"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Slug        string  `json:"slug" gorm:"size:100"`
	UserID      uint    `json:"user_id" gorm:"not null"`
	CategoryID  uint    `json:"category_id" gorm:"not null"`
}
