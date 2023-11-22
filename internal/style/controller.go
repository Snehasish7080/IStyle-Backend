package style

import (
	"encoding/json"
	"errors"
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
	Data    *GetStyleUploadUrlData `json:"data"`
	Message string                 `json:"message"`
	Success bool                   `json:"success"`
}

type getStyleUploadUrlRequest struct {
	LinkCount int `json:"linkCount"`
}

func getLinks(ch chan<- GetStyleUploadUrl, wg *sync.WaitGroup) {

	defer wg.Done()

	id := uuid.New()
	linkUrl, _ := signedurl.GetSignedUrl(id.String())

	ch <- GetStyleUploadUrl{
		Url: linkUrl,
		Key: id.String(),
	}

}

func (t *StyleController) getStyleUploadUrl(c *fiber.Ctx) error {

	var req getStyleUploadUrlRequest
	c.BodyParser(&req)

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
	var structData *GetStyleUploadUrlData
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(getStyleUploadUrlResponse{
		Data:    structData,
		Message: "url created successfully",
		Success: true,
	})
}

type createStyleRequest struct {
	Image string   `json:"image"`
	Links []link   `json:"links"`
	Tags  []string `json:"tags"`
}
type createStyleResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (s *StyleController) createStyle(c *fiber.Ctx) error {
	var req createStyleRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(createStyleResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	message, err := s.storage.create(userName, req.Image, req.Links, req.Tags, c.Context())

	fmt.Println(err)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(createStyleResponse{
			Message: "something went wrong",
			Success: false,
		})

	}

	return c.Status(fiber.StatusBadRequest).JSON(createStyleResponse{
		Message: message,
		Success: true,
	})
}

type markTrendRequest struct {
	Id string `json:"id"`
}
type markTrendResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (s *StyleController) markTrend(c *fiber.Ctx) error {
	var req markTrendRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(markTrendResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	message, err := s.storage.trend(userName, req.Id, c.Context())

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(markTrendResponse{
			Message: "something went wrong",
			Success: false,
		})

	}

	return c.Status(fiber.StatusBadRequest).JSON(markTrendResponse{
		Message: message,
		Success: true,
	})
}

type styleClickedRequest struct {
	Id string `json:"id"`
}
type styleClickedResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (s *StyleController) styleClicked(c *fiber.Ctx) error {
	var req styleClickedRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(styleClickedResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	message, err := s.storage.clicked(userName, req.Id, c.Context())

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(styleClickedResponse{
			Message: "something went wrong",
			Success: false,
		})

	}

	return c.Status(fiber.StatusBadRequest).JSON(styleClickedResponse{
		Message: message,
		Success: true,
	})
}
