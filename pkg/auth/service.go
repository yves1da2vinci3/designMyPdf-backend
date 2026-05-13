package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"designmypdf/api/handlers/presenter"
	"designmypdf/config/database"
	"designmypdf/pkg/email"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/user"
	"errors"
	"fmt"
	"strings"

	fbauth "firebase.google.com/go/v4/auth"
)

type Service interface {
	Login(email string, password string) (*presenter.LoginResponse, error)
	LoginWithFirebaseIDToken(ctx context.Context, idToken string) (*presenter.LoginResponse, error)
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
	repository   user.Repository
	firebaseAuth *fbauth.Client
}

// NewService is used to create a single instance of the service
func NewService(_ user.Repository, firebaseAuth *fbauth.Client) Service {
	return &service{
		repository:   *user.NewRepository(database.DB),
		firebaseAuth: firebaseAuth,
	}
}

func (s *service) issueTokensForUser(user *entities.User) (*presenter.LoginResponse, error) {
	accessToken, err := GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &presenter.LoginResponse{
		Data:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Status:       true,
	}, nil
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
	return s.issueTokensForUser(user)
}

// LoginWithFirebaseIDToken verifies a Firebase ID token and returns the same session shape as email login.
func (s *service) LoginWithFirebaseIDToken(ctx context.Context, idToken string) (*presenter.LoginResponse, error) {
	if s.firebaseAuth == nil {
		return nil, errors.New("firebase authentication is not configured")
	}
	if strings.TrimSpace(idToken) == "" {
		return nil, errors.New("idToken is required")
	}
	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid firebase token: %w", err)
	}
	uid := strings.TrimSpace(token.UID)
	if uid == "" {
		return nil, errors.New("invalid token: missing uid")
	}
	emailRaw, _ := token.Claims["email"].(string)
	emailAddr := strings.TrimSpace(strings.ToLower(emailRaw))
	if emailAddr == "" {
		return nil, errors.New("email is required from identity provider")
	}
	verified, ok := token.Claims["email_verified"].(bool)
	if !ok || !verified {
		return nil, errors.New("email must be verified")
	}
	displayName, _ := token.Claims["name"].(string)
	displayName = strings.TrimSpace(displayName)

	byFirebase, err := s.repository.GetByFirebaseUID(uid)
	if err != nil {
		return nil, err
	}
	if byFirebase != nil {
		return s.issueTokensForUser(byFirebase)
	}

	byEmail, err := s.repository.GetByEmailOrNil(emailAddr)
	if err != nil {
		return nil, err
	}
	if byEmail != nil {
		byEmail.FirebaseUID = stringPtr(uid)
		if err := s.repository.Update(byEmail); err != nil {
			return nil, err
		}
		return s.issueTokensForUser(byEmail)
	}

	userName := displayName
	if userName == "" {
		userName = strings.Split(emailAddr, "@")[0]
	}
	randomPW, err := randomPasswordHex()
	if err != nil {
		return nil, err
	}
	hashed, err := HashPassword(randomPW)
	if err != nil {
		return nil, err
	}
	u := entities.User{
		UserName:    userName,
		Email:       emailAddr,
		Password:    hashed,
		FirebaseUID: stringPtr(uid),
	}
	if err := s.repository.Create(&u); err != nil {
		return nil, err
	}
	return s.issueTokensForUser(&u)
}

func stringPtr(s string) *string {
	return &s
}

func randomPasswordHex() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
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
	if err != nil || session == nil {
		newSession := &entities.Session{
			UserID:       userID,
			RefreshToken: refreshToken,
		}
		if err := s.repository.CreateSession(newSession); err != nil {
			return errors.New("error creating session")
		}
		return nil
	}
	session.RefreshToken = refreshToken
	if err := s.repository.UpdateSession(session); err != nil {
		return errors.New("error updating session")
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
