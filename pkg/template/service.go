package template

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/namespace"

	"gorm.io/datatypes"
)

// Service defines the interface for template-related operations.
type Service interface {
	Create(name string, content string, variables datatypes.JSON, fonts entities.MultiString, namespaceID uint) (*entities.Template, error)
	Delete(ID uint) (*entities.Template, error)
	GetUserTemplates(userID uint) (*[]entities.Template, error)
	Get(ID uint) (*entities.Template, error)
	GetByUUID(UUID string) (*entities.Template, error)
	Update(ID uint, name string, content string, variables datatypes.JSON, fonts entities.MultiString) (*entities.Template, error)
	ChangeTemplateNamespace(ID uint, NamespaceID uint) error
}

type service struct {
	repository Repository
}

// NewService creates a new instance of the template service.
func NewService(r Repository) Service {
	return &service{
		repository: *NewRepository(database.DB),
	}
}

// Create creates a new template with the given name and userID.
func (s *service) Create(name string, content string, variables datatypes.JSON, fonts entities.MultiString, namespaceID uint) (*entities.Template, error) {
	template := &entities.Template{
		Name:        name,
		Content:     content,
		NamespaceID: namespaceID,
		Variables:   variables,
		Fonts:       fonts,
		Framework:   entities.Tailwind,
	}
	if err := s.repository.Create(template); err != nil {
		return nil, err
	}
	return template, nil
}

// Delete deletes the template with the given ID.
func (s *service) Delete(ID uint) (*entities.Template, error) {
	template, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.Delete(template); err != nil {
		return nil, err
	}
	return template, nil
}

// GetUserTemplates retrieves all Templates for the given userID.
func (s *service) GetUserTemplates(userID uint) (*[]entities.Template, error) {
	Templates, err := s.repository.GetAllUserTemplates(userID)
	if err != nil {
		return nil, err
	}
	return Templates, nil
}

// Update updates the name of the template with the given ID.
func (s *service) Update(ID uint, name string, content string, variables datatypes.JSON, fonts entities.MultiString) (*entities.Template, error) {
	template, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}
	template.Name = name
	template.Content = content
	template.Variables = variables
	template.Fonts = fonts

	if err := s.repository.Update(template); err != nil {
		return nil, err
	}
	return template, nil
}

// Update updates the name of the template with the given ID.
func (s *service) ChangeTemplateNamespace(ID uint, NamespaceID uint) error {
	template, err := s.repository.Get(ID)
	if err != nil {
		return err
	}
	namespaceRepo := namespace.NewRepository(database.DB)
	_, err = namespaceRepo.Get(NamespaceID)
	if err != nil {
		return err
	}
	template.NamespaceID = NamespaceID

	if err := s.repository.Update(template); err != nil {
		return nil
	}
	return nil
}

// Update the name of the template with the given ID.
func (s *service) Get(ID uint) (*entities.Template, error) {
	template, err := s.repository.Get(ID)
	if err != nil {
		return nil, err
	}

	return template, nil
}

// Get By Uid updates the name of the template with the given ID.
func (s *service) GetByUUID(UUID string) (*entities.Template, error) {
	template, err := s.repository.GetByUUID(UUID)
	if err != nil {
		return nil, err
	}

	return template, nil
}
