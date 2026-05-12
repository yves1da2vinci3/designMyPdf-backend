package pdfjob

import (
	"designmypdf/config/database"
	"designmypdf/pkg/entities"
)

type Repository struct{}

func (r Repository) Create(job *entities.PdfGenerationJob) error {
	return database.DB.Create(job).Error
}

func (r Repository) GetByID(id string) (*entities.PdfGenerationJob, error) {
	var job entities.PdfGenerationJob
	err := database.DB.Preload("Key").First(&job, "id = ?", id).Error
	return &job, err
}

func (r Repository) UpdateStatus(id string, status entities.JobStatus, resultPath, errMsg string) error {
	return database.DB.Model(&entities.PdfGenerationJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"result_path":   resultPath,
			"error_message": errMsg,
		}).Error
}
