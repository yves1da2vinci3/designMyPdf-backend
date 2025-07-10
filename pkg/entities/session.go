package entities

import "gorm.io/gorm"

type Session struct {
	gorm.Model
	RefreshToken string `json:"refresh_token"`
	UserID       uint   `json:"user_id"`
}
