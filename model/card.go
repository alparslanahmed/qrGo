package model

import (
	"gorm.io/gorm"
)

type Card struct {
	gorm.Model
	UserID         uint   `gorm:"not null"`
	MaskedNumber   string `gorm:"size:20"`
	Brand          string `gorm:"size:20"`
	ExpiryDate     string `gorm:"size:7"`
	CardholderName string `gorm:"size:100"`
	IsDefault      bool   `gorm:"default:false"`
	Token          string `gorm:"size:100"` // Token from payment processor
}
