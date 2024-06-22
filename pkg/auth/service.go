package auth

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/user"
	"errors"
)

type Service interface {
	Login(email string, password string) (*presenter.LoginResponse, error)
	Register(userName string, email string, password string) (*entities.User, error)
	Logout(token string) error
	Refresh(token string) (uint, error)
	Update(id int, userName string, email string, password string) (*entities.User, error)
}

type service struct {
	repository user.Repository
}

// NewService is used to create a single instance of the service
func NewService(r user.Repository) Service {
	return &service{
		repository: *user.NewRepository(database.DB),
	}
}

// Login implements Service.
func (s *service) Login(email string, password string) (*presenter.LoginResponse, error) {
	user, err := s.repository.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}
	accessToken, err := GenerateAccessToken(user.ID)
	refreshToken, err := GenerateRefreshToken(user.ID)
	loginResponse := &presenter.LoginResponse{
		Data:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Error:        nil,
		Status:       true,
	}
	return loginResponse, nil
}

// Logout implements Service.
func (s *service) Logout(token string) error {
	panic("unimplemented")
}

// Refresh implements Service.
func (s *service) Refresh(token string) (uint, error) {
	isValid, err := ValidateRefreshToken(token)
	if err != nil {
		return 0, err
	}

	if !isValid {
		return 0, errors.New("invalid token")
	}

	claims, err := DecodeRefreshToken(token)
	if err != nil {
		return 0, errors.New("error on Generating refresh token")
	}
	return claims.Content, nil
}

// Register implements Service.
func (s *service) Register(userName string, email string, password string) (*entities.User, error) {
	userFromRepo, _ := s.repository.GetByEmail(email)

	if userFromRepo != nil {
		return nil, errors.New("user already exists")
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := entities.User{
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
func (s *service) Update(id int, userName string, email string, password string) (*entities.User, error) {
	user, err := s.repository.Get(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if userName != "" {
		user.UserName = userName
	}
	if email != "" {
		user.Email = email
	}
	if password != "" {
		hashedPassword, err := HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}
	err = s.repository.Update(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
