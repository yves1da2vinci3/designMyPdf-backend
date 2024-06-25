package logs

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
	"errors"
	"time"
)

type Service interface {
	CreateLog(log *entities.Log) error
	GetLogsByTimeRange(start, end time.Time) (*[]entities.Log, error)
	GetLogsByUserID(userID uint) (*[]entities.Log, error)
	GetLogStats(userID uint, period string) (*[]LogStats, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: *NewRepository(database.DB)}
}

func (s *service) CreateLog(log *entities.Log) error {
	return s.repo.CreateLog(log)
}

func (s *service) GetLogsByTimeRange(start, end time.Time) (*[]entities.Log, error) {
	return s.repo.GetLogsByTimeRange(start, end)
}

func (s *service) GetLogsByUserID(userID uint) (*[]entities.Log, error) {
	logs, err := s.repo.GetLogsByUserID(userID)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		return nil, errors.New("received nil logs from repository")
	}
	return logs, nil
}

func (s *service) GetLogStats(userID uint, period string) (*[]LogStats, error) {
	logStats, err := s.repo.GetLogStats(userID, period)
	if err != nil {
		return nil, err
	}
	if logStats == nil {
		return nil, errors.New("received nil logStats from repository")
	}
	return logStats, nil
}
