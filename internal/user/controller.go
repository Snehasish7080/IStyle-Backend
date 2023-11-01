package user

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/zone/IStyle/pkg/otp"
)

type UserController struct {
	storage *UserStorage
}

func NewUserController(storage *UserStorage) *UserController {
	return &UserController{
		storage: storage,
	}
}

var validate = validator.New()

type signUpRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	UserName  string `json:"userName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
}

type signUpResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) register(c *fiber.Ctx) error {
	var req signUpRequest

	c.BodyParser(&req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(signUpResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(signUpResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	token, err := u.storage.signUp(req.FirstName, req.LastName, req.UserName, req.Email, req.Password, c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(signUpResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(signUpResponse{
		Token:   token,
		Success: true,
		Message: "Otp sent successfully",
	})
}

type verifyRequest struct {
	Otp string `json:"otp" validate:"required"`
}
type verifyResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) verifyEmail(c *fiber.Ctx) error {

	var req verifyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(verifyResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(signUpResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	token, err := u.storage.verifyEmail(req.Otp, userName, c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(verifyResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(verifyResponse{
		Token:   token,
		Success: true,
		Message: "Verified",
	})

}

func (u *UserController) verifyMobile(c *fiber.Ctx) error {

	var req verifyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(verifyResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(signUpResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	token, err := u.storage.verifyMobile(req.Otp, userName, c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(verifyResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(verifyResponse{
		Token:   token,
		Success: true,
		Message: "Verified",
	})

}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type loginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) loginUser(c *fiber.Ctx) error {
	var req loginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(loginResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(loginResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	token, err := u.storage.login(req.Email, req.Password, c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(loginResponse{
			Message: err.Error(),
			Success: false,
		})
	}
	return c.Status(fiber.StatusOK).JSON(loginResponse{
		Token:   token,
		Success: true,
		Message: "Found Successfully",
	})
}

type userDetailResponse struct {
	Data    userDetail `json:"data"`
	Message string     `json:"message"`
	Success bool       `json:"success"`
}
type userDetail struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserName  string `json:"userName"`
}

func (u *UserController) getUserDetail(c *fiber.Ctx) error {
	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}
	user, err := u.storage.getUser(userName, c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(userDetailResponse{
			Message: err.Error(),
			Success: false,
		})
	}
	return c.Status(fiber.StatusOK).JSON(userDetailResponse{
		Data: userDetail{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserName:  user.UserName,
		},
		Message: "found successfully",
		Success: true,
	})
}

type updateUserDetailRequest struct {
	Mobile string `json:"mobile" validate:"required"`
	Image  string `json:"image"`
}
type updateUserDetailResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) updateUserDetail(c *fiber.Ctx) error {
	var req updateUserDetailRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(updateUserDetailResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(updateUserDetailResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	var updateFields map[string]interface{}
	data, _ := json.Marshal(req)
	json.Unmarshal(data, &updateFields)

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	for k := range updateFields {
		if updateFields[k] == "" {
			delete(updateFields, k)
		}
	}

	generatedOtp := otp.EncodeToString(6)

	if generatedOtp != "" {
		updateFields["mobileOtp"] = generatedOtp
	}

	message, err := u.storage.updateUser(userName, updateFields, c.Context())

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(updateUserDetailResponse{
			Message: "Update failed",
			Success: false,
		})

	}

	return c.Status(fiber.StatusOK).JSON(updateUserDetailResponse{
		Message: message,
		Success: true,
	})
}
