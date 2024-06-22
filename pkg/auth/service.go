package auth

import "designmypdf/pkg/user"

type Service interface {
	Login(email string, password string) (string, error)
	Register(userName string, email string, password string) (string, error)
	Logout(token string) error
	Refresh(token string) (string, error)
	Update(userName string, email string, password string) (string, error)
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
func (s *service) Login(email string, password string) (string, error) {
	panic("unimplemented")
}

// Logout implements Service.
func (s *service) Logout(token string) error {
	panic("unimplemented")
}

// Refresh implements Service.
func (s *service) Refresh(token string) (string, error) {
	panic("unimplemented")
}

// Register implements Service.
func (s *service) Register(userName string, email string, password string) (string, error) {
	panic("unimplemented")
}

// Update implements Service.
func (s *service) Update(userName string, email string, password string) (string, error) {
	panic("unimplemented")
}
