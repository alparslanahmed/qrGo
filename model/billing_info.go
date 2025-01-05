package model

import (
	"gorm.io/gorm"
)

type BillingInfo struct {
	gorm.Model
	UserID      uint   `gorm:"not null" json:"user_id"`
	FullName    string `gorm:"not null" json:"full_name"`
	Email       string `gorm:"not null" json:"email"`
	Address     string `gorm:"not null" json:"address"`
	City        string `gorm:"not null" json:"city"`
	State       string `gorm:"not null" json:"state"`
	Country     string `gorm:"not null" json:"country"`
	ZipCode     string `gorm:"not null" json:"zip_code"`
	TaxID       string `gorm:"not null" json:"tax_id"`
	PhoneNumber string `gorm:"not null" json:"phone_number"`
}
