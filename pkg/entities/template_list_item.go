package entities

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TemplateListItem is a row for dashboard template cards (includes content for MiniPreview).
type TemplateListItem struct {
	gorm.Model
	UUID               string      `json:"uuid"`
	Name               string      `json:"name"`
	Content            string      `json:"content"`
	Variables          datatypes.JSON `json:"variables"`
	Framework          FrameworkType `json:"framework"`
	Fonts              MultiString `json:"fonts"`
	NamespaceID        uint        `json:"NamespaceID"`
	Description        string      `json:"description"`
	Price              int         `json:"price"`
	IsMarketplace      bool        `json:"is_marketplace"`
	IsPublished        bool        `json:"is_published"`
	Category           string      `json:"category"`
	UsesCount          int         `json:"uses_count"`
	PdfBackgroundColor string      `json:"pdf_background_color"`
	PdfContentPadding  string      `json:"pdf_content_padding"`
}
