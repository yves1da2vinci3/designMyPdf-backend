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

	// Backblaze upload (same env names as frontend: BACKBLAZE_KEY_ID, BACKBLAZE_APP_KEY, BACKBLAZE_BUCKET_NAME)
	var b2 *storage.BackblazeStorage
	keyID, appKey, bucketName, b2OK := storage.B2ConfigFromEnv()
	if b2OK {
		var err error
		b2, err = storage.NewBackblazeStorage(keyID, appKey, bucketName)
		if err != nil {
			log.Printf("Warning: Backblaze storage init failed: %v", err)
		}
	} else {
		log.Println("Warning: Backblaze env not set or still placeholders (BACKBLAZE_KEY_ID / BACKBLAZE_APP_KEY / BACKBLAZE_BUCKET_NAME or legacy B2_*), image upload disabled")
	}
	UploadRouter(api, b2)

	// Handle PDF generation
	api.Post("/generate-pdf/:templateId", handlers.GeneratePdf)
}
