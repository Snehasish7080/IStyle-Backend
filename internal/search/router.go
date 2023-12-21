package search

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddSearchRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *SearchController) {
	search := app.Group("/auth/search", middleware.VerifyUser)

	// add routes here

	search.Get("/:text", controller.getSearchByTextResult)
}
