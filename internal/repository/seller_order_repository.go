package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type SellerOrderRepository interface {
	FindByStoreID(storeID uint, status string) ([]entity.Order, error)
	FindOrderByID(orderID uint) (*entity.Order, error)
	UpdateOrderStatus(orderID uint, status string) error
	UpdateShippingStatus(orderID uint, status string) error
}

type sellerOrderRepository struct {
	db *gorm.DB
}

func NewSellerOrderRepository(db *gorm.DB) SellerOrderRepository {
	return &sellerOrderRepository{db}
}

func (r *sellerOrderRepository) FindByStoreID(storeID uint, status string) ([]entity.Order, error) {
	var orders []entity.Order
	query := r.db.Preload("OrderItems").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
		Joins("JOIN products ON products.id = product_variants.product_id").
		Where("products.store_id = ?", storeID).
		Group("orders.id")

	if status != "" {
		query = query.Where("orders.payment_status = ?", status)
	}

	err := query.Find(&orders).Error
	return orders, err
}

func (r *sellerOrderRepository) FindOrderByID(orderID uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.Preload("OrderItems").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *sellerOrderRepository) UpdateOrderStatus(orderID uint, status string) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", orderID).Update("payment_status", status).Error
}

func (r *sellerOrderRepository) UpdateShippingStatus(orderID uint, status string) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", orderID).Update("shipping_status", status).Error
}
