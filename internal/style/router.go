package style

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddStyleRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *StyleController) {
	auth := app.Group("/auth")

	style := auth.Group("/style", middleware.VerifyUser)
	style.Post("/upload-url", controller.getStyleUploadUrl)
	style.Post("/create", controller.createStyle)
	style.Get("/all", controller.getAllUserStyles)
	style.Post("/mark-trend", controller.markTrend)
	style.Post("/unmark-trend", controller.unMarkTrend)
	style.Post("/style-clicked", controller.styleClicked)
	style.Get("/:id", controller.getStyleById)
	style.Get("/liked/:id", controller.getALlLikedUsers)

	styleByUserName := style.Group("/user/:userName", middleware.CheckUserNameExists)
	styleByUserName.Get("/", controller.getAllStylesByUserName)
}
