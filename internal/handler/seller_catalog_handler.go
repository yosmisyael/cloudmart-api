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

type SellerCatalogHandler struct {
	svc service.SellerCatalogService
}

type CategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type CreateProductRequest struct {
	CategoryID  uint   `json:"category_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description"`
}

type UpdateProductRequest struct {
	CategoryID  uint   `json:"category_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description"`
}

type VariantRequest struct {
	SKU   string  `json:"sku" validate:"required"`
	Color string  `json:"color" validate:"required"`
	Size  string  `json:"size" validate:"required"`
	Price float64 `json:"price" validate:"required,gt=0"`
	Stock int     `json:"stock" validate:"gte=0"`
}

func NewSellerCatalogHandler(router fiber.Router, svc service.SellerCatalogService, userRepo repository.UserRepository, cfg *config.Config) {
	h := &SellerCatalogHandler{svc: svc}
	seller := router.Group("/api/seller", middleware.SellerOnly(userRepo, cfg))

	seller.Get("/categories", h.GetMyCategories)
	seller.Post("/categories", h.CreateCategory)
	seller.Put("/categories/:id", h.UpdateCategory)
	seller.Delete("/categories/:id", h.DeleteCategory)

	seller.Get("/products", h.GetMyProducts)
	seller.Post("/products", h.CreateProduct)
	seller.Put("/products/:id", h.UpdateProduct)
	seller.Delete("/products/:id", h.DeleteProduct)

	seller.Get("/products/:id/variants", h.GetVariants)
	seller.Post("/products/:id/variants", h.CreateVariant)
	seller.Put("/variants/:id", h.UpdateVariant)
	seller.Delete("/variants/:id", h.DeleteVariant)
}

// @Summary     Get seller categories
// @Description Retrieve all default categories plus the seller's own custom categories
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Category} "Category list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/categories [get]
func (h *SellerCatalogHandler) GetMyCategories(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	categories, err := h.svc.GetMyCategories(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: categories,
	})
}

// @Summary     Create category
// @Description Create a new custom category owned by the authenticated seller
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CategoryRequest true "Category payload"
// @Success     201 {object} response.WebResponse{data=entity.Category} "Category created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Router      /api/seller/categories [post]
func (h *SellerCatalogHandler) CreateCategory(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req CategoryRequest
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

	category, err := h.svc.CreateCategory(userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: category,
	})
}

// @Summary     Update category
// @Description Update a custom category owned by the authenticated seller
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     int             true "Category ID"
// @Param       request body CategoryRequest true "Category payload"
// @Success     200 {object} response.WebResponse{data=entity.Category} "Category updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden — default category or not owned"
// @Failure     404 {object} response.WebResponse "Category not found"
// @Router      /api/seller/categories/{id} [put]
func (h *SellerCatalogHandler) UpdateCategory(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req CategoryRequest
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

	category, err := h.svc.UpdateCategory(userID, uint(id), req.Name)
	if err != nil {
		if err.Error() == "kategori tidak ditemukan" {
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
		Code: fiber.StatusOK, Status: "OK", Data: category,
	})
}

// @Summary     Delete category
// @Description Delete a custom category owned by the authenticated seller
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Category ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden — default category or not owned"
// @Failure     404 {object} response.WebResponse "Category not found"
// @Router      /api/seller/categories/{id} [delete]
func (h *SellerCatalogHandler) DeleteCategory(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteCategory(userID, uint(id))
	if err != nil {
		if err.Error() == "kategori tidak ditemukan" {
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

// @Summary     Get seller products
// @Description Retrieve all products belonging to the authenticated seller's store
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Product} "Product list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/products [get]
func (h *SellerCatalogHandler) GetMyProducts(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	products, err := h.svc.GetMyProducts(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code: fiber.StatusOK, Status: "OK", Data: products,
	})
}

// @Summary     Create product
// @Description Create a new product in the authenticated seller's store
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CreateProductRequest true "Product payload"
// @Success     201 {object} response.WebResponse{data=entity.Product} "Product created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Router      /api/seller/products [post]
func (h *SellerCatalogHandler) CreateProduct(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req CreateProductRequest
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

	product, err := h.svc.CreateProduct(userID, service.CreateProductInput{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err.Error() == "toko tidak ditemukan" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: product,
	})
}

// @Summary     Update product
// @Description Update a product owned by the authenticated seller
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                true "Product ID"
// @Param       request body UpdateProductRequest true "Product payload"
// @Success     200 {object} response.WebResponse{data=entity.Product} "Product updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Product not found"
// @Router      /api/seller/products/{id} [put]
func (h *SellerCatalogHandler) UpdateProduct(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req UpdateProductRequest
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

	product, err := h.svc.UpdateProduct(userID, uint(id), service.UpdateProductInput{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
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
		Code: fiber.StatusOK, Status: "OK", Data: product,
	})
}

// @Summary     Delete product
// @Description Delete a product owned by the authenticated seller
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Product ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Product not found"
// @Router      /api/seller/products/{id} [delete]
func (h *SellerCatalogHandler) DeleteProduct(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteProduct(userID, uint(id))
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
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

// @Summary     Get product variants
// @Description Retrieve all variants for a product owned by the authenticated seller
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Product ID"
// @Success     200 {object} response.WebResponse{data=[]entity.ProductVariant} "Variant list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Product not found"
// @Router      /api/seller/products/{id}/variants [get]
func (h *SellerCatalogHandler) GetVariants(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	variants, err := h.svc.GetVariants(userID, uint(id))
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
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
		Code: fiber.StatusOK, Status: "OK", Data: variants,
	})
}

// @Summary     Create variant
// @Description Add a new variant to a product owned by the authenticated seller
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int           true "Product ID"
// @Param       request body VariantRequest true "Variant payload"
// @Success     201 {object} response.WebResponse{data=entity.ProductVariant} "Variant created"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Product not found"
// @Failure     409 {object} response.WebResponse "SKU already exists"
// @Router      /api/seller/products/{id}/variants [post]
func (h *SellerCatalogHandler) CreateVariant(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req VariantRequest
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

	variant, err := h.svc.CreateVariant(userID, uint(id), service.VariantInput{
		SKU: req.SKU, Color: req.Color, Size: req.Size, Price: req.Price, Stock: req.Stock,
	})
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(response.WebResponse{
				Code: fiber.StatusNotFound, Status: "Not Found", Errors: err.Error(),
			})
		}
		if err.Error() == "akses ditolak" {
			return c.Status(fiber.StatusForbidden).JSON(response.WebResponse{
				Code: fiber.StatusForbidden, Status: "Forbidden", Errors: err.Error(),
			})
		}
		if err.Error() == "SKU sudah digunakan" {
			return c.Status(fiber.StatusConflict).JSON(response.WebResponse{
				Code: fiber.StatusConflict, Status: "Conflict", Errors: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code: fiber.StatusInternalServerError, Status: "Internal Server Error", Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code: fiber.StatusCreated, Status: "Created", Data: variant,
	})
}

// @Summary     Update variant
// @Description Update a variant belonging to the authenticated seller's product
// @Tags        Seller - Catalog
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int           true "Variant ID"
// @Param       request body VariantRequest true "Variant payload"
// @Success     200 {object} response.WebResponse{data=entity.ProductVariant} "Variant updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Variant not found"
// @Router      /api/seller/variants/{id} [put]
func (h *SellerCatalogHandler) UpdateVariant(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	var req VariantRequest
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

	variant, err := h.svc.UpdateVariant(userID, uint(id), service.VariantInput{
		SKU: req.SKU, Color: req.Color, Size: req.Size, Price: req.Price, Stock: req.Stock,
	})
	if err != nil {
		if err.Error() == "varian tidak ditemukan" {
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
		Code: fiber.StatusOK, Status: "OK", Data: variant,
	})
}

// @Summary     Delete variant
// @Description Delete a variant belonging to the authenticated seller's product
// @Tags        Seller - Catalog
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Variant ID"
// @Success     200 {object} response.WebResponse "Deleted successfully"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     404 {object} response.WebResponse "Variant not found"
// @Router      /api/seller/variants/{id} [delete]
func (h *SellerCatalogHandler) DeleteVariant(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code: fiber.StatusBadRequest, Status: "Bad Request", Errors: "ID tidak valid",
		})
	}

	err = h.svc.DeleteVariant(userID, uint(id))
	if err != nil {
		if err.Error() == "varian tidak ditemukan" {
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
