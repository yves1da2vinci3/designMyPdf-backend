package presenter

import (
	"designmypdf/pkg/entities"
	"time"

	"github.com/gofiber/fiber/v2"
)

type KeyResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	Value        string    `json:"value"`
	Name         string    `json:"name"`
	KeyCount     int       `json:"key_count"`
	KeyCountUsed int       `json:"key_count_used"`
	CreateAt     time.Time `json:"created_at"`
}

func KeySuccessResponse(key *entities.Key) *fiber.Map {
	keyData := KeyResponse{
		ID:           key.ID,
		UserID:       key.UserID,
		Value:        key.Value,
		Name:         key.Name,
		KeyCount:     key.KeyCount,
		KeyCountUsed: key.KeyCountUsed,
		CreateAt:     key.CreatedAt,
	}
	return &fiber.Map{
		"status": true,
		"key":    keyData,
		"error":  nil,
	}
}

func KeysSuccessResponse(keys []entities.Key) *fiber.Map {
	keyData := make([]KeyResponse, len(keys))
	for i, key := range keys {
		keyData[i] = KeyResponse{
			ID:           key.ID,
			UserID:       key.UserID,
			Value:        key.Value,
			Name:         key.Name,
			KeyCount:     key.KeyCount,
			KeyCountUsed: key.KeyCountUsed,
			CreateAt:     key.CreatedAt,
		}
	}
	return &fiber.Map{
		"status": true,
		"keys":   keyData,
		"error":  nil,
	}
}

func KeyErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"key":    nil,
		"error":  err.Error(),
	}
}
