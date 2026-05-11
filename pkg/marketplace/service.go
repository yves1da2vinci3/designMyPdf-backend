package marketplace

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
	"fmt"
	"strings"
)

const (
	minDescriptionLen = 80
)

var allowedCategories = map[string]struct{}{
	"INVOICE":          {},
	"FINANCIAL REPORT": {},
	"MARKETING":        {},
	"LEGAL":            {},
	"OTHER":            {},
}

// ValidateListingMetadata returns an error if marketplace listing fields are insufficient.
func ValidateListingMetadata(name, description, category, coverImageURL string, features entities.MultiString) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if len(strings.TrimSpace(description)) < minDescriptionLen {
		return fmt.Errorf("description must be at least %d characters", minDescriptionLen)
	}
	if _, ok := allowedCategories[strings.TrimSpace(category)]; !ok {
		return errors.New("invalid category")
	}
	if strings.TrimSpace(coverImageURL) == "" {
		return errors.New("coverImageURL is required")
	}
	if len(features) == 0 {
		return errors.New("at least one feature is required")
	}
	if len(features) > 0 {
		nonEmpty := 0
		for _, f := range features {
			if strings.TrimSpace(f) != "" {
				nonEmpty++
			}
		}
		if nonEmpty == 0 {
			return errors.New("at least one feature is required")
		}
	}
	return nil
}

type Service interface {
	GetAll(category string) ([]*entities.Template, error)
	GetByID(id uint) (*entities.Template, error)
	GetUserListings(userID uint) ([]*entities.Template, error)
	Publish(templateID, userID uint, name, description string, price int, category string, features entities.MultiString, coverImageURL string) (*entities.Template, error)
	UpdateListing(templateID, userID uint, name, description string, price int, category string, features entities.MultiString, coverImageURL string, isPublished *bool) (*entities.Template, error)
	SetListingPublished(templateID, userID uint, published bool) (*entities.Template, error)
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

func (s *service) Publish(templateID, userID uint, name, description string, price int, category string, features entities.MultiString, coverImageURL string) (*entities.Template, error) {
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}
	if err := ValidateListingMetadata(name, description, category, coverImageURL, features); err != nil {
		return nil, err
	}

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

	template.Name = strings.TrimSpace(name)
	template.Description = strings.TrimSpace(description)
	template.Price = price
	template.Category = strings.TrimSpace(category)
	template.Features = features
	template.CoverImageURL = strings.TrimSpace(coverImageURL)
	template.IsMarketplace = true
	template.IsPublished = true

	if err := s.repo.Save(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *service) UpdateListing(templateID, userID uint, name, description string, price int, category string, features entities.MultiString, coverImageURL string, isPublished *bool) (*entities.Template, error) {
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}
	if err := ValidateListingMetadata(name, description, category, coverImageURL, features); err != nil {
		return nil, err
	}

	template, err := s.repo.GetWithNamespace(templateID)
	if err != nil {
		return nil, err
	}

	var nsUserID uint
	if err := database.DB.Table("namespaces").Select("user_id").Where("id = ?", template.NamespaceID).Scan(&nsUserID).Error; err != nil {
		return nil, err
	}
	if nsUserID != userID {
		return nil, errors.New("unauthorized: template does not belong to user")
	}
	if !template.IsMarketplace {
		return nil, errors.New("template is not a marketplace listing")
	}

	template.Name = strings.TrimSpace(name)
	template.Description = strings.TrimSpace(description)
	template.Price = price
	template.Category = strings.TrimSpace(category)
	template.Features = features
	template.CoverImageURL = strings.TrimSpace(coverImageURL)
	if isPublished != nil {
		template.IsPublished = *isPublished
	}

	if err := s.repo.Save(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *service) SetListingPublished(templateID, userID uint, published bool) (*entities.Template, error) {
	template, err := s.repo.GetWithNamespace(templateID)
	if err != nil {
		return nil, err
	}
	var nsUserID uint
	if err := database.DB.Table("namespaces").Select("user_id").Where("id = ?", template.NamespaceID).Scan(&nsUserID).Error; err != nil {
		return nil, err
	}
	if nsUserID != userID {
		return nil, errors.New("unauthorized: template does not belong to user")
	}
	if !template.IsMarketplace {
		return nil, errors.New("template is not a marketplace listing")
	}
	template.IsPublished = published
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
