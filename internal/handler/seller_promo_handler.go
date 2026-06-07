package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
)

type SellerPromoHandler struct {
	svc service.SellerPromoService
}

type VoucherRequest struct {
	Name      string    `json:"name" validate:"required"`
	Type      string    `json:"type" validate:"required,oneof=percentage price free_shipping"`
	Amount    float64   `json:"amount" validate:"required,gt=0"`
	Max       float64   `json:"max"`
	ExpiredAt time.Time `json:"expired_at" validate:"required"`
}

type PaymentConfigRequest struct {
	Name string `json:"name" validate:"required"`
}

type BankRequest struct {
	Name        string `json:"name" validate:"required"`
	AccountID   string `json:"account_id" validate:"required"`
	AccountName string `json:"account_name" validate:"required"`
}

func NewSellerPromoHandler(router fiber.Router, svc service.SellerPromoService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerPromoHandler{svc: svc}
	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))

	seller.Get("/vouchers", h.GetVouchers)
	seller.Post("/vouchers", h.CreateVoucher)
	seller.Put("/vouchers/:id", h.UpdateVoucher)
	seller.Delete("/vouchers/:id", h.DeleteVoucher)

	seller.Get("/payment-configs", h.GetPaymentConfigs)
	seller.Post("/payment-configs", h.CreatePaymentConfig)
	seller.Delete("/payment-configs/:id", h.DeletePaymentConfig)
	seller.Post("/payment-configs/:id/banks", h.AddBank)
	seller.Delete("/banks/:id", h.DeleteBank)
}

// @Summary     Get vouchers
// @Description Retrieve all vouchers belonging to the authenticated seller's store
// @Tags        Seller - Promo
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Voucher} "Voucher list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/vouchers [get]
func (h *SellerPromoHandler) GetVouchers(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	vouchers, err := h.svc.GetVouchers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: vouchers,
	})
}

// @Summary     Create voucher
// @Description Create a new voucher for the authenticated seller's store
// @Tags        Seller - Promo
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body VoucherRequest true "Voucher payload"
// @Success     201 {object} response.WebResponse{data=entity.Voucher} "Voucher created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Router      /api/seller/vouchers [post]
func (h *SellerPromoHandler) CreateVoucher(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req VoucherRequest
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

	voucher, err := h.svc.CreateVoucher(userID, service.VoucherInput{
		Code: req.Name, Name: req.Name, Type: req.Type, Amount: req.Amount, Max: req.Max, ExpiredAt: req.ExpiredAt,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: voucher,
	})
}

// @Summary     Update voucher
// @Description Update a voucher owned by the authenticated seller
// @Tags        Seller - Promo
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int            true "Voucher ID"
// @Param       request body VoucherRequest true "Voucher payload"
// @Success     200 {object} response.WebResponse{data=entity.Voucher} "Voucher updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Voucher not found"
// @Router      /api/seller/vouchers/{id} [put]
func (h *SellerPromoHandler) UpdateVoucher(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req VoucherRequest
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

	voucher, err := h.svc.UpdateVoucher(userID, uint(id), service.VoucherInput{
		Code: req.Name, Name: req.Name, Type: req.Type, Amount: req.Amount, Max: req.Max, ExpiredAt: req.ExpiredAt,
	})
	if err != nil {
		if err.Error() == "voucher tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: voucher,
	})
}

// @Summary     Delete voucher
// @Description Delete a voucher owned by the authenticated seller
// @Tags        Seller - Promo
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Voucher ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Voucher not found"
// @Router      /api/seller/vouchers/{id} [delete]
func (h *SellerPromoHandler) DeleteVoucher(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteVoucher(userID, uint(id))
	if err != nil {
		if err.Error() == "voucher tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
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

// @Summary     Get payment configurations
// @Description Retrieve all payment configurations for the authenticated seller's store
// @Tags        Seller - Promo
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.PaymentConfiguration} "Payment config list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/payment-configs [get]
func (h *SellerPromoHandler) GetPaymentConfigs(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	configs, err := h.svc.GetPaymentConfigs(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: configs,
	})
}

// @Summary     Create payment configuration
// @Description Create a new payment configuration for the authenticated seller's store
// @Tags        Seller - Promo
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body PaymentConfigRequest true "Payment config payload"
// @Success     201 {object} response.WebResponse{data=entity.PaymentConfiguration} "Payment config created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Router      /api/seller/payment-configs [post]
func (h *SellerPromoHandler) CreatePaymentConfig(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req PaymentConfigRequest
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

	config, err := h.svc.CreatePaymentConfig(userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: config,
	})
}

// @Summary     Delete payment configuration
// @Description Delete a payment configuration owned by the authenticated seller
// @Tags        Seller - Promo
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Payment config ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Payment config not found"
// @Router      /api/seller/payment-configs/{id} [delete]
func (h *SellerPromoHandler) DeletePaymentConfig(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeletePaymentConfig(userID, uint(id))
	if err != nil {
		if err.Error() == "konfigurasi pembayaran tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
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

// @Summary     Add bank to payment configuration
// @Description Add a bank account to an existing payment configuration
// @Tags        Seller - Promo
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int         true "Payment config ID"
// @Param       request body BankRequest true "Bank payload"
// @Success     201 {object} response.WebResponse{data=entity.PaymentBank} "Bank added"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Payment config not found"
// @Router      /api/seller/payment-configs/{id}/banks [post]
func (h *SellerPromoHandler) AddBank(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req BankRequest
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

	bank, err := h.svc.AddBank(userID, uint(id), service.BankInput{
		Name: req.Name, AccountID: req.AccountID, AccountName: req.AccountName,
	})
	if err != nil {
		if err.Error() == "konfigurasi pembayaran tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: bank,
	})
}

// @Summary     Delete bank
// @Description Delete a bank account from a payment configuration
// @Tags        Seller - Promo
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Bank ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Bank not found"
// @Router      /api/seller/banks/{id} [delete]
func (h *SellerPromoHandler) DeleteBank(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteBank(userID, uint(id))
	if err != nil {
		if err.Error() == "bank tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
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
