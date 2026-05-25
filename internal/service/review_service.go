package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/pkg/upload"
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID, orderItemID uint, rating int, comment string, images []*multipart.FileHeader, videoHeader *multipart.FileHeader) (*entity.Review, error)
	GetProductReviews(productID uint) ([]entity.Review, error)
	ReplyReview(sellerUserID, reviewID uint, replyText string) error
	GetStoreReviews(sellerUserID uint) ([]entity.Review, error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	storeRepo   repository.StoreRepository
	s3Svc       S3Service
}

func NewReviewService(reviewRepo repository.ReviewRepository, orderRepo repository.OrderRepository, productRepo repository.ProductRepository, storeRepo repository.StoreRepository, s3Svc S3Service) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		storeRepo:   storeRepo,
		s3Svc:       s3Svc,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, userID, orderItemID uint, rating int, comment string, images []*multipart.FileHeader, videoHeader *multipart.FileHeader) (*entity.Review, error) {
	orderItem, err := s.orderRepo.FindOrderItemByID(orderItemID)
	if err != nil {
		return nil, errors.New("item pesanan tidak ditemukan")
	}

	if orderItem.Order.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	if orderItem.Order.PaymentStatus != "settlement" {
		return nil, errors.New("order belum selesai")
	}

	existingReview, _ := s.reviewRepo.FindByOrderItemID(orderItemID)
	if existingReview != nil {
		return nil, errors.New("ulasan sudah dikirim untuk item ini")
	}

	if rating < 1 || rating > 5 {
		return nil, errors.New("rating harus antara 1 dan 5")
	}

	reviewImages := []entity.ReviewImage{}
	for _, header := range images {
		file, err := header.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		filename := upload.GenerateFilename(header.Filename)
		contentType := header.Header.Get("Content-Type")

		url, err := s.s3Svc.UploadFile(ctx, "reviews", filename, data, contentType)
		if err == nil {
			reviewImages = append(reviewImages, entity.ReviewImage{
				ImageURL: url,
			})
		}
	}

	var videoURL string
	if videoHeader != nil {
		file, err := videoHeader.Open()
		if err == nil {
			data, err := io.ReadAll(file)
			file.Close()
			if err == nil {
				filename := upload.GenerateFilename(videoHeader.Filename)
				contentType := videoHeader.Header.Get("Content-Type")
				url, err := s.s3Svc.UploadFile(ctx, "reviews/videos", filename, data, contentType)
				if err == nil {
					videoURL = url
				}
			}
		}
	}

	review := entity.Review{
		OrderItemID: orderItemID,
		UserID:      userID,
		ProductID:   orderItem.VariantID, // note: the plan says ProductID but the item only has VariantID - I will resolve ProductID via Variant
		Rating:      rating,
		Comment:     comment,
		VideoURL:    videoURL,
		Images:      reviewImages,
	}

	// Resolve ProductID from VariantID
	variant, err := s.productRepo.FindVariantByID(orderItem.VariantID)
	if err == nil {
		review.ProductID = variant.ProductID
	}

	if err := s.reviewRepo.Create(&review); err != nil {
		return nil, errors.New("gagal menyimpan ulasan")
	}

	return &review, nil
}

func (s *reviewService) GetProductReviews(productID uint) ([]entity.Review, error) {
	return s.reviewRepo.FindByProductID(productID)
}

func (s *reviewService) ReplyReview(sellerUserID, reviewID uint, replyText string) error {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return errors.New("ulasan tidak ditemukan")
	}

	store, err := s.storeRepo.FindByUserID(sellerUserID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(review.ProductID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if review.RepliedAt != nil {
		return errors.New("ulasan sudah dibalas")
	}

	return s.reviewRepo.Reply(reviewID, replyText, time.Now())
}

func (s *reviewService) GetStoreReviews(sellerUserID uint) ([]entity.Review, error) {
	store, err := s.storeRepo.FindByUserID(sellerUserID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	return s.reviewRepo.FindByStoreID(store.ID)
}
