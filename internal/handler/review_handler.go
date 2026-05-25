package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/middleware"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"
	"github.com/yosmisyael/cloudmart-web-service/pkg/response"
	"github.com/yosmisyael/cloudmart-web-service/pkg/upload"
	"github.com/yosmisyael/cloudmart-web-service/pkg/validator"
	_ "github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(router fiber.Router, reviewService service.ReviewService, userRepo repository.UserRepository, cfg *config.Config) {
	handler := &ReviewHandler{reviewService}

	// Protected routes
	protected := router.Group("/api/orders/items", middleware.Protected(cfg))
	protected.Post("/:item_id/review", handler.CreateReview)

	// Public routes
	router.Get("/api/products/:id/reviews", handler.GetProductReviews)

	// Seller routes
	seller := router.Group("/api/seller/reviews", middleware.Protected(cfg), middleware.SellerOnly(userRepo, cfg))
	seller.Put("/:id/reply", handler.ReplyReview)
	seller.Get("/", handler.GetStoreReviews)
}

// @Summary     Submit product review
// @Description Submit a review for a purchased order item. Order must be settled. One review per item.
// @Tags        Review
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       item_id path     int    true  "Order Item ID"
// @Param       rating  formData int    true  "Rating (1–5)"
// @Param       comment formData string false "Review comment"
// @Param       images  formData file   false "Product images (jpg/png, max 3MB each, max 5 files)"
// @Param       video   formData file   false "Review video (mp4/mov, max 60MB)"
// @Success     201 {object} response.WebResponse{data=entity.Review} "Review submitted"
// @Failure     400 {object} response.WebResponse "Validation error or file constraint violation"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Order not settled or item not owned"
// @Failure     409 {object} response.WebResponse "Review already submitted for this item"
// @Router      /api/orders/items/{item_id}/review [post]
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	orderItemID, err := strconv.ParseUint(c.Params("item_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID item pesanan tidak valid",
		})
	}

	rating, _ := strconv.Atoi(c.FormValue("rating"))
	comment := c.FormValue("comment")

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Gagal memproses form",
		})
	}

	imageHeaders := form.File["images"]
	if len(imageHeaders) > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "Maksimal 5 gambar",
		})
	}

	for _, header := range imageHeaders {
		if err := upload.ValidateImage(header, 3); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
				Code:   fiber.StatusBadRequest,
				Status: "Bad Request",
				Errors: err.Error(),
			})
		}
	}

	videoHeader, _ := c.FormFile("video")
	if videoHeader != nil {
		if err := upload.ValidateVideo(videoHeader, 60); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
				Code:   fiber.StatusBadRequest,
				Status: "Bad Request",
				Errors: err.Error(),
			})
		}
	}

	review, err := h.reviewService.CreateReview(c.Context(), userID, uint(orderItemID), rating, comment, imageHeaders, videoHeader)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "ulasan sudah dikirim untuk item ini" {
			status = fiber.StatusConflict
		} else if err.Error() == "akses ditolak" || err.Error() == "order belum selesai" {
			status = fiber.StatusForbidden
		} else if err.Error() == "rating harus antara 1 dan 5" {
			status = fiber.StatusBadRequest
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
		Data:   review,
	})
}

// @Summary     Get product reviews
// @Description Retrieve all reviews for a product, including images and seller replies
// @Tags        Review
// @Produce     json
// @Param       id path int true "Product ID"
// @Success     200 {object} response.WebResponse{data=[]entity.Review} "Review list"
// @Failure     404 {object} response.WebResponse "Product not found"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/products/{id}/reviews [get]
func (h *ReviewHandler) GetProductReviews(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID produk tidak valid",
		})
	}

	reviews, err := h.reviewService.GetProductReviews(uint(productID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.WebResponse{
			Code:   fiber.StatusInternalServerError,
			Status: "Internal Server Error",
			Errors: "Gagal mengambil data ulasan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.WebResponse{
		Code:   fiber.StatusOK,
		Status: "OK",
		Data:   reviews,
	})
}

type ReplyRequest struct {
	Reply string `json:"reply" validate:"required,min=1"`
}

// @Summary     Reply to a review
// @Description Seller posts a single reply to a buyer's review on their product
// @Tags        Seller - Review
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int          true "Review ID"
// @Param       request body ReplyRequest true "Reply payload"
// @Success     200 {object} response.WebResponse "Reply posted"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden — not your product"
// @Failure     404 {object} response.WebResponse "Review not found"
// @Failure     409 {object} response.WebResponse "Already replied"
// @Router      /api/seller/reviews/{id}/reply [put]
func (h *ReviewHandler) ReplyReview(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	reviewID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.WebResponse{
			Code:   fiber.StatusBadRequest,
			Status: "Bad Request",
			Errors: "ID ulasan tidak valid",
		})
	}

	var req ReplyRequest
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

	if err := h.reviewService.ReplyReview(userID, uint(reviewID), req.Reply); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "ulasan sudah dibalas" {
			status = fiber.StatusConflict
		} else if err.Error() == "akses ditolak" {
			status = fiber.StatusForbidden
		} else if err.Error() == "ulasan tidak ditemukan" {
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
		Data:   "Balasan berhasil dikirim",
	})
}

// @Summary     Get store reviews
// @Description Retrieve all reviews for all products in the authenticated seller's store
// @Tags        Seller - Review
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Review} "Review list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Forbidden"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/seller/reviews [get]
func (h *ReviewHandler) GetStoreReviews(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	reviews, err := h.reviewService.GetStoreReviews(userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "toko tidak ditemukan" {
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
		Data:   reviews,
	})
}
