package relational

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FrameworkType string

const (
	Bootstrap FrameworkType = "bootstrap"
	Tailwind  FrameworkType = "tailwind"
)

type Template struct {
	gorm.Model
	Name        string         `json:"name"`
	Content     string         `json:"content"`
	Framework   FrameworkType  `json:"framework"`
	Variables   datatypes.JSON `json:"variables" gorm:"type:json"`
	Fonts       MultiString    `json:"fonts" gorm:"type:text"`
	Logs        []Log          `gorm:"constraint:OnDelete:SET NULL;foreignKey:TemplateID"`
	NamespaceID uint
}
