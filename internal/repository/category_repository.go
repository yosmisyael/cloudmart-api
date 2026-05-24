package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]entity.Category, error)
	FindByUserID(userID uint) ([]entity.Category, error)
	Create(c *entity.Category) error
	Update(c *entity.Category) error
	Delete(id uint) error
	FindByID(id uint) (*entity.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db}
}

func (r *categoryRepository) FindAll() ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByUserID(userID uint) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Where("user_id = ? OR is_default = true", userID).Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Create(c *entity.Category) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) Update(c *entity.Category) error {
	return r.db.Save(c).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Category{}, id).Error
}

func (r *categoryRepository) FindByID(id uint) (*entity.Category, error) {
	var category entity.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
