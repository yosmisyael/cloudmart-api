package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
)

type SellerDashboardHandler struct {
	svc service.SellerDashboardService
}

func NewSellerDashboardHandler(router fiber.Router, svc service.SellerDashboardService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerDashboardHandler{svc: svc}
	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))

	seller.Get("/dashboard", h.GetSummary)
}

// @Summary     Get seller dashboard summary
// @Description Retrieve aggregated statistics for the authenticated seller's store
// @Tags        Seller - Dashboard
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=service.DashboardSummary} "Dashboard summary"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/dashboard [get]
func (h *SellerDashboardHandler) GetSummary(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	summary, err := h.svc.GetSummary(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: summary,
	})
}
