package user

import (
	"designmypdf/pkg/enities/relational"

	"gorm.io/gorm"
)

// GORM implementation of UserRepository
type gormRepository struct {
	db *gorm.DB
}

func (r *gormRepository) Create(user interface{}) error {
	return r.db.Create(user).Error
}

func (r *gormRepository) Get(id interface{}) (interface{}, error) {
	var user relational.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) Update(user interface{}) error {
	return r.db.Save(user).Error
}

func (r *gormRepository) Delete(user interface{}) error {
	return r.db.Delete(user).Error
}

func (r *gormRepository) GetAll() ([]interface{}, error) {
	var users []relational.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]interface{}, len(users))
	for i, user := range users {
		result[i] = user
	}
	return result, nil
}

func (r *gormRepository) GetByEmail(email string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) GetByUserName(userName string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) GetByUserNameAndPassword(userName string, password string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("user_name = ? AND password = ?", userName, password).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
