package routes

import (
	"designmypdf/api/handlers"
	"designmypdf/api/middleware"
	"designmypdf/pkg/template"

	"github.com/gofiber/fiber/v2"
)

func TemplateRouter(api fiber.Router, templateService template.Service) {
	// auth
	templateRouter := api.Group("/templates", middleware.Protected())
	templateRouter.Post("/:namespaceID", handlers.CreateTemplate(templateService))
	templateRouter.Delete("/:templateID", handlers.DeleteTemplate(templateService))
	templateRouter.Put("/:templateID", handlers.UpdateTemplate(templateService))
	templateRouter.Put("/:templateID/namespace/:namespaceID", handlers.ChangeTemplateNamespace(templateService))
	templateRouter.Get("/", handlers.GetTemplates(templateService))
}
