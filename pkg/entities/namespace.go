package entities

import "gorm.io/gorm"

type Namespace struct {
	gorm.Model
	Name      string     `json:"name"`
	Templates []Template `json:"templates" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID    uint       `json:"user_id"`
}
