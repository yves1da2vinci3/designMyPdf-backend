package auth

import (
	"designmypdf/pkg/user"
	"errors"
)

//	type User struct {
//		ID       string `json:"id"omitempty"`
//		UserName string
//		Email    string
//		Password string
//	}
type User struct {
	UserName string
	Email    string
	Password string
}

type Service interface {
	Login(email string, password string) (interface{}, error)
	Register(userName string, email string, password string) (interface{}, error)
	Logout(token string) error
	Refresh(token string) (string, error)
	Update(id string, userName string, email string, password string) (interface{}, error)
}

type service struct {
	repository user.UserRepository
}

// NewService is used to create a single instance of the service
func NewService(r user.UserRepository) Service {
	return &service{
		repository: user.NewUserRepository(),
	}
}

// Login implements Service.
func (s *service) Login(email string, password string) (interface{}, error) {
	userFromRepo, err := s.repository.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	user := userFromRepo.(*User)
	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

// Logout implements Service.
func (s *service) Logout(token string) error {
	panic("unimplemented")
}

// Refresh implements Service.
func (s *service) Refresh(token string) (string, error) {
	isValid, err := ValidateRefreshToken(token)

	if !isValid {
		return "", errors.New("invalid token")
	}

	claims, err := DecodeRefreshToken(token)
	if err != nil {
		return "", errors.New("error on Generating refresh token")
	}
	return claims.Content, nil

}

// Register implements Service.
func (s *service) Register(userName string, email string, password string) (interface{}, error) {
	userFromRepo, _ := s.repository.GetByEmail(email)

	if userFromRepo != nil {
		return nil, errors.New("user already exists")
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := User{
		UserName: userName,
		Email:    email,
		Password: hashedPassword,
	}
	err = s.repository.Create(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update implements Service.
func (s *service) Update(id string, string, email string, password string) (interface{}, error) {
	userFromRepo, err := s.repository.Get(id)
	if err != nil {
		return "", err
	}
	if userFromRepo == nil {
		return "", errors.New("user not found")
	}
	user := userFromRepo.(*User)
	if email != "" {
		user.Email = email
	}
	if password != "" {
		hashedPassword, err := HashPassword(password)
		if err != nil {
			return "", err
		}
		user.Password = hashedPassword
	}
	err = s.repository.Update(user)
	if err != nil {
		return "", err
	}
	return user, nil
}
