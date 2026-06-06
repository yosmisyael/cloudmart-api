package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *entity.Order, items []entity.OrderItem, userID uint, cartItemIDs []uint) error
	FindByUserID(userID uint) ([]entity.Order, error)
	FindByID(id, userID uint) (*entity.Order, error)
	FindByOrderID(orderID uint) (*entity.Order, error)
	UpdatePaymentStatus(orderID uint, status string) error
	UpdateSnapToken(orderID uint, snapToken, paymentURL string) error
	UpdatePaymentMethod(orderID uint, method string) error
	FindOrderItemByID(itemID uint) (*entity.OrderItem, error)
	CancelOrder(orderID uint) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db}
}

// CreateOrder creates an order with items, decrements stock, and clears the user's cart
// all within a single database transaction for atomicity.
func (r *orderRepository) CreateOrder(order *entity.Order, items []entity.OrderItem, userID uint, cartItemIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for _, item := range items {
			item.OrderID = order.ID
			if err := tx.Create(&item).Error; err != nil {
				return err
			}

			// Defensive stock deduction: WHERE clause ensures stock >= quantity
			// preventing negative stock even under concurrent requests.
			res := tx.Model(&entity.ProductVariant{}).
				Where("id = ? AND stock >= ?", item.VariantID, item.Quantity).
				Update("stock", gorm.Expr("stock - ?", item.Quantity))

			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		if len(cartItemIDs) > 0 {
			if err := tx.Where("id IN ?", cartItemIDs).Delete(&entity.Cart{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("user_id = ?", userID).Delete(&entity.Cart{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *orderRepository) FindByUserID(userID uint) ([]entity.Order, error) {
	var orders []entity.Order
	err := r.db.Where("user_id = ?", userID).
		Preload("OrderItems").
		Preload("OrderItems.Variant").
		Preload("OrderItems.Variant.Product").
		Preload("OrderItems.Variant.Product.Store").
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByID(id, userID uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("OrderItems").
		Preload("OrderItems.Variant").
		Preload("OrderItems.Variant.Product").
		Preload("OrderItems.Variant.Product.Store").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderID(orderID uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.Preload("OrderItems").First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdatePaymentStatus(orderID uint, status string) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", orderID).Update("payment_status", status).Error
}

func (r *orderRepository) UpdateSnapToken(orderID uint, snapToken, paymentURL string) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"snap_token":  snapToken,
			"payment_url": paymentURL,
		}).Error
}

func (r *orderRepository) UpdatePaymentMethod(orderID uint, method string) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", orderID).Update("payment_method", method).Error
}

func (r *orderRepository) FindOrderItemByID(itemID uint) (*entity.OrderItem, error) {
	var item entity.OrderItem
	err := r.db.Preload("Order").First(&item, itemID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderRepository) CancelOrder(orderID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var items []entity.OrderItem
		if err := tx.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
			return err
		}

		for _, item := range items {
			if err := tx.Model(&entity.ProductVariant{}).
				Where("id = ?", item.VariantID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&entity.Order{}).Where("id = ?", orderID).
			Update("payment_status", "cancel").Error; err != nil {
			return err
		}

		return nil
	})
}
