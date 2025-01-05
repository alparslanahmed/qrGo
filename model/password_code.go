package model

type PasswordCode struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Code   string `json:"code"`
	UserId uint   `json:"user_id"`
}
