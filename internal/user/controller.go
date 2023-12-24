package user

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zone/IStyle/pkg/otp"
	"github.com/zone/IStyle/pkg/signedurl"
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
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	UserName         string `json:"userName"`
	Bio              string `json:"bio"`
	ProfilePic       string `json:"profilePic"`
	IsMobileVerified bool   `json:"isMobileVerified"`
	IsComplete       bool   `json:"isComplete"`
	IsFollowing      bool   `json:"isFollowing"`
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
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			UserName:         user.UserName,
			Bio:              user.Bio,
			ProfilePic:       user.ProfilePic,
			IsMobileVerified: user.IsMobileVerified,
			IsComplete:       user.IsComplete,
		},
		Message: "found successfully",
		Success: true,
	})
}

type updateUserDetailRequest struct {
	ProfilePic string `json:"profilePic"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Bio        string `json:"bio"`
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

type updateUserMobileRequest struct {
	Mobile string `json:"mobile" validate:"required"`
}
type updateUserMobileResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) updateUserMobile(c *fiber.Ctx) error {
	var req updateUserMobileRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(updateUserMobileResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(updateUserMobileResponse{
			Message: "Invalid request body",
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return errors.New("not able to covert")
	}

	generatedOtp := otp.EncodeToString(6)

	message, err := u.storage.updateMobile(userName, req.Mobile, generatedOtp, c.Context())
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

type GetProfileUploadKeyData struct {
	Url string `json:"url"`
	Key string `json:"key"`
}
type getProfileUploadKeyResponse struct {
	Data    GetProfileUploadKeyData `json:"data"`
	Message string                  `json:"message"`
	Success bool                    `json:"success"`
}

func (u *UserController) getProfileUploadKey(c *fiber.Ctx) error {
	id := uuid.New()
	url, err := signedurl.GetSignedUrl(id.String())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getProfileUploadKeyResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(getProfileUploadKeyResponse{
		Data: GetProfileUploadKeyData{
			Url: url,
			Key: id.String(),
		},
		Message: "upload url",
		Success: true,
	})
}

func (u *UserController) getUserDetailByUserName(c *fiber.Ctx) error {
	userName := c.Params("userName")

	if userName == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(userDetailResponse{
			Message: "invalid request",
			Success: false,
		})
	}
	localData := c.Locals("userName")
	loggedInuser, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(userDetailResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	user, err := u.storage.getUserByUserName(userName, loggedInuser, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(userDetailResponse{
			Message: err.Error(),
			Success: false,
		})
	}
	return c.Status(fiber.StatusOK).JSON(userDetailResponse{
		Data: userDetail{
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			UserName:    user.UserName,
			Bio:         user.Bio,
			ProfilePic:  user.ProfilePic,
			IsFollowing: user.IsFollowing,
		},
		Message: "found successfully",
		Success: true,
	})
}

type markUserFavTagsRequest struct {
	Tags []string `json:"tags"`
}
type markUserFavTagsResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) markUserFavTags(c *fiber.Ctx) error {
	var req markUserFavTagsRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(markUserFavTagsResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(markUserFavTagsResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	message, err := u.storage.markFavTags(userName, req.Tags, c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(markUserFavTagsResponse{
			Message: "somthing went wrong",
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(markUserFavTagsResponse{
		Message: message,
		Success: true,
	})
}

type followUserRequest struct {
	UserName string `json:"userName"`
}

type followUserResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (u *UserController) followUser(c *fiber.Ctx) error {
	var req followUserRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(followUserResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(followUserResponse{
			Message: "something went wrong",
			Success: false,
		})
	}
	message, err := u.storage.follow(userName, req.UserName, c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(followUserResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(followUserResponse{
		Message: message,
		Success: true,
	})
}

func (u *UserController) unfollowUser(c *fiber.Ctx) error {
	var req followUserRequest
	c.BodyParser(&req)

	err := validate.Struct(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(followUserResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(followUserResponse{
			Message: "something went wrong",
			Success: false,
		})
	}
	message, err := u.storage.unfollow(userName, req.UserName, c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(followUserResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(followUserResponse{
		Message: message,
		Success: true,
	})
}

type getFollowersResponse struct {
	Data    []follower `json:"data"`
	Message string     `json:"message"`
	Success bool       `json:"success"`
}

func (u *UserController) getFollowers(c *fiber.Ctx) error {
	localData := c.Locals("userName")
	userName, cnvErr := localData.(string)

	if !cnvErr {
		return c.Status(fiber.StatusInternalServerError).JSON(getFollowersResponse{
			Message: "something went wrong",
			Success: false,
		})
	}

	result, err := u.storage.followers(userName, c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(getFollowersResponse{
			Message: err.Error(),
			Success: false,
		})
	}

	jsonData, _ := json.Marshal(result)
	var structData []follower
	json.Unmarshal(jsonData, &structData)

	return c.Status(fiber.StatusOK).JSON(getFollowersResponse{
		Data:    structData,
		Message: "found successfully",
		Success: true,
	})
}
