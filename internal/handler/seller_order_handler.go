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

type SellerOrderHandler struct {
	svc service.SellerOrderService
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=processing shipped delivered cancelled"`
}

func NewSellerOrderHandler(router fiber.Router, svc service.SellerOrderService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerOrderHandler{svc: svc}
	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))

	seller.Get("/orders", h.GetOrders)
	seller.Get("/orders/:id", h.GetOrderByID)
	seller.Put("/orders/:id/status", h.UpdateOrderStatus)
}

// @Summary     Get seller orders
// @Description Retrieve all orders that contain items from the authenticated seller's store
// @Tags        Seller - Order
// @Produce     json
// @Security    BearerAuth
// @Param       status query string false "Filter by payment_status (pending, processing, shipped, delivered, cancelled)"
// @Success     200 {object} response.WebResponse{data=[]entity.Order} "Order list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/orders [get]
func (h *SellerOrderHandler) GetOrders(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	status := c.Query("status", "")

	orders, err := h.svc.GetOrders(userID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: orders,
	})
}

// @Summary     Get seller order detail
// @Description Retrieve a specific order detail for the authenticated seller
// @Tags        Seller - Order
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Order ID"
// @Success     200 {object} response.WebResponse{data=entity.Order} "Order detail"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Order not found"
// @Router      /api/seller/orders/{id} [get]
func (h *SellerOrderHandler) GetOrderByID(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	order, err := h.svc.GetOrderByID(userID, uint(id))
	if err != nil {
		if err.Error() == "pesanan tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: order,
	})
}

// @Summary     Update order status
// @Description Update the status of an order belonging to the authenticated seller's store
// @Tags        Seller - Order
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                      true "Order ID"
// @Param       request body UpdateOrderStatusRequest true "Status payload"
// @Success     200 {object} response.WebResponse "Status updated"
// @Failure     400 {object} response.WebResponse "Validation error or invalid status"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Order not found"
// @Router      /api/seller/orders/{id}/status [put]
func (h *SellerOrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req UpdateOrderStatusRequest
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

	err = h.svc.UpdateOrderStatus(userID, uint(id), req.Status)
	if err != nil {
		if err.Error() == "pesanan tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "status tidak valid" {
			return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
				Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: err.Error(),
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
