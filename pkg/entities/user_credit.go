package entities

import "gorm.io/gorm"

type UserCredit struct {
	gorm.Model
	UserID       uint   `gorm:"not null;uniqueIndex:idx_user_month"`
	Month        string `gorm:"not null;uniqueIndex:idx_user_month"` // "2006-01"
	CreditsUsed  int    `gorm:"default:0"`
	CreditsLimit int    `gorm:"default:1000000"`
}
