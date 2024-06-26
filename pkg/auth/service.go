package auth

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/config/database"
	"designmypdf/pkg/email"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/user"
	"errors"
	"fmt"
)

type Service interface {
	Login(email string, password string) (*presenter.LoginResponse, error)
	Register(userName string, email string, password string) (*entities.User, error)
	Logout(sessionID uint) error
	Refresh(RefreshToken string) (string, error)
	Update(id float64, userName string, password string) (*entities.User, error)
	SetSession(userID uint, refreshToken string) error
	GetSessionByToken(token string) (*entities.Session, error)
	ForgotPassword(mail string) error
	ResetPassword(token, password string) error
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
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

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
func (s *service) Logout(sessionID uint) error {

	err := s.repository.DeleteSession(sessionID)
	if err != nil {
		return err
	}
	return nil
}

// Refresh implements Service.
func (s *service) Refresh(RefreshToken string) (string, error) {

	claims, err := DecodeRefreshToken(RefreshToken)
	if err != nil {
		return "", errors.New("error decoding refresh token")
	}

	newAccessToken, err := GenerateAccessToken(claims.Content)
	if err != nil {
		return "", errors.New("error generating new access token")
	}
	return newAccessToken, nil
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
func (s *service) Update(id float64, userName string, password string) (*entities.User, error) {
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

// ForgotPassword implements Service.
func (s *service) ForgotPassword(mail string) error {
	_, err := s.repository.GetByEmail(mail)
	if err != nil {
		return errors.New("email not found")
	}

	token, err := GenerateResetToken(mail)
	if err != nil {
		return errors.New("error generating token")
	}
	err = email.SendForgotPasswordEmail(mail, token)
	if err != nil {
		return errors.New("error sending email")
	}
	return nil
}

// SetSession implements Service.
func (s *service) SetSession(userID uint, refreshToken string) error {
	session, err := s.repository.FindSessionByUserID(userID)
	if err != nil {
		newSession := &entities.Session{
			UserID:       userID,
			RefreshToken: refreshToken,
		}
		err = s.repository.CreateSession(newSession)
		if err != nil {
			return errors.New("error creating session")
		}
	}
	if session != nil {
		err = s.repository.DeleteSession(session.ID)
		if err != nil {
			return errors.New("error deleting session")
		}
	}

	return nil
}

// ResetPassword implements Service.
func (s *service) ResetPassword(token, password string) error {
	claims, err := VerifyResetToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}
	fmt.Printf("Claim %v", claims)
	user, err := s.repository.GetByEmail(claims.Email)
	if err != nil {
		return errors.New("user not found")
	}

	// Update the user's password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return errors.New("error hashing password")
	}

	user.Password = hashedPassword
	err = s.repository.Update(user)
	if err != nil {
		return errors.New("error updating password")
	}
	return nil
}

func (s *service) GetSessionByToken(token string) (*entities.Session, error) {
	session, err := s.repository.FindSessionByToken(token)
	if err != nil {
		return nil, errors.New("error finding session")
	}
	if session == nil {
		return nil, errors.New("session not found")
	}
	return session, nil
}
