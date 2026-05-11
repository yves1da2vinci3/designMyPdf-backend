package marketplace

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
)

type Service interface {
	GetAll(category string) ([]*entities.Template, error)
	GetByID(id uint) (*entities.Template, error)
	GetUserListings(userID uint) ([]*entities.Template, error)
	Publish(templateID, userID uint, description string, price int, category string, features entities.MultiString, coverImageURL string) (*entities.Template, error)
	CopyToNamespace(templateID, namespaceID, userID uint) (*entities.Template, error)
}

type service struct {
	repo *Repository
}

func NewService() Service {
	return &service{
		repo: NewRepository(database.DB),
	}
}

func (s *service) GetAll(category string) ([]*entities.Template, error) {
	return s.repo.GetAll(category)
}

func (s *service) GetByID(id uint) (*entities.Template, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetUserListings(userID uint) ([]*entities.Template, error) {
	return s.repo.GetUserListings(userID)
}

func (s *service) Publish(templateID, userID uint, description string, price int, category string, features entities.MultiString, coverImageURL string) (*entities.Template, error) {
	template, err := s.repo.GetWithNamespace(templateID)
	if err != nil {
		return nil, err
	}

	// Verify namespace belongs to requesting user via separate namespace query
	var nsUserID uint
	if err := database.DB.Table("namespaces").Select("user_id").Where("id = ?", template.NamespaceID).Scan(&nsUserID).Error; err != nil {
		return nil, err
	}
	if nsUserID != userID {
		return nil, errors.New("unauthorized: template does not belong to user")
	}

	template.Description = description
	template.Price = price
	template.Category = category
	template.Features = features
	template.CoverImageURL = coverImageURL
	template.IsMarketplace = true
	template.IsPublished = true

	if err := s.repo.Save(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *service) CopyToNamespace(templateID, namespaceID, userID uint) (*entities.Template, error) {
	source, err := s.repo.GetByID(templateID)
	if err != nil {
		return nil, err
	}

	// Verify target namespace belongs to requesting user
	var nsUserID uint
	if err := database.DB.Table("namespaces").Select("user_id").Where("id = ?", namespaceID).Scan(&nsUserID).Error; err != nil {
		return nil, err
	}
	if nsUserID != userID {
		return nil, errors.New("unauthorized: namespace does not belong to user")
	}

	copy := &entities.Template{
		Name:        source.Name,
		Content:     source.Content,
		Framework:   source.Framework,
		Variables:   source.Variables,
		Fonts:       source.Fonts,
		NamespaceID: namespaceID,
	}
	if err := s.repo.Create(copy); err != nil {
		return nil, err
	}

	// Increment uses count on source
	_ = s.repo.IncrementUses(templateID)

	return copy, nil
}
