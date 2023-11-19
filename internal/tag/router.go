package tag

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddTagRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *TagController) {
	tag := app.Group("/auth/tag")

	// add routes here

	tag.Post("/create", controller.createTag)
	tag.Get("/all", controller.getAllTags)

}
