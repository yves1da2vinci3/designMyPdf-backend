package handlers

import (
	"context"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/pdfjob"
	"designmypdf/pkg/template"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

func GeneratePdf(c *fiber.Ctx) error {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	debug.FreeOSMemory()

	keyService := key.NewService(key.Repository{})

	keyValue := c.Get("dmp_KEY")
	if keyValue == "" {
		return logAndRespond(c, nil, nil, "No key provided", fiber.StatusUnauthorized)
	}

	keyEntity, err := keyService.GetKeyByValue(keyValue)
	if err != nil {
		return logAndRespond(c, nil, nil, "Invalid key", fiber.StatusUnauthorized)
	}

	if keyEntity.KeyCountUsed >= keyEntity.KeyCount {
		return logAndRespond(c, keyEntity, nil, "Key usage limit reached", fiber.StatusTooManyRequests)
	}

	templateID := c.Params("templateId")
	if templateID == "" {
		return logAndRespond(c, keyEntity, nil, "No template provided", fiber.StatusBadRequest)
	}

	templateService := template.NewService(template.Repository{})
	templateEntity, err := templateService.GetByUUID(templateID)
	if err != nil {
		return logAndRespond(c, keyEntity, nil, fmt.Sprintf("failed to get template: %v", err), fiber.StatusInternalServerError)
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return logAndRespond(c, nil, nil, fmt.Sprintf("failed to parse request body: %v", err), fiber.StatusBadRequest)
	}

	format := c.Query("format", "A4")

	pdfURL, err := pdfjob.GeneratePdfForKey(ctx, keyEntity, templateEntity, data, format)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, err.Error(), fiber.StatusInternalServerError)
	}

	go logPdfGeneration(keyEntity.ID, templateEntity.ID, c.Body(), pdfURL, "", entities.Success)
	fmt.Printf("Total execution time: %v\n", time.Since(startTime))

	return c.JSON(fiber.Map{"path": pdfURL})
}

func logPdfGeneration(keyID, templateID uint, requestBody []byte, pdfURL, errorMessage string, statusCode entities.StatusCode) {
	logService := logs.NewService(logs.Repository{})

	responseBody := fiber.Map{"path": pdfURL}
	responseBodyBytes, err := json.Marshal(responseBody)
	if err != nil {
		fmt.Printf("failed to marshal response body: %v\n", err)
		return
	}

	logEntry := &entities.Log{
		TemplateID:   templateID,
		KeyID:        keyID,
		CalledAt:     time.Now(),
		RequestBody:  requestBody,
		ResponseBody: datatypes.JSON(responseBodyBytes),
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
	}

	if err := logService.CreateLog(logEntry); err != nil {
		fmt.Printf("failed to log PDF generation: %v\n", err)
	}
}

func logAndRespond(c *fiber.Ctx, k *entities.Key, t *entities.Template, errorMessage string, statusCode int) error {
	logService := logs.NewService(logs.Repository{})
	logEntry := &entities.Log{
		CalledAt:     time.Now(),
		RequestBody:  c.Body(),
		ResponseBody: datatypes.JSON([]byte(fmt.Sprintf(`{"message": "%s"}`, errorMessage))),
		StatusCode:   entities.StatusCode(statusCode),
		ErrorMessage: errorMessage,
	}

	if k != nil {
		logEntry.KeyID = k.ID
	}
	if t != nil {
		logEntry.TemplateID = t.ID
	}

	if err := logService.CreateLog(logEntry); err != nil {
		fmt.Printf("failed to log error: %v\n", err)
	}

	return c.Status(statusCode).JSON(fiber.Map{"message": errorMessage})
}
