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

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(router fiber.Router, cartService service.CartService, cfg *config.Config) {
	handler := &CartHandler{cartService}
	cart := router.Group("/api/cart", middleware.Protected(cfg))
	cart.Get("/", handler.GetCart)
	cart.Post("/", handler.AddToCart)
	cart.Put("/:id", handler.UpdateQuantity)
	cart.Delete("/:id", handler.RemoveItem)
}

type AddToCartRequest struct {
	VariantID uint `json:"variant_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,min=1"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

// @Summary Get user cart
// @Description Retrieve the authenticated user's cart items with subtotal per item and grand total
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.WebResponse{data=object{items=[]service.CartItemResponse,grand_total=number}} "Cart items"
// @Failure 500 {object} response.WebResponse "Internal server error"
// @Router /api/cart [get]
func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	items, grandTotal, err := h.cartService.GetCart(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data keranjang",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data: fiber.Map{
			"items":       items,
			"grand_total": grandTotal,
		},
	})
}

// @Summary Add item to cart
// @Description Add a product variant to cart or increment quantity if it already exists
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToCartRequest true "Variant and quantity"
// @Success 201 {object} response.WebResponse{data=string} "Item added"
// @Failure 400 {object} response.WebResponse "Validation error or variant not found"
// @Router /api/cart [post]
func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var req AddToCartRequest
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

	if err := h.cartService.AddToCart(userID, req.VariantID, req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.WebResponse{
		Code:   fiber.StatusCreated,
		Status: "Created",
		Data:   "Item berhasil ditambahkan ke keranjang",
	})
}

// @Summary Update cart item quantity
// @Description Update the exact quantity of a specific cart item
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Param request body UpdateCartRequest true "New quantity"
// @Success 200 {object} response.WebResponse{data=string} "Cart updated"
// @Failure 400 {object} response.WebResponse "Validation error or item not found"
// @Router /api/cart/{id} [put]
func (h *CartHandler) UpdateQuantity(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	cartID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID tidak valid",
		})
	}

	var req UpdateCartRequest
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

	if err := h.cartService.UpdateQuantity(uint(cartID), userID, req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Keranjang berhasil diperbarui",
	})
}

// @Summary Remove item from cart
// @Description Delete a specific item from the user's cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Success 200 {object} response.WebResponse{data=string} "Item removed"
// @Failure 400 {object} response.WebResponse "Invalid ID or item not found"
// @Router /api/cart/{id} [delete]
func (h *CartHandler) RemoveItem(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	cartID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID tidak valid",
		})
	}

	if err := h.cartService.RemoveItem(uint(cartID), userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   "Item berhasil dihapus dari keranjang",
	})
}
