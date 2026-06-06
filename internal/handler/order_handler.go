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
	orders.Post("/estimate", handler.EstimateOrder)
	orders.Post("/:id/pay", handler.InitiatePayment)
	orders.Post("/:id/cancel", handler.CancelOrder)
	orders.Get("/", handler.GetOrders)
	orders.Get("/:id", handler.GetOrderByID)
}

type CheckoutRequest struct {
	AddressID         *uint  `json:"address_id"`
	Address           string `json:"address"`
	LogisticServiceID uint   `json:"logistic_service_id" validate:"required"`
	VoucherCode       string `json:"voucher_code"`
	CartItemIDs       []uint `json:"cart_item_ids"`
}

// @Summary     Checkout order
// @Description Create an order from cart with shipping selection, optional voucher, and full cost calculation
// @Tags        Order
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CheckoutRequest true "Checkout payload"
// @Success     201 {object} response.WebResponse{data=entity.Order} "Order created"
// @Failure     400 {object} response.WebResponse "Validation error or invalid logistic"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Address or logistic not found"
// @Failure     409 {object} response.WebResponse "Stock insufficient or voucher invalid"
// @Router      /api/orders/checkout [post]
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

	order, err := h.orderService.Checkout(userID, req.AddressID, req.Address, req.LogisticServiceID, req.VoucherCode, req.CartItemIDs)
	if err != nil {
		status := fiber.StatusConflict
		if err.Error() == "alamat pengiriman harus diisi" || err.Error() == "alamat minimal 10 karakter" {
			status = fiber.StatusBadRequest
		} else if err.Error() == "layanan logistik tidak ditemukan" || err.Error() == "alamat tidak ditemukan atau bukan milik Anda" || err.Error() == "voucher tidak ditemukan" {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(response.WebResponse{
			Code:   status,
			Status: "Error",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code:   fiber.StatusCreated,
		Status: "Created",
		Data:   order,
	})
}

// @Summary     Estimate order cost
// @Description Preview the full cost breakdown for cart contents with a selected logistic and optional voucher, without creating an order
// @Tags        Order
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CheckoutRequest true "Estimate payload"
// @Success     200 {object} response.WebResponse{data=service.OrderEstimate} "Cost estimate"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Logistic or voucher not found"
// @Router      /api/orders/estimate [post]
func (h *OrderHandler) EstimateOrder(c *fiber.Ctx) error {
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

	estimate, err := h.orderService.EstimateOrder(userID, req.LogisticServiceID, req.VoucherCode, req.CartItemIDs)
	if err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "layanan logistik tidak ditemukan" || err.Error() == "voucher tidak ditemukan" {
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
		Data:   estimate,
	})
}

// @Summary     Cancel order
// @Description Cancel a pending order and restock all variants atomically
// @Tags        Order
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Order ID"
// @Success     200 {object} response.WebResponse "Order cancelled"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Order cannot be cancelled — not pending"
// @Failure     404 {object} response.WebResponse "Order not found"
// @Router      /api/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID pesanan tidak valid",
		})
	}

	if err := h.orderService.CancelOrder(userID, uint(id)); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "pesanan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if err.Error() == "pesanan tidak dapat dibatalkan karena tidak dalam status pending" {
			status = fiber.StatusForbidden
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
		Data:   "Pesanan berhasil dibatalkan",
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

// @Summary     Initiate payment
// @Description Create a Midtrans Snap transaction for a pending order and return the payment token and URL
// @Tags        Order
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Order ID"
// @Success     200 {object} response.WebResponse{data=object{snap_token=string,payment_url=string}} "Payment initiated"
// @Failure     400 {object} response.WebResponse "Invalid order ID"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Order not found"
// @Failure     409 {object} response.WebResponse "Payment already processed"
// @Failure     500 {object} response.WebResponse "Payment gateway error"
// @Router      /api/orders/{id}/pay [post]
func (h *OrderHandler) InitiatePayment(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID pesanan tidak valid",
		})
	}

	order, err := h.orderService.InitiatePayment(userID, uint(id))
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "pembayaran sudah diproses" {
			status = fiber.StatusConflict
		} else if err.Error() == "pesanan tidak ditemukan" {
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
		Data: fiber.Map{
			"snap_token":  order.SnapToken,
			"payment_url": order.PaymentURL,
		},
	})
}
