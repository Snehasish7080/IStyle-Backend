package tag

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddTagRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *TagController) {
	auth := app.Group("/auth/tag")

	// add routes here

	auth.Post("/create", controller.createTag)
	auth.Get("/all", controller.getAllTags)

}
