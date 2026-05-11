package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/marketplace"

	"github.com/gofiber/fiber/v2"
)

func MarketplaceRouter(api fiber.Router, svc marketplace.Service) {
	mp := api.Group("/marketplace")
	mp.Get("/", handlers.ListMarketplace(svc))
	mp.Get("/my-listings", middleware.Protected(), handlers.GetMyListings(svc))
	mp.Get("/:id", handlers.GetMarketplaceListing(svc))
	mp.Post("/publish", middleware.Protected(), handlers.PublishToMarketplace(svc))
	mp.Post("/:id/copy", middleware.Protected(), handlers.CopyMarketplaceTemplate(svc))
	mp.Post("/:id/purchase", middleware.Protected(), handlers.PurchaseMarketplaceTemplate(svc))
}
