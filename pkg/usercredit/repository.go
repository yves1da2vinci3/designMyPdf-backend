package usercredit

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
)

type Repository struct{}

func (r Repository) GetOrCreate(userID uint, month string) (*entities.UserCredit, error) {
	var uc entities.UserCredit
	result := database.DB.Where("user_id = ? AND month = ?", userID, month).First(&uc)
	if result.Error != nil {
		uc = entities.UserCredit{
			UserID:       userID,
			Month:        month,
			CreditsUsed:  0,
			CreditsLimit: 1_000_000,
		}
	}
	return &uc, nil
}

func (r Repository) Save(uc *entities.UserCredit) error {
	return database.DB.Save(uc).Error
}
