package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"

)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(router fiber.Router, orderService service.OrderService, cfg *config.Config) {
	handler := &OrderHandler{orderService: orderService}
	orders := router.Group("/api/orders", middleware.Protected(cfg))
	orders.Post("/checkout", handler.Checkout)
	orders.Get("/", handler.GetOrders)
	orders.Get("/:id", handler.GetOrderByID)
}

type CheckoutRequest struct {
	Address string `json:"address" validate:"required,min=10"`
}

// @Summary Checkout order
// @Description Create an order from the user's cart, validate stock, decrement inventory, and clear cart atomically
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CheckoutRequest true "Shipping address"
// @Success 201 {object} response.WebResponse{data=entity.Order} "Order created"
// @Failure 400 {object} response.WebResponse "Validation error"
// @Failure 409 {object} response.WebResponse "Stock insufficient or transaction failed"
// @Router /api/orders/checkout [post]
func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req CheckoutRequest
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

	order, err := h.orderService.Checkout(userID, req.Address)
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
		Data:   order,
	})
}

// @Summary Get order history
// @Description Retrieve the authenticated user's order history
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.WebResponse{data=[]entity.Order} "Order list"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/orders [get]
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	orders, err := h.orderService.GetOrders(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data pesanan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   orders,
	})
}

// @Summary Get order detail
// @Description Retrieve a specific order and its associated items
// @Tags Order
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} response.WebResponse{data=entity.Order} "Order detail"
// @Failure 400 {object} response.WebResponse "Invalid order ID"
// @Failure 404 {object} response.WebResponse "Order not found"
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID pesanan tidak valid",
		})
	}

	order, err := h.orderService.GetOrderByID(uint(id), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
			Code:   fiber.StatusNotFound,
			Status: "Not Found",
			Errors: "Pesanan tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   order,
	})
}
