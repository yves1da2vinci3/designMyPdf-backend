package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/pkg/amqp"
	"designmypdf/pkg/auth"
	"designmypdf/pkg/key"
	"designmypdf/pkg/logs"
	"designmypdf/pkg/marketplace"
	"designmypdf/pkg/namespace"
	"designmypdf/pkg/pdfjob"
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

	// Synchronous PDF generation (unchanged)
	api.Post("/generate-pdf/:templateId", handlers.GeneratePdf)

	// Async PDF generation via RabbitMQ
	var jobSvc *pdfjob.Service
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL != "" {
		amqpClient, err := amqp.NewClient(rabbitmqURL)
		if err != nil {
			log.Printf("Warning: RabbitMQ connect failed: %v — async routes disabled", err)
		} else {
			jobSvc = pdfjob.NewService(amqpClient)
		}
	} else {
		log.Println("Warning: RABBITMQ_URL not set — async PDF routes disabled")
	}

	if jobSvc != nil {
		api.Post("/generate-pdf/:templateId/async", handlers.GeneratePdfAsync(jobSvc))
		api.Get("/pdf-jobs/:jobId", handlers.GetJobStatus(jobSvc))
	}

	// AI quota
	AiQuotaRouter(api)

	// Webhook subscription management
	WebhookRouter(api)
}
