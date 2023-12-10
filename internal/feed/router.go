package feed

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddFeedRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *FeedController) {
	auth := app.Group("/auth")

	feed := auth.Group("/feed", middleware.VerifyUser)
	feed.Get("/", controller.getUserFeed)
}
