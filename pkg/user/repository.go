package user

import (
	"designmypdf/pkg/entities"
	"time"

	"gorm.io/gorm"
)

// Repository is a GORM implementation of UserRepository
type Repository struct {
	db *gorm.DB
}

type Session struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   // Example: User ID associated with the session
	Token     string // Example: Session token
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) Get(id float64) (*entities.User, error) {
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

// Create inserts a new session record into the database
func (r *Repository) CreateSession(session *entities.Session) error {
	return r.db.Create(session).Error
}

// FindByID retrieves a session record by its ID from the database
func (r *Repository) FindSessionByID(id uint) (*entities.Session, error) {
	var session entities.Session
	if err := r.db.First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByID retrieves a session record by its ID from the database
func (r *Repository) FindSessionByUserID(userID uint) (*entities.Session, error) {
	var session entities.Session
	if err := r.db.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByID retrieves a session record by its ID from the database
func (r *Repository) FindSessionByToken(refreshToken string) (*entities.Session, error) {
	var session entities.Session
	if err := r.db.Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// Delete removes a session record from the database by its ID
func (r *Repository) DeleteSession(id uint) error {
	return r.db.Delete(&Session{}, id).Error
}
