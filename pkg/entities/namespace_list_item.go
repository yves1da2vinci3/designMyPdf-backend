package entities

import "gorm.io/gorm"

// NamespaceListItem is returned by GET /namespaces without embedded templates.
type NamespaceListItem struct {
	gorm.Model
	Name          string `json:"name"`
	UserID        uint   `json:"user_id"`
	TemplateCount int64  `json:"template_count"`
}
