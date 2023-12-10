package feed

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type FeedController struct {
	storage *FeedStorage
}

func NewFeedController(storage *FeedStorage) *FeedController {
	return &FeedController{
		storage: storage,
	}
}

var validate = validator.New()

type style struct {
	Id         string `json:"id"`
	Image      string `json:"image"`
	Links      []link `json:"links"`
	Created_at string `json:"created_at"`
}
type link struct {
	Id    string `json:"id"`
	Image string `json:"image"`
	Url   string `json:"url"`
}
type getUserFeedResponse struct {
	Data    []style `json:"data"`
	Message string  `json:"message"`
	Success bool    `json:"success"`
}

func (f *FeedController) getUserFeed(c *fiber.Ctx) error {
	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(getUserFeedResponse{
			Message: "Something went wrong",
			Success: false,
		})
	}
	result, err := f.storage.feed(userName, c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getUserFeedResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []style
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(getUserFeedResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}
