package explore

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ExploreController struct {
	storage *ExploreStorage
}

func NewFeedController(storage *ExploreStorage) *ExploreController {
	return &ExploreController{
		storage: storage,
	}
}

var validate = validator.New()

type getUserFeedResponse struct {
	Data    []exploreStyle `json:"data"`
	Message string         `json:"message"`
	Success bool           `json:"success"`
}

func (f *ExploreController) getUserExplore(c *fiber.Ctx) error {
	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(getUserFeedResponse{
			Message: "Something went wrong",
			Success: false,
		})
	}
	result, err := f.storage.explore(userName, c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getUserFeedResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []exploreStyle
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(getUserFeedResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}
