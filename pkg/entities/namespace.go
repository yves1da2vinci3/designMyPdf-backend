package entities

import (
	"fmt"

	"gorm.io/gorm"
)

type Namespace struct {
	gorm.Model
	Name      string     `json:"name"`
	Templates []Template `json:"templates" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID    uint       `json:"user_id"`
}

func (ns *Namespace) BeforeDelete(tx *gorm.DB) error {
	if ns.ID == 0 {
		return nil
	}
	if err := tx.Model(ns).Association("Templates").Find(&ns.Templates); err != nil {
		return fmt.Errorf("failed to find templates association for namespace  %v", err)
	}
	for _, template := range ns.Templates {
		tx.Delete(&template)
	}
	return nil
}
