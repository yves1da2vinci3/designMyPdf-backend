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
func (r *Repository) GetAllUserNamespaces(userID uint) (*[]entities.Namespace, error) {
	var namespaces *[]entities.Namespace
	if err := r.db.Where("user_id = ?", userID).Preload("Templates").Find(&namespaces).Error; err != nil {
		return nil, err
	}
	return namespaces, nil
}
