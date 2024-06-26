package handlers

import (
	"designmypdf/config/storage"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/template"
	"designmypdf/utils"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func GeneratePdf(c *fiber.Ctx) error {
	keyService := key.NewService(key.Repository{})

	// Handle Key
	keyValue := c.Get("dmp_KEY")
	if keyValue == "" {
		return logAndRespond(c, nil, nil, "No key provided", fiber.StatusUnauthorized)
	}

	keyEntity, err := keyService.GetKeyByValue(keyValue)
	if err != nil {
		return logAndRespond(c, nil, nil, "Invalid key", fiber.StatusUnauthorized)
	}

	// Check if the key's usage limit is reached
	if keyEntity.KeyCountUsed >= keyEntity.KeyCount {
		return logAndRespond(c, keyEntity, nil, "Key usage limit reached", fiber.StatusTooManyRequests)
	}

	// Get the Template
	templateID, err := c.ParamsInt("templateId")
	if err != nil {
		return logAndRespond(c, keyEntity, nil, fmt.Sprintf("failed to parse templateId: %v", err), fiber.StatusBadRequest)
	}

	templateService := template.NewService(template.Repository{})
	templateEntity, err := templateService.Get(uint(templateID))
	if err != nil {
		return logAndRespond(c, keyEntity, nil, fmt.Sprintf("failed to get template: %v", err), fiber.StatusInternalServerError)
	}

	// Render HTML from template
	var data map[string]interface{}
	if err := json.Unmarshal(c.Body(), &data); err != nil {
		return logAndRespond(c, nil, nil, fmt.Sprintf("failed to parse request body: %v", err), fiber.StatusBadRequest)
	}
	renderedHTML, err := utils.RenderTemplate(templateEntity.Content, data)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to render template: %v", err), fiber.StatusInternalServerError)
	}

	// Combine structural HTML with rendered content
	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Preview</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    %s
    <style>
       %s
    </style>
</head>
<body class="overflow-x-hidden overflow-y-auto">
    <div class="content">
        %s
    </div>
</body>
</html>`, utils.ImportFontCreation(templateEntity.Fonts), utils.FontCssCreation(templateEntity.Fonts), renderedHTML)

	// Generate PDF
	id := uuid.New()
	outputPath := fmt.Sprintf("./uploads/template/template_%s.pdf", id.String())

	format := c.Query("format")
	if format == "" {
		format = "A4"
	}

	err = utils.GeneratePDF(fullHTML, format, outputPath)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to generate PDF: %v", err), fiber.StatusInternalServerError)
	}

	fmt.Printf("PDF saved to: %s\n", outputPath)

	// Initialize Firebase Storage
	bucket, err := storage.InitializeFirebaseStorage()
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to initialize Firebase storage: %v", err), fiber.StatusInternalServerError)
	}

	// Upload PDF file to Firebase Storage
	storagePath := "/templates/" + id.String() + ".pdf"
	url, err := storage.UploadFile(bucket, outputPath, storagePath)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to upload PDF to Firebase Storage: %v", err), fiber.StatusInternalServerError)
	}

	// Delete the PDF file locally
	err = os.Remove(outputPath)
	if err != nil {
		fmt.Printf("failed to delete local PDF file: %v\n", err)
	}

	// Log the successful PDF generation
	logService := logs.NewService(logs.Repository{})
	responseBody := fiber.Map{
		"path": url,
	}
	responseBodyBytes, err := json.Marshal(responseBody)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to marshal response body: %v", err), fiber.StatusInternalServerError)
	}

	// Increase key usage count
	err = keyService.IncreaseUsageCount(keyEntity.ID)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to increase usage count: %v", err), fiber.StatusInternalServerError)
	}

	logEntry := &entities.Log{
		TemplateID:   uint(templateID),
		KeyID:        keyEntity.ID,
		CalledAt:     time.Now(),
		RequestBody:  c.Body(),
		ResponseBody: datatypes.JSON(responseBodyBytes),
		StatusCode:   entities.Success,
		ErrorMessage: "",
	}

	err = logService.CreateLog(logEntry)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to log PDF generation: %v", err), fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"path": url,
	})
}

func logAndRespond(c *fiber.Ctx, key *entities.Key, template *entities.Template, errorMessage string, statusCode int) error {
	// Log the error
	logService := logs.NewService(logs.Repository{})
	logEntry := &entities.Log{
		CalledAt:     time.Now(),
		RequestBody:  c.Body(),
		ResponseBody: datatypes.JSON([]byte(fmt.Sprintf(`{"message": "%s"}`, errorMessage))),
		StatusCode:   entities.StatusCode(statusCode),
		ErrorMessage: errorMessage,
	}

	if key != nil {
		logEntry.KeyID = key.ID
	} else {
		// Handle the case where key is nil
		logEntry.KeyID = 0 // or some other default value or handling
	}

	if template != nil {
		logEntry.TemplateID = template.ID
	} else {
		// Handle the case where template is nil
		logEntry.TemplateID = 0 // or some other default value or handling
	}

	err := logService.CreateLog(logEntry)
	if err != nil {
		fmt.Printf("failed to log error: %v\n", err)
	}

	// Respond with the error message
	return c.Status(statusCode).JSON(fiber.Map{
		"message": errorMessage,
	})
}
