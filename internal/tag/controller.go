package tag

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TagController struct {
	storage *TagStorage
}

func NewTagController(storage *TagStorage) *TagController {
	return &TagController{
		storage: storage,
	}
}

var validate = validator.New()

type createTagRequest struct {
	Name string `json:"name" validate:"required"`
}

type createTagResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (t *TagController) createTag(c *fiber.Ctx) error {
	var req createTagRequest

	c.BodyParser(&req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(createTagResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(createTagResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	message, err := t.storage.create(req.Name, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(createTagResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(createTagResponse{
		Success: true,
		Message: message,
	})
}

type tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getAllTagResponse struct {
	Data    []tag  `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (t *TagController) getAllTags(c *fiber.Ctx) error {
	result, err := t.storage.getAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getAllTagResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []tag
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(getAllTagResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}
