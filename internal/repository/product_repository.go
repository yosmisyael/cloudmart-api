package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll(page, limit int, categoryID uint, search string) ([]entity.Product, int64, error)
	FindByID(id uint) (entity.Product, error)
	FindVariantByID(id uint) (entity.ProductVariant, error)
	CreateProduct(p *entity.Product) error
	UpdateProduct(p *entity.Product) error
	DeleteProduct(id uint) error
	FindByStoreID(storeID uint) ([]entity.Product, error)
	CreateVariant(v *entity.ProductVariant) error
	UpdateVariant(v *entity.ProductVariant) error
	DeleteVariant(id uint) error
	FindVariantsByProductID(productID uint) ([]entity.ProductVariant, error)
	UpdateImageURL(productID uint, url string) error
	UpdateVariantImageURL(variantID uint, url string) error
	GetByIDForSeller(productID, sellerUserID uint) (*entity.Product, error)
	GetVariantByIDForSeller(variantID, sellerUserID uint) (*entity.ProductVariant, error)
	UpdateVariantStock(variant *entity.ProductVariant) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) FindAll(page, limit int, categoryID uint, search string) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	query := r.db.Model(&entity.Product{})

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Category").Preload("Variants").
		Limit(limit).Offset(offset).
		Order("id DESC").
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) FindByID(id uint) (entity.Product, error) {
	var product entity.Product
	err := r.db.Preload("Store").Preload("Category").Preload("Variants").First(&product, id).Error
	return product, err
}

func (r *productRepository) FindVariantByID(id uint) (entity.ProductVariant, error) {
	var productVariant entity.ProductVariant
	err := r.db.Preload("Product").Preload("Product.Category").First(&productVariant, id).Error
	return productVariant, err
}

func (r *productRepository) CreateProduct(p *entity.Product) error {
	return r.db.Create(p).Error
}

func (r *productRepository) UpdateProduct(p *entity.Product) error {
	return r.db.Save(p).Error
}

func (r *productRepository) DeleteProduct(id uint) error {
	return r.db.Delete(&entity.Product{}, id).Error
}

func (r *productRepository) FindByStoreID(storeID uint) ([]entity.Product, error) {
	var products []entity.Product
	err := r.db.Preload("Category").Preload("Variants").Where("store_id = ?", storeID).Find(&products).Error
	return products, err
}

func (r *productRepository) CreateVariant(v *entity.ProductVariant) error {
	return r.db.Create(v).Error
}

func (r *productRepository) UpdateVariant(v *entity.ProductVariant) error {
	return r.db.Save(v).Error
}

func (r *productRepository) DeleteVariant(id uint) error {
	return r.db.Delete(&entity.ProductVariant{}, id).Error
}

func (r *productRepository) FindVariantsByProductID(productID uint) ([]entity.ProductVariant, error) {
	var variants []entity.ProductVariant
	err := r.db.Where("product_id = ?", productID).Find(&variants).Error
	return variants, err
}

func (r *productRepository) UpdateImageURL(productID uint, url string) error {
	return r.db.Model(&entity.Product{}).Where("id = ?", productID).Update("image_url", url).Error
}

func (r *productRepository) UpdateVariantImageURL(variantID uint, url string) error {
	return r.db.Model(&entity.ProductVariant{}).Where("id = ?", variantID).Update("image_url", url).Error
}

func (r *productRepository) GetByIDForSeller(productID, sellerUserID uint) (*entity.Product, error) {
	var product entity.Product
	result := r.db.
		Joins("JOIN stores ON stores.id = products.store_id").
		Where("products.id = ? AND stores.user_id = ?", productID, sellerUserID).
		Preload("Variants").
		Preload("Category").
		Preload("Store").
		First(&product)

	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func (r *productRepository) GetVariantByIDForSeller(variantID, sellerUserID uint) (*entity.ProductVariant, error) {
	var variant entity.ProductVariant
	result := r.db.
		Joins("JOIN products ON products.id = product_variants.product_id").
		Joins("JOIN stores ON stores.id = products.store_id").
		Where("product_variants.id = ? AND stores.user_id = ?", variantID, sellerUserID).
		First(&variant)
	if result.Error != nil {
		return nil, result.Error
	}
	return &variant, nil
}

func (r *productRepository) UpdateVariantStock(variant *entity.ProductVariant) error {
	return r.db.Model(variant).Update("stock", variant.Stock).Error
}
