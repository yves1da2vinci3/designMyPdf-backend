package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/marketplace"
	"designmypdf/pkg/namespace"
	"designmypdf/pkg/storage"
	"designmypdf/pkg/template"
	"designmypdf/pkg/user"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	api.Get("/", handlers.HelloWorld)

	// Auth
	authService := auth.NewService(user.Repository{})
	AuthRouter(api, authService)
	// Namespace
	namespaceService := namespace.NewService(namespace.Repository{})
	NampesaceRouter(api, namespaceService)
	// Template
	templateService := template.NewService(template.Repository{})
	TemplateRouter(api, templateService)
	// Key
	keyService := key.NewService(key.Repository{})
	KeyRouter(api, keyService)
	// Logs
	logService := logs.NewService(logs.Repository{})
	LogRouter(api, logService)

	// Marketplace
	marketplaceService := marketplace.NewService()
	MarketplaceRouter(api, marketplaceService)

	// Backblaze upload
	var b2 *storage.BackblazeStorage
	accountID := os.Getenv("B2_ACCOUNT_ID")
	appKey := os.Getenv("B2_APPLICATION_KEY")
	bucketName := os.Getenv("B2_BUCKET_NAME")
	if accountID != "" && appKey != "" && bucketName != "" {
		var err error
		b2, err = storage.NewBackblazeStorage(accountID, appKey, bucketName)
		if err != nil {
			log.Printf("Warning: Backblaze storage init failed: %v", err)
		}
	} else {
		log.Println("Warning: Backblaze env vars not set, image upload disabled")
	}
	UploadRouter(api, b2)

	// Handle PDF generation
	api.Post("/generate-pdf/:templateId", handlers.GeneratePdf)
}
