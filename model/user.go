package model

import (
	"gorm.io/gorm"
)

// User struct
type User struct {
	gorm.Model
	Name          string `gorm:"not null" json:"name"`
	Email         string `gorm:"uniqueIndex;not null" json:"email"`
	Password      string `gorm:"not null" json:"password"`
	EmailVerified bool   `json:"email_verified"`
	BusinessName  string `json:"business_name"`
	TaxOffice     string `json:"tax_office"`
	TaxNumber     string `json:"tax_number"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	LogoURL       string `json:"logo_url"`
}

func (u *User) CheckBillingInfo(db *gorm.DB) bool {
	var billingInfo BillingInfo
	return db.Where("user_id = ?", u.ID).First(&billingInfo).Error == nil
}

// UserPublic func to return public user data
func (u *User) UserPublic(db *gorm.DB) interface{} {
	return struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
		Email         string `json:"email"`
		BusinessName  string `json:"business_name"`
		TaxOffice     string `json:"tax_office"`
		TaxNumber     string `json:"tax_number"`
		Address       string `json:"address"`
		Phone         string `json:"phone"`
		LogoURL       string `json:"logo_url"`
	}{
		ID:            u.ID,
		Name:          u.Name,
		EmailVerified: u.EmailVerified,
		Email:         u.Email,
		BusinessName:  u.BusinessName,
		TaxOffice:     u.TaxOffice,
		TaxNumber:     u.TaxNumber,
		Address:       u.Address,
		Phone:         u.Phone,
		LogoURL:       u.LogoURL,
	}
}
