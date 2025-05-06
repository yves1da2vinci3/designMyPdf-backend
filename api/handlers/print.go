package handlers

import (
	"context"
	"crypto/md5"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/storage"
	"designmypdf/pkg/template"
	"designmypdf/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

// Structure pour le cache de PDF
type pdfCache struct {
	cache map[string]string // clé: hash, valeur: URL
	mu    sync.RWMutex
	maxItems int
}

var (
	pdfCacheInstance = &pdfCache{
		cache: make(map[string]string),
		maxItems: 100, // Limiter la taille du cache
	}
	storageInstance *storage.BackblazeStorage
	storageMu       sync.Mutex
)

// Générer un hash pour le contenu du template et les données
func generateHash(templateContent string, data map[string]interface{}, format string) string {
	dataBytes, _ := json.Marshal(data)
	hash := md5.Sum([]byte(templateContent + string(dataBytes) + format))
	return hex.EncodeToString(hash[:])
}

// Récupérer l'instance de stockage (lazy initialization)
func getStorageInstance() (*storage.BackblazeStorage, error) {
	storageMu.Lock()
	defer storageMu.Unlock()
	
	if storageInstance != nil {
		return storageInstance, nil
	}
	
	// Charger les variables d'environnement si nécessaire
	if err := godotenv.Load(); err != nil {
		// Ignorer l'erreur si le fichier n'existe pas
	}
	
	b2Storage, err := storage.NewBackblazeStorage(
		os.Getenv("B2_ACCOUNT_ID"),
		os.Getenv("B2_APPLICATION_KEY"),
		os.Getenv("B2_BUCKET_NAME"),
	)
	if err != nil {
		return nil, err
	}
	
	storageInstance = b2Storage
	return storageInstance, nil
}

func GeneratePdf(c *fiber.Ctx) error {
	// Début du chronométrage
	startTime := time.Now()

	// Utiliser un contexte avec timeout pour éviter les opérations bloquantes
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	// Garbage collection préventif pour libérer la mémoire
	debug.FreeOSMemory()

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

	// Parse the request body
	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return logAndRespond(c, nil, nil, fmt.Sprintf("failed to parse request body: %v", err), fiber.StatusBadRequest)
	}
	
	// Get format parameter
	format := c.Query("format", "A4")
	
	// Générer un hash pour vérifier le cache
	contentHash := generateHash(templateEntity.Content, data, format)
	
	// Vérifier si le PDF est déjà en cache
	pdfCacheInstance.mu.RLock()
	cachedURL, found := pdfCacheInstance.cache[contentHash]
	pdfCacheInstance.mu.RUnlock()
	
	if found {
		fmt.Printf("PDF found in cache, returning cached URL: %s\n", cachedURL)
		
		// Incrémenter le compteur d'utilisation de la clé de manière asynchrone
		go func() {
			if err := keyService.IncreaseUsageCount(keyEntity.ID); err != nil {
				fmt.Printf("warning: failed to increase usage count: %v\n", err)
			}
		}()
		
		// Log asynchrone
		go logPdfGeneration(keyEntity.ID, templateEntity.ID, c.Body(), cachedURL, "", entities.Success)
		
		return c.JSON(fiber.Map{
			"path": cachedURL,
		})
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

	// Combiner le HTML structurel avec le contenu rendu
	fontImports := utils.ImportFontCreation(templateEntity.Fonts)
	fontCSS := utils.FontCssCreation(templateEntity.Fonts)
	
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
</html>`, fontImports, fontCSS, renderedHTML)

	// Créer le dossier de uploads/template s'il n'existe pas
	if err := os.MkdirAll("./uploads/template", 0755); err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to create upload directory: %v", err), fiber.StatusInternalServerError)
	}

	// Générer un ID unique pour le PDF
	id := uuid.New()
	outputPath := fmt.Sprintf("./uploads/template/template_%s.pdf", id.String())

	// Forcer un garbage collection avant de générer le PDF pour libérer de la mémoire
	runtime.GC()

	// Utiliser un timeout pour la génération PDF
	pdfDone := make(chan error, 1)
	go func() {
		// Génération du PDF optimisée
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

	// Obtenir l'instance de stockage
	b2Storage, err := getStorageInstance()
	if err != nil {
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to initialize Backblaze storage: %v", err), fiber.StatusInternalServerError)
	}

	// Upload PDF file to Backblaze
	storagePath := fmt.Sprintf("templates/%s.pdf", id.String())
	
	// Optimisation: Exécuter l'upload et l'incrémentation de compteur en parallèle
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
		// Nettoyer les fichiers temporaires en cas d'erreur
		os.Remove(outputPath)
		return logAndRespond(c, keyEntity, templateEntity, fmt.Sprintf("failed to upload PDF to Backblaze: %v", uploadErr), fiber.StatusInternalServerError)
	}
	
	if countErr != nil {
		fmt.Printf("warning: failed to increase usage count: %v\n", countErr)
	}

	// Mettre en cache l'URL du PDF
	pdfCacheInstance.mu.Lock()
	// Si le cache est plein, supprimer un élément au hasard
	if len(pdfCacheInstance.cache) >= pdfCacheInstance.maxItems {
		for k := range pdfCacheInstance.cache {
			delete(pdfCacheInstance.cache, k)
			break
		}
	}
	pdfCacheInstance.cache[contentHash] = url
	pdfCacheInstance.mu.Unlock()

	// Delete the PDF file locally (en arrière-plan)
	go func() {
		time.Sleep(500 * time.Millisecond) // Attendre un peu pour s'assurer que le fichier est bien lu
		if err := os.Remove(outputPath); err != nil {
			fmt.Printf("warning: failed to delete local PDF file: %v\n", err)
		}
	}()

	// Log asynchrone
	go logPdfGeneration(keyEntity.ID, templateEntity.ID, c.Body(), url, "", entities.Success)

	fmt.Printf("Temps total d'exécution: %v\n", time.Since(startTime))

	return c.JSON(fiber.Map{
		"path": url,
	})
}

// Fonction utilitaire pour logger les générations de PDF de manière asynchrone
func logPdfGeneration(keyID, templateID uint, requestBody []byte, url, errorMessage string, statusCode entities.StatusCode) {
	logService := logs.NewService(logs.Repository{})
	
	responseBody := fiber.Map{
		"path": url,
	}
	
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
