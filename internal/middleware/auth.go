package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/pkg/jwtclaim"
)

type AuthMiddleware struct {
	storage *MiddlewareStorage
}

func NewAuthMiddleware(storage *MiddlewareStorage) *AuthMiddleware {
	return &AuthMiddleware{
		storage: storage,
	}
}

func (a *AuthMiddleware) VerifyOtpToken(c *fiber.Ctx) error {

	reqToken := c.Request().Header.Peek("Authorization")

	userName, valid := jwtclaim.ExtractVerifyUsername(string(reqToken))

	if !valid {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized access")
	}
	c.Locals("userName", userName)
	return c.Next()
}

func (a *AuthMiddleware) VerifyUser(c *fiber.Ctx) error {

	reqToken := c.Request().Header.Peek("Authorization")

	userName, valid := jwtclaim.ExtractVerifyUsername(string(reqToken))

	if !valid {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized access")
	}
	c.Locals("userName", userName)
	return c.Next()
}
