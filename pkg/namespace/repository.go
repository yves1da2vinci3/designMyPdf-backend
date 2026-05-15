package namespace

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

func (r *Repository) Create(namespace *entities.Namespace) error {
	return r.db.Create(namespace).Error
}

func (r *Repository) Get(id uint) (*entities.Namespace, error) {
	var namespace entities.Namespace
	if err := r.db.First(&namespace, id).Error; err != nil {
		return nil, err
	}
	return &namespace, nil
}

func (r *Repository) Update(namespace *entities.Namespace) error {
	return r.db.Save(namespace).Error
}

func (r *Repository) Delete(namespace *entities.Namespace) error {
	return r.db.Delete(namespace).Error
}

func (r *Repository) GetAll() ([]*entities.Namespace, error) {
	var namespaces []*entities.Namespace
	if err := r.db.Find(&namespaces).Error; err != nil {
		return nil, err
	}
	return namespaces, nil
}
func (r *Repository) GetAllUserNamespaces(userID uint) (*[]entities.NamespaceListItem, error) {
	var namespaces []entities.NamespaceListItem
	err := r.db.Table("namespaces").
		Select(`namespaces.id, namespaces.created_at, namespaces.updated_at, namespaces.deleted_at,
			namespaces.name, namespaces.user_id,
			COUNT(templates.id) AS template_count`).
		Joins("LEFT JOIN templates ON templates.namespace_id = namespaces.id AND templates.deleted_at IS NULL").
		Where("namespaces.user_id = ? AND namespaces.deleted_at IS NULL", userID).
		Group("namespaces.id, namespaces.created_at, namespaces.updated_at, namespaces.deleted_at, namespaces.name, namespaces.user_id").
		Scan(&namespaces).Error
	if err != nil {
		return nil, err
	}
	return &namespaces, nil
}
