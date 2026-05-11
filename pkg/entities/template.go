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
	UUID          string         `json:"uuid" gorm:"type:uuid;index;unique"`
	Name          string         `json:"name"`
	Content       string         `json:"content"`
	Framework     FrameworkType  `json:"framework"`
	Variables     datatypes.JSON `json:"variables" gorm:"type:json"`
	Fonts         MultiString    `json:"fonts"`
	Logs          []Log          `gorm:"constraint:OnDelete:SET NULL;foreignKey:TemplateID"`
	NamespaceID   uint
	Description   string         `json:"description"`
	CoverImageURL string         `json:"cover_image_url"`
	Price         int            `json:"price"`
	IsMarketplace bool           `json:"is_marketplace" gorm:"default:false"`
	IsPublished   bool           `json:"is_published" gorm:"default:false"`
	Category      string         `json:"category"`
	Features      MultiString    `json:"features"`
	UsesCount     int            `json:"uses_count" gorm:"default:0"`
}

func (template *Template) BeforeCreate(tx *gorm.DB) (err error) {
	template.UUID = uuid.New().String()
	return
}
