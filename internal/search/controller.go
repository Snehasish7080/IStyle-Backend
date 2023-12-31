package search

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type SearchController struct {
	storage *SearchStorage
}

func NewTagController(storage *SearchStorage) *SearchController {
	return &SearchController{
		storage: storage,
	}
}

type searchTextResponse struct {
	Data    []searchTextResult `json:"data"`
	Message string             `json:"message"`
	Success bool               `json:"success"`
}

func (s *SearchController) getSearchByTextResult(c *fiber.Ctx) error {
	text := c.Params("text")

	if text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(searchTextResponse{
			Message: "invalid request",
			Success: false,
		})
	}
	result, err := s.storage.searchText(text, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(searchTextResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []searchTextResult
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(searchTextResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}

type styleByTextResponse struct {
	Data    []stylesByTextResult `json:"data"`
	Message string               `json:"message"`
	Success bool                 `json:"success"`
}

func (s *SearchController) getStyleByTextResult(c *fiber.Ctx) error {
	text := c.Params("text")

	if text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(styleByTextResponse{
			Message: "invalid request",
			Success: false,
		})
	}
	result, err := s.storage.stylesByText(text, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(styleByTextResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []stylesByTextResult
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(styleByTextResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}
