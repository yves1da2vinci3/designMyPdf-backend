package user

import (
	"designmypdf/pkg/enities/relational"

	"gorm.io/gorm"
)

// GORM implementation of UserRepository
type gormUserRepository struct {
	db *gorm.DB
}

func (r *gormUserRepository) Create(user interface{}) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) Get(id interface{}) (interface{}, error) {
	var user relational.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Update(user interface{}) error {
	return r.db.Save(user).Error
}

func (r *gormUserRepository) Delete(user interface{}) error {
	return r.db.Delete(user).Error
}

func (r *gormUserRepository) GetAll() ([]interface{}, error) {
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

func (r *gormUserRepository) GetByEmail(email string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) GetByUserName(userName string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) GetByUserNameAndPassword(userName string, password string) (interface{}, error) {
	var user relational.User
	if err := r.db.Where("user_name = ? AND password = ?", userName, password).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
