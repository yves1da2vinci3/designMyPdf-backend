package presenter

import (
	"designmypdf/pkg/entities"
	"time"
)

// MarketplaceListItem is a lightweight row for GET /marketplace (no HTML body or variables JSON).
type MarketplaceListItem struct {
	ID              uint                 `json:"ID"`
	UUID            string               `json:"uuid"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Category        string               `json:"category"`
	CoverImageURL   string               `json:"cover_image_url"`
	Price           int                  `json:"price"`
	IsMarketplace   bool                 `json:"is_marketplace"`
	IsPublished     bool                 `json:"is_published"`
	NamespaceID     uint                 `json:"NamespaceID"`
	UsesCount       int                  `json:"uses_count"`
	Features        entities.MultiString `json:"features"`
	Framework       entities.FrameworkType `json:"framework"`
	CreatedAt       *time.Time           `json:"CreatedAt,omitempty"`
	UpdatedAt       *time.Time           `json:"UpdatedAt,omitempty"`
}

// ToMarketplaceListItem maps a template entity to a list card (omits content and variables).
func ToMarketplaceListItem(t *entities.Template) MarketplaceListItem {
	if t == nil {
		return MarketplaceListItem{}
	}
	var created, updated *time.Time
	if !t.CreatedAt.IsZero() {
		created = &t.CreatedAt
	}
	if !t.UpdatedAt.IsZero() {
		updated = &t.UpdatedAt
	}
	return MarketplaceListItem{
		ID:              t.ID,
		UUID:            t.UUID,
		Name:            t.Name,
		Description:     t.Description,
		Category:        t.Category,
		CoverImageURL:   t.CoverImageURL,
		Price:           t.Price,
		IsMarketplace:   t.IsMarketplace,
		IsPublished:     t.IsPublished,
		NamespaceID:     t.NamespaceID,
		UsesCount:       t.UsesCount,
		Features:        t.Features,
		Framework:       t.Framework,
		CreatedAt:       created,
		UpdatedAt:       updated,
	}
}
