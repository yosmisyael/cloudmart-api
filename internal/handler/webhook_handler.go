package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type WebhookHandler struct {
	webhookService service.WebhookService
}

func NewWebhookHandler(router fiber.Router, webhookService service.WebhookService) {
	handler := &WebhookHandler{webhookService}
	router.Post("/api/webhooks/midtrans", handler.MidtransNotification)
}

type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
}

// @Summary Midtrans payment notification
// @Description Receive asynchronous payment status notification from Midtrans and update order payment status
// @Tags Webhook
// @Accept json
// @Produce json
// @Param request body MidtransNotificationRequest true "Midtrans notification payload"
// @Success 200 {object} response.WebResponse{data=string} "Notification processed"
// @Failure 400 {object} response.WebResponse "Invalid input or order ID"
// @Failure 403 {object} response.WebResponse "Invalid signature"
// @Router /api/webhooks/midtrans [post]
func (h *WebhookHandler) MidtransNotification(c *fiber.Ctx) error {
	var req MidtransNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Invalid input",
		})
	}

	orderID, err := strconv.ParseUint(req.OrderID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Order ID tidak valid",
		})
	}

	if err := h.webhookService.HandleMidtransNotification(
		uint(orderID),
		req.StatusCode,
		req.GrossAmount,
		req.SignatureKey,
		req.TransactionStatus,
	); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
			Code:   fiber.StatusForbidden,
			Status: "Forbidden",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Notifikasi berhasil diproses",
	})
}
