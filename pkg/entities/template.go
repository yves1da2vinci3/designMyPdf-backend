package entities

import (
	"github.com/google/uuid"
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
	UUID        string         `json:"uuid" gorm:"type:uuid;index;unique"`
	Name        string         `json:"name"`
	Content     string         `json:"content"`
	Framework   FrameworkType  `json:"framework"`
	Variables   datatypes.JSON `json:"variables" gorm:"type:json"`
	Fonts       MultiString    `json:"fonts"`
	Logs        []Log          `gorm:"constraint:OnDelete:SET NULL;foreignKey:TemplateID"`
	NamespaceID uint
}

func (template *Template) BeforeCreate(tx *gorm.DB) (err error) {
	template.UUID = uuid.New().String()
	return
}
