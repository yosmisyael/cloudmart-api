package repository

import (
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *entity.Review) error
	FindByProductID(productID uint) ([]entity.Review, error)
	FindByID(id uint) (*entity.Review, error)
	FindByOrderItemID(orderItemID uint) (*entity.Review, error)
	AddImage(img *entity.ReviewImage) error
	Reply(reviewID uint, replyText string, repliedAt time.Time) error
	FindByStoreID(storeID uint) ([]entity.Review, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db}
}

func (r *reviewRepository) Create(review *entity.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) FindByProductID(productID uint) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.Where("product_id = ?", productID).Preload("Images").Preload("User").Preload("OrderItem").Preload("OrderItem.Variant").Preload("OrderItem.Variant.Product").Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) FindByID(id uint) (*entity.Review, error) {
	var review entity.Review
	err := r.db.First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) FindByOrderItemID(orderItemID uint) (*entity.Review, error) {
	var review entity.Review
	err := r.db.Where("order_item_id = ?", orderItemID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) AddImage(img *entity.ReviewImage) error {
	return r.db.Create(img).Error
}

func (r *reviewRepository) Reply(reviewID uint, replyText string, repliedAt time.Time) error {
	return r.db.Model(&entity.Review{}).Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"reply_text": replyText,
			"replied_at": repliedAt,
		}).Error
}

func (r *reviewRepository) FindByStoreID(storeID uint) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.Joins("JOIN products ON products.id = reviews.product_id").
		Where("products.store_id = ?", storeID).
		Preload("Images").
		Preload("User").
		Preload("OrderItem").
		Preload("OrderItem.Variant").
		Preload("OrderItem.Variant.Product").
		Find(&reviews).Error
	return reviews, err
}
