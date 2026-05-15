package template

import (
	"designmypdf/pkg/entities"

	"gorm.io/gorm"
)

// Repository is a GORM implementation of NamespaceRepository
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(template *entities.Template) error {
	return r.db.Create(template).Error
}

func (r *Repository) Get(id uint) (*entities.Template, error) {
	var template entities.Template
	if err := r.db.First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}
func (r *Repository) GetByUUID(uuid string) (*entities.Template, error) {
	var template entities.Template
	if err := r.db.Where("uuid = ?", uuid).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *Repository) Update(template *entities.Template) error {
	return r.db.Save(template).Error
}

func (r *Repository) Delete(template *entities.Template) error {
	return r.db.Delete(template).Error
}

func (r *Repository) GetAll() ([]*entities.Template, error) {
	var templates []*entities.Template
	if err := r.db.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}
func (r *Repository) GetAllUserTemplates(userID uint) (*[]entities.Template, error) {
	var templates []entities.Template
	if err := r.db.Joins("JOIN namespaces ON namespaces.id = templates.namespace_id").
		Where("namespaces.user_id = ?", userID).Find(&templates).Error; err != nil {
		return nil, err
	}
	return &templates, nil
}

type ListUserTemplatesFilter struct {
	UserID      uint
	NamespaceID *uint
	Query       string
	Offset      int
	Limit       int
}

type ListUserTemplatesResult struct {
	Items []entities.TemplateListItem
	Total int64
}

func (r *Repository) ListUserTemplates(f ListUserTemplatesFilter) (*ListUserTemplatesResult, error) {
	base := r.db.Model(&entities.Template{}).
		Joins("JOIN namespaces ON namespaces.id = templates.namespace_id").
		Where("namespaces.user_id = ?", f.UserID)

	if f.NamespaceID != nil {
		base = base.Where("templates.namespace_id = ?", *f.NamespaceID)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		base = base.Where("templates.name LIKE ?", like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []entities.TemplateListItem
	q := r.db.Model(&entities.Template{}).
		Select(`templates.id, templates.created_at, templates.updated_at, templates.deleted_at,
			templates.uuid, templates.name, templates.content, templates.variables,
			templates.framework, templates.fonts,
			templates.namespace_id, templates.description,
			templates.price, templates.is_marketplace, templates.is_published, templates.category,
			templates.uses_count, templates.pdf_background_color, templates.pdf_content_padding`).
		Joins("JOIN namespaces ON namespaces.id = templates.namespace_id").
		Where("namespaces.user_id = ?", f.UserID)

	if f.NamespaceID != nil {
		q = q.Where("templates.namespace_id = ?", *f.NamespaceID)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("templates.name LIKE ?", like)
	}

	if err := q.Order("templates.updated_at DESC").
		Offset(f.Offset).
		Limit(f.Limit).
		Scan(&items).Error; err != nil {
		return nil, err
	}

	return &ListUserTemplatesResult{Items: items, Total: total}, nil
}

func (r *Repository) UpdateFields(id uint, fields map[string]interface{}) (*entities.Template, error) {
	var template entities.Template
	if err := r.db.First(&template, id).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&template).Updates(fields).Error; err != nil {
		return nil, err
	}
	return &template, nil
}
