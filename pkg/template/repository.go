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
