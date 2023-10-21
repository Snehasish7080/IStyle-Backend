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

	// verify Email token
	verifyEmail := auth.Group("/verify/email", middleware.VerifyOtpToken)
	verifyEmail.Post("/", controller.verifyEmail)

	// verify Mobile token
	verifyMobile := auth.Group("/verify/mobile", middleware.VerifyOtpToken)
	verifyMobile.Post("/", controller.verifyMobile)

	// user
	user := auth.Group("/user", middleware.VerifyUser)
	user.Get("/", controller.getUserDetail)
	user.Post("/update", controller.updateUserDetail)

}
