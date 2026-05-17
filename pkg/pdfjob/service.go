package pdfjob

import (
	"context"
	"designmypdf/pkg/amqp"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/template"
	"designmypdf/pkg/webhook"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo       Repository
	amqpClient *amqp.Client
}

func NewService(amqpClient *amqp.Client) *Service {
	return &Service{repo: Repository{}, amqpClient: amqpClient}
}

// EnqueueJob persists a new job in queued state and publishes it to RabbitMQ.
func (s *Service) EnqueueJob(keyID uint, templateUUID string, payload []byte, format string) (*entities.PdfGenerationJob, error) {
	job := &entities.PdfGenerationJob{
		ID:           uuid.New().String(),
		KeyID:        keyID,
		TemplateUUID: templateUUID,
		Payload:      payload,
		Format:       format,
		Status:       entities.JobStatusQueued,
	}

	if err := s.repo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	if err := s.amqpClient.Publish(job.ID); err != nil {
		// Job is in DB — worker can be retried manually; log and continue.
		fmt.Printf("warning: failed to publish job %s to queue: %v\n", job.ID, err)
	}

	return job, nil
}

// ProcessJob is called by the worker for each message consumed from RabbitMQ.
func (s *Service) ProcessJob(jobID string) error {
	job, err := s.repo.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("job %s not found: %w", jobID, err)
	}

	if err := s.repo.UpdateStatus(jobID, entities.JobStatusRunning, "", ""); err != nil {
		fmt.Printf("warning: failed to mark job %s running: %v\n", jobID, err)
	}

	templateSvc := template.NewService(template.Repository{})
	templateEntity, err := templateSvc.GetByUUID(job.TemplateUUID)
	if err != nil {
		return s.failJob(job, nil, fmt.Sprintf("template not found: %v", err))
	}

	var data map[string]interface{}
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, &data); err != nil {
			return s.failJob(job, templateEntity, fmt.Sprintf("invalid payload: %v", err))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pdfURL, err := GeneratePdfForKey(ctx, &job.Key, templateEntity, data, job.Format)
	if err != nil {
		return s.failJob(job, templateEntity, err.Error())
	}

	if err := s.repo.UpdateStatus(jobID, entities.JobStatusCompleted, pdfURL, ""); err != nil {
		fmt.Printf("warning: failed to mark job %s completed: %v\n", jobID, err)
	}

	go logs.RecordPdfGeneration(
		job.KeyID,
		templateEntity.ID,
		job.TemplateUUID,
		job.Payload,
		map[string]interface{}{
			"path":   pdfURL,
			"job_id": job.ID,
		},
		entities.Success,
		nil,
	)

	publisher := webhook.NewPublisher()
	publisher.Publish(webhook.EventPdfJobCompleted, jobID, job.Key.UserID, job.KeyID, map[string]interface{}{
		"path":          pdfURL,
		"template_uuid": job.TemplateUUID,
	})

	return nil
}

func (s *Service) failJob(job *entities.PdfGenerationJob, templateEntity *entities.Template, errMsg string) error {
	if err := s.repo.UpdateStatus(job.ID, entities.JobStatusFailed, "", errMsg); err != nil {
		fmt.Printf("warning: failed to mark job %s failed: %v\n", job.ID, err)
	}

	templateID := uint(0)
	if templateEntity != nil {
		templateID = templateEntity.ID
	}
	go logs.RecordPdfGeneration(
		job.KeyID,
		templateID,
		job.TemplateUUID,
		job.Payload,
		map[string]interface{}{
			"job_id":  job.ID,
			"message": errMsg,
		},
		entities.Fail,
		errors.New(errMsg),
	)

	publisher := webhook.NewPublisher()
	publisher.Publish(webhook.EventPdfJobFailed, job.ID, job.Key.UserID, job.KeyID, map[string]interface{}{
		"error":         errMsg,
		"template_uuid": job.TemplateUUID,
	})

	return fmt.Errorf("job %s failed: %s", job.ID, errMsg)
}
