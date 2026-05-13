package entities

import "gorm.io/gorm"

type AiGenerationUsage struct {
	gorm.Model
	UserID         uint   `gorm:"not null;index"`
	Date           string `gorm:"not null;index"` // "2006-01-02"
	WithImageCount int    `gorm:"default:0"`
	TextOnlyCount  int    `gorm:"default:0"`
}
