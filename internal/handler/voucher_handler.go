package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type VoucherHandler struct {
	svc service.VoucherService
}

func NewVoucherHandler(router fiber.Router, svc service.VoucherService, cfg *config.Config) {
	h := &VoucherHandler{svc: svc}
	vouchers := router.Group("/api/vouchers", middleware.Protected(cfg))

	vouchers.Get("/", h.GetAvailableVouchers)
	vouchers.Get("/mine", h.GetMyVouchers)
	vouchers.Post("/:code/claim", h.ClaimVoucher)
}

// @Summary     Browse available vouchers
// @Description Retrieve all active, non-expired vouchers for a specific store
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Param       store_id query int true "Store ID"
// @Success     200 {object} response.WebResponse{data=[]entity.Voucher} "Voucher list"
// @Failure     400 {object} response.WebResponse "Missing store_id"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/vouchers [get]
func (h *VoucherHandler) GetAvailableVouchers(c *fiber.Ctx) error {
	storeIDStr := c.Query("store_id", "")
	if storeIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Parameter store_id wajib diisi",
		})
	}

	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "store_id tidak valid",
		})
	}

	vouchers, err := h.svc.GetAvailableVouchers(uint(storeID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   vouchers,
	})
}

// @Summary     Claim a voucher
// @Description Add a voucher to the authenticated user's account by voucher code
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Param       code path string true "Voucher code"
// @Success     200 {object} response.WebResponse "Voucher claimed"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Voucher not found or expired"
// @Failure     409 {object} response.WebResponse "Already claimed"
// @Router      /api/vouchers/{code}/claim [post]
func (h *VoucherHandler) ClaimVoucher(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	code := c.Params("code")

	err := h.svc.ClaimVoucher(userID, code)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "voucher tidak ditemukan" || err.Error() == "voucher kadaluarsa" {
			status = fiber.StatusNotFound
		} else if err.Error() == "sudah diklaim" {
			status = fiber.StatusConflict
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
		Data:   "Voucher berhasil diklaim",
	})
}

// @Summary     Get my vouchers
// @Description Retrieve all vouchers claimed by the authenticated user
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Voucher} "My voucher list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/vouchers/mine [get]
func (h *VoucherHandler) GetMyVouchers(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	vouchers, err := h.svc.GetMyVouchers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   vouchers,
	})
}
