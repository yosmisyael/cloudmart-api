package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
)

type SellerLogisticHandler struct {
	svc service.SellerLogisticService
}

type LogisticRequest struct {
	Name string `json:"name" validate:"required"`
}

type LogisticServiceRequest struct {
	Name      string  `json:"name" validate:"required"`
	BasePrice float64 `json:"base_price" validate:"required,gt=0"`
}

func NewSellerLogisticHandler(router fiber.Router, svc service.SellerLogisticService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerLogisticHandler{svc: svc}
	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))

	seller.Get("/logistics", h.GetLogistics)
	seller.Post("/logistics", h.CreateLogistic)
	seller.Put("/logistics/:id", h.UpdateLogistic)
	seller.Delete("/logistics/:id", h.DeleteLogistic)
	seller.Post("/logistics/:id/services", h.AddService)
	seller.Put("/logistics/services/:id", h.UpdateService)
	seller.Delete("/logistics/services/:id", h.DeleteService)
}

// @Summary     Get logistics
// @Description Retrieve all logistics providers with their services
// @Tags        Seller - Logistic
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Logistic} "Logistic list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/logistics [get]
func (h *SellerLogisticHandler) GetLogistics(c *fiber.Ctx) error {
	logistics, err := h.svc.GetLogistics()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: logistics,
	})
}

// @Summary     Create logistic
// @Description Create a new logistics provider
// @Tags        Seller - Logistic
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body LogisticRequest true "Logistic payload"
// @Success     201 {object} response.WebResponse{data=entity.Logistic} "Logistic created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Router      /api/seller/logistics [post]
func (h *SellerLogisticHandler) CreateLogistic(c *fiber.Ctx) error {
	var req LogisticRequest
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

	logistic, err := h.svc.CreateLogistic(req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: logistic,
	})
}

// @Summary     Update logistic
// @Description Update a logistics provider's name
// @Tags        Seller - Logistic
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int             true "Logistic ID"
// @Param       request body LogisticRequest true "Logistic payload"
// @Success     200 {object} response.WebResponse{data=entity.Logistic} "Logistic updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Logistic not found"
// @Router      /api/seller/logistics/{id} [put]
func (h *SellerLogisticHandler) UpdateLogistic(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req LogisticRequest
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

	logistic, err := h.svc.UpdateLogistic(uint(id), req.Name)
	if err != nil {
		if err.Error() == "logistik tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: logistic,
	})
}

// @Summary     Delete logistic
// @Description Delete a logistics provider
// @Tags        Seller - Logistic
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Logistic ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Logistic not found"
// @Router      /api/seller/logistics/{id} [delete]
func (h *SellerLogisticHandler) DeleteLogistic(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteLogistic(uint(id))
	if err != nil {
		if err.Error() == "logistik tidak ditemukan" {
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

// @Summary     Add logistic service
// @Description Add a shipping service tier to an existing logistics provider
// @Tags        Seller - Logistic
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                    true "Logistic ID"
// @Param       request body LogisticServiceRequest true "Service payload"
// @Success     201 {object} response.WebResponse{data=entity.LogisticService} "Service added"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Logistic not found"
// @Router      /api/seller/logistics/{id}/services [post]
func (h *SellerLogisticHandler) AddService(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req LogisticServiceRequest
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

	svc, err := h.svc.AddService(uint(id), service.LogisticServiceInput{
		Name: req.Name, BasePrice: req.BasePrice,
	})
	if err != nil {
		if err.Error() == "logistik tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: svc,
	})
}

// @Summary     Update logistic service
// @Description Update a shipping service tier
// @Tags        Seller - Logistic
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                    true "Service ID"
// @Param       request body LogisticServiceRequest true "Service payload"
// @Success     200 {object} response.WebResponse{data=entity.LogisticService} "Service updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Service not found"
// @Router      /api/seller/logistics/services/{id} [put]
func (h *SellerLogisticHandler) UpdateService(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req LogisticServiceRequest
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

	svc, err := h.svc.UpdateService(uint(id), service.LogisticServiceInput{
		Name: req.Name, BasePrice: req.BasePrice,
	})
	if err != nil {
		if err.Error() == "layanan tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: svc,
	})
}

// @Summary     Delete logistic service
// @Description Delete a shipping service tier
// @Tags        Seller - Logistic
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Service ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Service not found"
// @Router      /api/seller/logistics/services/{id} [delete]
func (h *SellerLogisticHandler) DeleteService(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteService(uint(id))
	if err != nil {
		if err.Error() == "layanan tidak ditemukan" {
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
