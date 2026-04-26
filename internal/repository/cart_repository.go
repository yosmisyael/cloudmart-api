package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type CartRepository interface {
	FindByUserID(userID uint) ([]entity.Cart, error)
	FindByUserAndVariant(userID, variantID uint) (*entity.Cart, error)
	FindByID(id, userID uint) (*entity.Cart, error)
	Create(cart *entity.Cart) error
	Update(cart *entity.Cart) error
	Delete(id, userID uint) error
	DeleteByUserID(userID uint) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db}
}

func (r *cartRepository) FindByUserID(userID uint) ([]entity.Cart, error) {
	var carts []entity.Cart
	err := r.db.Where("user_id = ?", userID).
		Preload("Variant").
		Preload("Variant.Product").
		Find(&carts).Error
	return carts, err
}

func (r *cartRepository) FindByUserAndVariant(userID, variantID uint) (*entity.Cart, error) {
	var cart entity.Cart
	err := r.db.Where("user_id = ? AND variant_id = ?", userID, variantID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) FindByID(id, userID uint) (*entity.Cart, error) {
	var cart entity.Cart
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Create(cart *entity.Cart) error {
	return r.db.Create(cart).Error
}

func (r *cartRepository) Update(cart *entity.Cart) error {
	return r.db.Save(cart).Error
}

func (r *cartRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Cart{}).Error
}

func (r *cartRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.Cart{}).Error
}
