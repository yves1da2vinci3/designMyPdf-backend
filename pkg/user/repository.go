package user

import (
	"designmypdf/pkg/entities"

	"gorm.io/gorm"
)

// Repository is a GORM implementation of UserRepository
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) Get(id int) (*entities.User, error) {
	var user entities.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) Delete(user *entities.User) error {
	return r.db.Delete(user).Error
}

func (r *Repository) GetAll() ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByUserName(userName string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByUserNameAndPassword(userName string, password string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("user_name = ? AND password = ?", userName, password).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
