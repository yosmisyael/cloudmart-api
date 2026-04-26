package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type CatalogHandler struct {
	catalogService service.CatalogService
}

func NewCatalogHandler(router fiber.Router, catalogService service.CatalogService) {
	handler := &CatalogHandler{catalogService}
	router.Get("/api/categories", handler.GetCategories)
	router.Get("/api/products", handler.GetProducts)
	router.Get("/api/products/:id", handler.GetProductByID)
}

// @Summary Get all categories
// @Description Retrieve a list of all product categories
// @Tags Catalog
// @Produce json
// @Success 200 {object} response.WebResponse{data=[]entity.Category} "List of categories"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/categories [get]
func (h *CatalogHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.catalogService.GetCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data kategori",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   categories,
	})
}

// @Summary Get products with pagination
// @Description Retrieve a paginated list of products with optional filtering by category and search keyword
// @Tags Catalog
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param category_id query int false "Filter by category ID"
// @Param search query string false "Search by product name"
// @Success 200 {object} response.WebResponse{data=object{products=[]entity.Product,total=int,page=int,limit=int}} "Paginated product list"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/products [get]
func (h *CatalogHandler) GetProducts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id", "0"), 10, 64)
	search := c.Query("search", "")

	products, total, err := h.catalogService.GetProducts(page, limit, uint(categoryID), search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data produk",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data: fiber.Map{
			"products": products,
			"total":    total,
			"page":     page,
			"limit":    limit,
		},
	})
}

// @Summary Get product detail
// @Description Retrieve detailed product information including variants, store, and category
// @Tags Catalog
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.WebResponse{data=entity.Product} "Product detail"
// @Failure 400 {object} response.WebResponse "Invalid product ID"
// @Failure 404 {object} response.WebResponse "Product not found"
// @Router /api/products/{id} [get]
func (h *CatalogHandler) GetProductByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID produk tidak valid",
		})
	}

	product, err := h.catalogService.GetProductByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
			Code:   fiber.StatusNotFound,
			Status: "Not Found",
			Errors: "Produk tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   product,
	})
}
