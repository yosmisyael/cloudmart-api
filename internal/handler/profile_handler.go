package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type ProfileHandler struct {
	userService service.UserService
}

func NewProfileHandler(router fiber.Router, userService service.UserService, cfg *config.Config) {
	handler := &ProfileHandler{userService}
	profile := router.Group("/api/profile", middleware.Protected(cfg))
	profile.Get("/", handler.GetProfile)
	profile.Get("/addresses", handler.GetAddresses)
	profile.Post("/addresses", handler.CreateAddress)
}

type ProfileResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

type CreateAddressRequest struct {
	Address               string  `json:"address" validate:"required"`
	City                  string  `json:"city" validate:"required"`
	State                 string  `json:"state" validate:"required"`
	Country               string  `json:"country" validate:"required"`
	PostalCode            string  `json:"postal_code" validate:"required"`
	Phone                 string  `json:"phone" validate:"required"`
	Recipient             string  `json:"recipient" validate:"required"`
	Type                  string  `json:"type" validate:"required"`
	AdditionalInformation *string `json:"additional_information"`
}

// @Summary Get user profile
// @Description Retrieve the authenticated user's profile information excluding password and refresh token
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.WebResponse{data=ProfileResponse} "User profile"
// @Failure 404 {object} response.WebResponse "User not found"
// @Router /api/profile [get]
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
			Code:   fiber.StatusNotFound,
			Status: "Not Found",
			Errors: "User tidak ditemukan",
		})
	}

	profileResp := ProfileResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  user.Role,
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   profileResp,
	})
}

// @Summary Get user addresses
// @Description Retrieve all saved shipping addresses for the authenticated user
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.WebResponse{data=[]entity.Address} "Address list"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/profile/addresses [get]
func (h *ProfileHandler) GetAddresses(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	addresses, err := h.userService.GetAddresses(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data alamat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   addresses,
	})
}

// @Summary Add a new address
// @Description Create a new shipping address for the authenticated user
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAddressRequest true "Address details"
// @Success 201 {object} response.WebResponse{data=entity.Address} "Address created"
// @Failure 400 {object} response.WebResponse "Validation error"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/profile/addresses [post]
func (h *ProfileHandler) CreateAddress(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Invalid input",
		})
	}

	if err := validator.Validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: response.FormatValidationError(err),
		})
	}

	address := &entity.Address{
		UserID:                userID,
		Address:               req.Address,
		City:                  req.City,
		State:                 req.State,
		Country:               req.Country,
		PostalCode:            req.PostalCode,
		Phone:                 req.Phone,
		Recipient:             req.Recipient,
		Type:                  req.Type,
		AdditionalInformation: req.AdditionalInformation,
	}

	if err := h.userService.CreateAddress(address); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal menyimpan alamat",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code:   fiber.StatusCreated,
		Status: "Created",
		Data:   address,
	})
}
