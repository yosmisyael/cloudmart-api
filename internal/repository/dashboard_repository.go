package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	CountOrdersByStore(storeID uint) (int64, error)
	CountPendingOrdersByStore(storeID uint) (int64, error)
	SumRevenueByStore(storeID uint) (float64, error)
	CountProductsByStore(storeID uint) (int64, error)
	CountLowStockByStore(storeID uint, threshold int) (int64, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db}
}

func (r *dashboardRepository) CountOrdersByStore(storeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Order{}).
		Where("id IN (?)",
			r.db.Table("order_items").
				Select("DISTINCT order_items.order_id").
				Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
				Joins("JOIN products ON products.id = product_variants.product_id").
				Where("products.store_id = ?", storeID),
		).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountPendingOrdersByStore(storeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Order{}).
		Where("payment_status = ?", "pending").
		Where("id IN (?)",
			r.db.Table("order_items").
				Select("DISTINCT order_items.order_id").
				Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
				Joins("JOIN products ON products.id = product_variants.product_id").
				Where("products.store_id = ?", storeID),
		).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) SumRevenueByStore(storeID uint) (float64, error) {
	var revenue float64
	err := r.db.Model(&entity.Order{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("id IN (?)",
			r.db.Table("order_items").
				Select("DISTINCT order_items.order_id").
				Joins("JOIN product_variants ON product_variants.id = order_items.variant_id").
				Joins("JOIN products ON products.id = product_variants.product_id").
				Where("products.store_id = ?", storeID),
		).Scan(&revenue).Error
	return revenue, err
}

func (r *dashboardRepository) CountProductsByStore(storeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Product{}).Where("store_id = ?", storeID).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountLowStockByStore(storeID uint, threshold int) (int64, error) {
	var count int64
	err := r.db.Model(&entity.ProductVariant{}).
		Joins("JOIN products ON products.id = product_variants.product_id").
		Where("products.store_id = ? AND product_variants.stock < ?", storeID, threshold).
		Count(&count).Error
	return count, err
}
