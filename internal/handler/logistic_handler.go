package handler

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
)

type LogisticHandler struct {
	svc service.CatalogService
}

func NewLogisticHandler(router fiber.Router, svc service.CatalogService) {
	h := &LogisticHandler{svc: svc}
	router.Get("/api/logistics", h.GetLogistics)
}

// @Summary     Get available logistics
// @Description Retrieve all logistics providers and their shipping service options with pricing
// @Tags        Catalog
// @Produce     json
// @Success     200 {object} response.WebResponse{data=[]entity.Logistic} "Logistic list"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/logistics [get]
func (h *LogisticHandler) GetLogistics(c *fiber.Ctx) error {
	logistics, err := h.svc.GetLogistics()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data logistik",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   logistics,
	})
}
