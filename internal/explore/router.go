package explore

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddExploreRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *ExploreController) {
	auth := app.Group("/auth")

	feed := auth.Group("/explore", middleware.VerifyUser)
	feed.Get("/", controller.getUserExplore)
}
