package marketplace

import (
	"designmypdf/pkg/entities"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(category string, excludeAuthorID *uint) ([]*entities.Template, error) {
	var templates []*entities.Template
	q := r.db.Joins("JOIN namespaces ON namespaces.id = templates.namespace_id").
		Where("templates.is_marketplace = ? AND templates.is_published = ?", true, true)
	if category != "" {
		q = q.Where("templates.category = ?", category)
	}
	if excludeAuthorID != nil {
		q = q.Where("namespaces.user_id <> ?", *excludeAuthorID)
	}
	q = q.Preload("Namespace").Preload("Namespace.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "user_name")
	})
	// Do not load large columns for catalog listing.
	if err := q.Omit("content", "variables").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *Repository) GetByID(id uint) (*entities.Template, error) {
	var template entities.Template
	if err := r.db.
		Preload("Namespace").
		Preload("Namespace.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name")
		}).
		Where("id = ? AND is_marketplace = ?", id, true).
		First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *Repository) GetUserListings(userID uint) ([]*entities.Template, error) {
	var templates []*entities.Template
	if err := r.db.Joins("JOIN namespaces ON namespaces.id = templates.namespace_id").
		Where("namespaces.user_id = ? AND templates.is_marketplace = ?", userID, true).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *Repository) GetWithNamespace(id uint) (*entities.Template, error) {
	var template entities.Template
	if err := r.db.Preload("Logs").First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *Repository) IncrementUses(templateID uint) error {
	return r.db.Model(&entities.Template{}).Where("id = ?", templateID).
		UpdateColumn("uses_count", gorm.Expr("uses_count + ?", 1)).Error
}

func (r *Repository) Save(template *entities.Template) error {
	return r.db.Save(template).Error
}

func (r *Repository) Create(template *entities.Template) error {
	return r.db.Create(template).Error
}
