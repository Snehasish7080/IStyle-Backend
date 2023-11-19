package style

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zone/IStyle/pkg/signedurl"
)

type StyleController struct {
	storage *StyleStorage
}

func NewStyleController(storage *StyleStorage) *StyleController {
	return &StyleController{
		storage: storage,
	}
}

var validate = validator.New()

type GetStyleUploadUrl struct {
	Url string `json:"url"`
	Key string `json:"key"`
}
type GetStyleUploadUrlData struct {
	Style GetStyleUploadUrl   `json:"style"`
	Links []GetStyleUploadUrl `json:"links"`
}

type getStyleUploadUrlResponse struct {
	Data    GetStyleUploadUrlData `json:"data"`
	Message string                `json:"message"`
	Success bool                  `json:"success"`
}

type getStyleUploadUrlRequest struct {
	LinkCount int `json:"linkCount" validate:"required"`
}

func getLinks(ch chan<- GetStyleUploadUrl, wg *sync.WaitGroup) error {

	defer wg.Done()

	id := uuid.New()
	linkUrl, err := signedurl.GetSignedUrl(id.String())

	if err != nil {
		return err
	}

	ch <- GetStyleUploadUrl{
		Url: linkUrl,
		Key: id.String(),
	}

	return nil
}

func (t *StyleController) getStyleUploadUrl(c *fiber.Ctx) error {

	var req getStyleUploadUrlRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getStyleUploadUrlResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getStyleUploadUrlResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	id := uuid.New()
	mainStyleUrl, err := signedurl.GetSignedUrl(id.String())
	var links []GetStyleUploadUrl

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getStyleUploadUrlResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	ch := make(chan GetStyleUploadUrl)
	var wg sync.WaitGroup

	for i := 0; i < req.LinkCount; i++ {
		wg.Add(1)
		go getLinks(ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for link := range ch {
		links = append(links, link)

	}

	var result = GetStyleUploadUrlData{
		Style: GetStyleUploadUrl{
			Url: mainStyleUrl,
			Key: id.String(),
		},
		Links: links,
	}

	jsonData, _ := json.Marshal(result)
	var structData GetStyleUploadUrlData
	json.Unmarshal(jsonData, &structData)

	fmt.Println("running .. this")

	return c.Status(fiber.StatusOK).JSON(getStyleUploadUrlResponse{
		Data:    structData,
		Message: "url created successfully",
		Success: true,
	})
}
