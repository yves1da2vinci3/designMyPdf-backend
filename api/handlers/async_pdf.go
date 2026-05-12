package handlers

import (
	"designmypdf/pkg/key"
	"designmypdf/pkg/pdfjob"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// GeneratePdfAsync enqueues a PDF generation job and returns its ID immediately.
// Auth: dmp_KEY header (same as the synchronous route).
func GeneratePdfAsync(jobSvc *pdfjob.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		keyService := key.NewService(key.Repository{})

		keyValue := c.Get("dmp_KEY")
		if keyValue == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "No key provided"})
		}

		keyEntity, err := keyService.GetKeyByValue(keyValue)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid key"})
		}

		if keyEntity.KeyCountUsed >= keyEntity.KeyCount {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"message": "Key usage limit reached"})
		}

		templateID := c.Params("templateId")
		if templateID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "No template provided"})
		}

		format := c.Query("format", "A4")

		job, err := jobSvc.EnqueueJob(keyEntity.ID, templateID, c.Body(), format)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": fmt.Sprintf("failed to enqueue job: %v", err)})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"job_id": job.ID,
			"status": job.Status,
		})
	}
}

// GetJobStatus returns the current status of a PDF generation job.
// Auth: same dmp_KEY that created the job must be provided.
func GetJobStatus(jobSvc *pdfjob.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		keyService := key.NewService(key.Repository{})

		keyValue := c.Get("dmp_KEY")
		if keyValue == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "No key provided"})
		}

		keyEntity, err := keyService.GetKeyByValue(keyValue)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid key"})
		}

		jobID := c.Params("jobId")
		repo := pdfjob.Repository{}
		job, err := repo.GetByID(jobID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Job not found"})
		}

		if job.KeyID != keyEntity.ID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Access denied"})
		}

		return c.JSON(fiber.Map{
			"job_id": job.ID,
			"status": job.Status,
			"path":   job.ResultPath,
			"error":  job.ErrorMessage,
		})
	}
}
