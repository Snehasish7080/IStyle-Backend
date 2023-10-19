package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/internal/middleware"
)

func AddUserRoutes(app *fiber.App, middleware *middleware.AuthMiddleware, controller *UserController) {
	auth := app.Group("/auth")

	// add routes here
	auth.Post("/sign-up", controller.register)
	auth.Post("/login", controller.loginUser)

	// verify token
	verify := auth.Group("/verify", middleware.VerifyOtpToken)
	verify.Post("/", controller.verifyOtp)

	// user
	user := auth.Group("/user", middleware.VerifyUser)
	user.Get("/", controller.getUserDetail)
	user.Post("/update", controller.updateUserDetail)

}
