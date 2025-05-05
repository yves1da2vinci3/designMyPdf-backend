package handlers

import (
	"context"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/storage"
	"designmypdf/pkg/template"
	"designmypdf/utils"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

func GeneratePdf(c *fiber.Ctx) error {
	// Début du chronométrage
	startTime := time.Now()

	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		return logAndRespond(c, nil, nil, "Erreur lors du chargement des variables d'environnement", fiber.StatusInternalServerError)
	}

	// Utiliser un contexte avec timeout pour éviter les opérations bloquantes
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

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
	templateID := c.Params("templateId")
	if templateID == "" {
		return logAndRespond(c, keyEntity, nil, "No template provided", fiber.StatusBadRequest)
	}

	templateService := template.NewService(template.Repository{})
	templateEntity, err := templateService.GetByUUID(templateID)
	if err != nil {
		return logAndRespond(c, keyEntity, nil, fmt.Sprintf("failed to get template: %v", err), fiber.StatusInternalServerError)
	}

	// Render HTML from template
	var data map[string]interface{}
	if err := json.Unmarshal(c.Body(), &data); err != nil {
		return logAndRespond(c, nil, nil, fmt.Sprintf("failed to parse request body: %v", err), fiber.StatusBadRequest)
	}
	
	// Optimisation: Utiliser une goroutine pour la génération HTML
	var renderedHTML string
	var renderErr error
	
	htmlDone := make(chan bool, 1)
	go func() {
		renderedHTML, renderErr = utils.RenderTemplate(templateEntity.Content, data)
		htmlDone <- true
	}()

	// Attendre la fin de la génération HTML avec timeout
	select {
	case <-htmlDone:
		if renderErr != nil {
			return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to render template: %v", renderErr), fiber.StatusInternalServerError)
		}
	case <-time.After(5 * time.Second):
		return logAndRespond(c, keyEntity, templateEntity, "Template rendering timeout", fiber.StatusRequestTimeout)
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

	// Optimisation: Forcer un garbage collection avant de générer le PDF pour libérer de la mémoire
	runtime.GC()

	// Utiliser un timeout pour la génération PDF
	pdfDone := make(chan error, 1)
	go func() {
		pdfDone <- utils.GeneratePDF(fullHTML, format, outputPath)
	}()

	// Attendre la fin de la génération PDF avec timeout
	select {
	case err := <-pdfDone:
		if err != nil {
			return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to generate PDF: %v", err), fiber.StatusInternalServerError)
		}
	case <-time.After(15 * time.Second):
		return logAndRespond(c, keyEntity, templateEntity, "PDF generation timeout", fiber.StatusRequestTimeout)
	}

	fmt.Printf("PDF saved to: %s\n", outputPath)
	fmt.Printf("Temps écoulé jusqu'à la génération du PDF: %v\n", time.Since(startTime))

	// Initialize Backblaze Storage
	b2Storage, err := storage.NewBackblazeStorage(
		os.Getenv("B2_ACCOUNT_ID"),
		os.Getenv("B2_APPLICATION_KEY"),
		os.Getenv("B2_BUCKET_NAME"),
	)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to initialize Backblaze storage: %v", err), fiber.StatusInternalServerError)
	}

	// Upload PDF file to Backblaze
	storagePath := "templates/" + id.String() + ".pdf"
	
	// Optimisation: Exécuter l'upload, l'incrémentation de compteur et la journalisation en parallèle
	var wg sync.WaitGroup
	var url string
	var uploadErr error
	var countErr error
	
	wg.Add(2)
	
	// Upload du fichier
	go func() {
		defer wg.Done()
		url, uploadErr = b2Storage.UploadFile(ctx, outputPath, storagePath)
	}()
	
	// Increase key usage count en parallèle
	go func() {
		defer wg.Done()
		countErr = keyService.IncreaseUsageCount(keyEntity.ID)
	}()
	
	// Attendre la fin des opérations
	wg.Wait()
	
	// Vérification des erreurs
	if uploadErr != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to upload PDF to Backblaze: %v", uploadErr), fiber.StatusInternalServerError)
	}
	
	if countErr != nil {
		fmt.Printf("warning: failed to increase usage count: %v\n", countErr)
	}

	// Delete the PDF file locally
	os.Remove(outputPath) // No need to check error, not critical

	// Log the successful PDF generation
	logService := logs.NewService(logs.Repository{})
	responseBody := fiber.Map{
		"path": url,
	}
	responseBodyBytes, err := json.Marshal(responseBody)
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to marshal response body: %v", err), fiber.StatusInternalServerError)
	}

	// Log asynchrone
	go func() {
		logEntry := &entities.Log{
			TemplateID:   templateEntity.ID,
			KeyID:        keyEntity.ID,
			CalledAt:     time.Now(),
			RequestBody:  c.Body(),
			ResponseBody: datatypes.JSON(responseBodyBytes),
			StatusCode:   entities.Success,
			ErrorMessage: "",
		}

		if err := logService.CreateLog(logEntry); err != nil {
			fmt.Printf("failed to log PDF generation: %v\n", err)
		}
	}()

	fmt.Printf("Temps total d'exécution: %v\n", time.Since(startTime))

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
