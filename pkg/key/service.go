package key

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
)

// Service defines the interface for key-related operations.
type Service interface {
	Create(name string, userID uint, keyCount int) (*entities.Key, error)
	Delete(ID uint) (*entities.Key, error)
	GetUserKeys(userID uint) ([]entities.Key, error)
	Update(ID uint, name string, keyCount int) (*entities.Key, error)
	GetKeyByValue(keyValue string) (*entities.Key, error)
	ValidateKey(keyValue string) (bool, error)
}

type service struct {
	repository Repository
}

// NewService creates a new instance of the key service.
func NewService(r Repository) Service {
	return &service{
		repository: *NewRepository(database.DB),
	}
}

// Create creates a new key with the given name and userID.
func (s *service) Create(name string, userID uint, keyCount int) (*entities.Key, error) {
	key := &entities.Key{
		Name:     name,
		UserID:   userID,
		KeyCount: keyCount,
	}
	if err := s.repository.Create(key); err != nil {
		return nil, err
	}
	return key, nil
}

// Delete deletes the key with the given ID.
func (s *service) Delete(ID uint) (*entities.Key, error) {
	key, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.Delete(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GetUserKeys retrieves all keys for the given userID.
func (s *service) GetUserKeys(userID uint) ([]entities.Key, error) {
	keys, err := s.repository.GetAllUserKeys(userID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Update updates the name of the key with the given ID.
func (s *service) Update(ID uint, name string, keyCount int) (*entities.Key, error) {
	key, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	if name != "" {
		key.Name = name
	}

	if keyCount != 0 {
		key.KeyCount = keyCount
	}

	if err := s.repository.Update(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GetKeyByValue retrieves a key by its value.
func (s *service) GetKeyByValue(keyValue string) (*entities.Key, error) {
	key, err := s.repository.GetKeyByValue(keyValue)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ValidateKey validates if a key is valid.
func (s *service) ValidateKey(keyValue string) (bool, error) {
	key, err := s.repository.GetKeyByValue(keyValue)
	if err != nil {
		return false, err
	}
	if key.KeyCountUsed >= key.KeyCount {
		return false, errors.New("key usage limit reached")
	}
	return true, nil
}
