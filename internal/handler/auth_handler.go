package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(router fiber.Router, authService service.AuthService) {
	handler := &AuthHandler{authService}
	router.Post("/api/register", handler.Register)
	router.Post("/api/login", handler.Login)
	router.Post("/api/refresh", handler.Refresh)
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Summary Register a new user
// @Description Create a new customer account with name, email, password, and phone
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration payload"
// @Success 201 {object} response.WebResponse{data=string} "Registration successful"
// @Failure 400 {object} response.WebResponse "Validation error"
// @Failure 409 {object} response.WebResponse "Email already registered"
// @Router /api/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
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

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
	}

	if err := h.authService.Register(user); err != nil {
		return c.Status(fiber.StatusConflict).JSON(response.WebResponse{
			Code:   fiber.StatusConflict,
			Status: "Conflict",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code:   fiber.StatusCreated,
		Status: "Created",
		Data:   "Registrasi berhasil",
	})
}

// @Summary User login
// @Description Authenticate user credentials and return access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.WebResponse{data=object{access_token=string,refresh_token=string}} "Login successful"
// @Failure 400 {object} response.WebResponse "Validation error"
// @Failure 401 {object} response.WebResponse "Invalid credentials"
// @Router /api/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
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

	accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.WebResponse{
			Code:   fiber.StatusUnauthorized,
			Status: "Unauthorized",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data: fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token payload"
// @Success 200 {object} response.WebResponse{data=object{access_token=string}} "Token refreshed"
// @Failure 400 {object} response.WebResponse "Validation error"
// @Failure 401 {object} response.WebResponse "Invalid or expired refresh token"
// @Router /api/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
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

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.WebResponse{
			Code:   fiber.StatusUnauthorized,
			Status: "Unauthorized",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data: fiber.Map{
			"access_token": accessToken,
		},
	})
}
