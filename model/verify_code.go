package model

type VerifyCode struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Email string `json:"email"`
	Code  string `json:"code"`
}
