package style

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddStyleRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *StyleController) {
	auth := app.Group("/auth")

	style := auth.Group("/style", middleware.VerifyUser)
	style.Post("/upload-url", controller.getStyleUploadUrl)

}
