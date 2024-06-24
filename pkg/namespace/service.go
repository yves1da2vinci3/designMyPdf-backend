package namespace

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
)

// Service defines the interface for namespace-related operations.
type Service interface {
	Create(name string, userID uint) (*entities.Namespace, error)
	Delete(ID uint) (*entities.Namespace, error)
	GetUserNamespaces(userID uint) (*[]entities.Namespace, error)
	Update(ID uint, name string) (*entities.Namespace, error)
}

type service struct {
	repository Repository
}

// NewService creates a new instance of the namespace service.
func NewService(r Repository) Service {
	return &service{
		repository: *NewRepository(database.DB),
	}
}

// Create creates a new namespace with the given name and userID.
func (s *service) Create(name string, userID uint) (*entities.Namespace, error) {
	ns := &entities.Namespace{
		Name:   name,
		UserID: userID,
	}
	if err := s.repository.Create(ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// Delete deletes the namespace with the given ID.
func (s *service) Delete(ID uint) (*entities.Namespace, error) {
	ns, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.Delete(ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// GetUserNamespaces retrieves all namespaces for the given userID.
func (s *service) GetUserNamespaces(userID uint) (*[]entities.Namespace, error) {
	namespaces, err := s.repository.GetAllUserNamespaces(userID)
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}

// Update updates the name of the namespace with the given ID.
func (s *service) Update(ID uint, name string) (*entities.Namespace, error) {
	ns, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	ns.Name = name
	if err := s.repository.Update(ns); err != nil {
		return nil, err
	}
	return ns, nil
}
