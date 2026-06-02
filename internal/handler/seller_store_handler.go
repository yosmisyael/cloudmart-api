package handler

import (
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/upload"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
)

type SellerStoreHandler struct {
	svc service.SellerStoreService
}

type CreateStoreRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type UpdateStoreRequest struct {
	Name      string `json:"name" validate:"required,min=3,max=100"`
	AddressID *uint  `json:"address_id"`
}

func NewSellerStoreHandler(router fiber.Router, svc service.SellerStoreService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerStoreHandler{svc: svc}

	open := router.Group("/api/seller", middleware.Protected(cfg))
	open.Post("/store", h.CreateStore)

	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))
	seller.Get("/store", h.GetStore)
	seller.Put("/store", h.UpdateStore)
	seller.Post("/store/logo", h.UploadStoreLogo)
	seller.Delete("/store/logo", h.DeleteStoreLogo)
}

// @Summary     Get seller store
// @Description Retrieve the authenticated seller's store profile
// @Tags        Seller - Store
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=entity.Store} "Store data"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Store not found"
// @Router      /api/seller/store [get]
func (h *SellerStoreHandler) GetStore(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	store, err := h.svc.GetStore(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
			Code:   fiber.StatusNotFound,
			Status: "Not Found",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   store,
	})
}

// @Summary     Create store
// @Description Register a new store for the authenticated user and upgrade their role to seller
// @Tags        Seller - Store
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CreateStoreRequest true "Store creation payload"
// @Success     201 {object} response.WebResponse{data=entity.Store} "Store created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     409 {object} response.WebResponse "Store already exists"
// @Router      /api/seller/store [post]
func (h *SellerStoreHandler) CreateStore(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req CreateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "Invalid input",
		})
	}
	if err := validator.Validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request",
			Errors: response.FormatValidationError(err),
		})
	}

	store, err := h.svc.CreateStore(userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(response.WebResponse{
			Code:   fiber.StatusConflict,
			Status: "Conflict",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code:   fiber.StatusCreated,
		Status: "Created",
		Data:   store,
	})
}

// @Summary     Update store
// @Description Update the authenticated seller's store name or address
// @Tags        Seller - Store
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body UpdateStoreRequest true "Store update payload"
// @Success     200 {object} response.WebResponse{data=entity.Store} "Store updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Store not found"
// @Router      /api/seller/store [put]
func (h *SellerStoreHandler) UpdateStore(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req UpdateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "Invalid input",
		})
	}
	if err := validator.Validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request",
			Errors: response.FormatValidationError(err),
		})
	}

	store, err := h.svc.UpdateStore(userID, req.Name, req.AddressID)
	if err != nil {
		if err.Error() == "toko tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
			Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   store,
	})
}

// @Summary     Upload store logo
// @Description Upload or replace the logo for the authenticated seller's store
// @Tags        Seller - Store
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       image formData file true "Logo image (jpg/png, max 5MB)"
// @Success     200 {object} response.WebResponse{data=object{url=string}} "Logo uploaded"
// @Failure     400 {object} response.WebResponse "Invalid file"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Upload failed"
// @Router      /api/seller/store/logo [post]
func (h *SellerStoreHandler) UploadStoreLogo(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "File gambar diperlukan",
		})
	}

	if err := upload.ValidateImage(file, 5); err != nil {
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

	url, err := h.svc.UploadStoreLogo(c.Context(), userID, data, filename, contentType)
	if err != nil {
		if err.Error() == "toko tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: fiber.Map{"url": url},
	})
}

// @Summary     Delete store logo
// @Description Remove the logo of the authenticated seller's store
// @Tags        Seller - Store
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse "Logo deleted"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Store not found"
// @Router      /api/seller/store/logo [delete]
func (h *SellerStoreHandler) DeleteStoreLogo(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	if err := h.svc.DeleteStoreLogo(c.Context(), userID); err != nil {
		if err.Error() == "toko tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK",
	})
}
