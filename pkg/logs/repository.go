package logs

import (
	"designmypdf/pkg/entities"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository is a GORM implementation of LogRepository
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateLog(log *entities.Log) error {
	return r.db.Create(log).Error
}

func (r *Repository) GetLogsByTimeRange(start, end time.Time) (*[]entities.Log, error) {
	var logs []entities.Log
	err := r.db.Where("called_at BETWEEN ? AND ?", start, end).Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return &logs, nil
}

func (r *Repository) GetLogsByUserID(userID uint) (*[]entities.Log, error) {
	var logs []entities.Log
	err := r.db.
		Model(&entities.Log{}).
		Preload("Template", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Preload("Key", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, value")
		}).
		Joins("JOIN keys ON keys.id = logs.key_id").
		Where("keys.user_id = ?", userID).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return &logs, nil
}

type LogStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (r *Repository) GetLogStats(userID uint, period string) (*[]LogStats, error) {
	var startDate time.Time
	var groupBy string

	switch period {
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
		groupBy = "TO_CHAR(called_at, 'DY')"
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
		groupBy = "TO_CHAR(called_at, 'YYYY-MM-DD')"
	case "3months":
		startDate = time.Now().AddDate(0, -3, 0)
		groupBy = "TO_CHAR(called_at, 'YYYY-MM')"
	case "6months":
		startDate = time.Now().AddDate(0, -6, 0)
		groupBy = "TO_CHAR(called_at, 'YYYY-MM')"
	case "1year":
		startDate = time.Now().AddDate(-1, 0, 0)
		groupBy = "TO_CHAR(called_at, 'YYYY-MM')"
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	var stats []LogStats
	err := r.db.
		Table("logs").
		Select("COUNT(*) as count, "+groupBy+" as date").
		Joins("JOIN keys ON keys.id = logs.key_id").
		Where("keys.user_id = ? AND logs.called_at >= ?", userID, startDate).
		Group(groupBy).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
