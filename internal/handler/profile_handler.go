package handler

import (
	"io"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/upload"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
)

type ProfileHandler struct {
	userService service.UserService
}

func NewProfileHandler(router fiber.Router, userService service.UserService, cfg *config.Config) {
	handler := &ProfileHandler{userService}
	profile := router.Group("/api/profile", middleware.Protected(cfg))
	profile.Get("/", handler.GetProfile)
	profile.Put("/", handler.UpdateProfile)
	profile.Put("/password", handler.ChangePassword)
	profile.Get("/addresses", handler.GetAddresses)
	profile.Post("/addresses", handler.CreateAddress)
	profile.Put("/addresses/:id", handler.UpdateAddress)
	profile.Delete("/addresses/:id", handler.DeleteAddress)
	profile.Post("/addresses/:id/default", handler.SetDefaultAddress)
	profile.Post("/avatar", handler.UploadAvatar)
	profile.Delete("/avatar", handler.DeleteAvatar)
}

type ProfileResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
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

type UpdateAddressRequest struct {
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

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Phone string `json:"phone" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
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
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		AvatarURL: user.AvatarURL,
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

// @Summary     Update profile
// @Description Update the authenticated user's name and phone number
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body UpdateProfileRequest true "Profile payload"
// @Success     200 {object} response.WebResponse{data=entity.User} "Profile updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/profile [put]
func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req UpdateProfileRequest
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

	if err := h.userService.UpdateProfile(userID, req.Name, req.Phone); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengupdate profil",
		})
	}

	user, _ := h.userService.GetProfile(userID)
	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   user,
	})
}

// @Summary     Change password
// @Description Change the authenticated user's password
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body ChangePasswordRequest true "Password payload"
// @Success     200 {object} response.WebResponse "Password changed"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Wrong current password"
// @Failure     404 {object} response.WebResponse "User not found"
// @Router      /api/profile/password [put]
func (h *ProfileHandler) ChangePassword(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req ChangePasswordRequest
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

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "password saat ini salah" {
			status = fiber.StatusUnauthorized
		} else if err.Error() == "user tidak ditemukan" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(response.WebResponse{
			Code:   status,
			Status: "Error",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Password berhasil diubah",
	})
}

// @Summary     Update address
// @Description Update a saved shipping address for the authenticated user
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                  true "Address ID"
// @Param       request body UpdateAddressRequest true "Address payload"
// @Success     200 {object} response.WebResponse{data=entity.Address} "Address updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id} [put]
func (h *ProfileHandler) UpdateAddress(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	addressID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID alamat tidak valid",
		})
	}

	var req UpdateAddressRequest
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
		ID:                    uint(addressID),
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

	if err := h.userService.UpdateAddress(address); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengupdate alamat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   address,
	})
}

// @Summary     Delete address
// @Description Delete a saved shipping address for the authenticated user
// @Tags        Profile
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Address ID"
// @Success     200 {object} response.WebResponse "Address deleted"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id} [delete]
func (h *ProfileHandler) DeleteAddress(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	addressID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID alamat tidak valid",
		})
	}

	if err := h.userService.DeleteAddress(uint(addressID), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal menghapus alamat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Alamat berhasil dihapus",
	})
}

// @Summary     Set default address
// @Description Mark an address as the user's default shipping address
// @Tags        Profile
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Address ID"
// @Success     200 {object} response.WebResponse "Default address set"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id}/default [post]
func (h *ProfileHandler) SetDefaultAddress(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	addressID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID alamat tidak valid",
		})
	}

	if err := h.userService.SetDefaultAddress(uint(addressID), userID); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "record not found" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(response.WebResponse{
			Code:   status,
			Status: "Error",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Alamat utama berhasil diatur",
	})
}

// @Summary     Upload avatar
// @Description Upload or replace the profile avatar for the authenticated user
// @Tags        Profile
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       image formData file true "Avatar image (jpg/png, max 3MB)"
// @Success     200 {object} response.WebResponse{data=object{url=string}} "Avatar uploaded"
// @Failure     400 {object} response.WebResponse "Invalid file"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     500 {object} response.WebResponse "Upload failed"
// @Router      /api/profile/avatar [post]
func (h *ProfileHandler) UploadAvatar(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "File gambar diperlukan",
		})
	}

	if err := upload.ValidateImage(file, 3); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: err.Error(),
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: "Gagal membuka file",
		})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: "Gagal membaca file",
		})
	}

	filename := upload.GenerateFilename(file.Filename)
	contentType := file.Header.Get("Content-Type")

	url, err := h.userService.UploadAvatar(c.Context(), userID, data, filename, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: fiber.Map{"url": url},
	})
}

// @Summary     Delete avatar
// @Description Remove the profile avatar for the authenticated user
// @Tags        Profile
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse "Avatar deleted"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/profile/avatar [delete]
func (h *ProfileHandler) DeleteAvatar(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	if err := h.userService.DeleteAvatar(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK",
	})
}
