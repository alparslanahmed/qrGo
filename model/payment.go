package model

import "gorm.io/gorm"

// In model/payment.go
type Payment struct {
	gorm.Model
	UserID uint
	Amount float64
	Status string
	CardID uint
}
