package service

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type CatalogService interface {
	GetCategories() ([]entity.Category, error)
	GetProducts(page, limit int, categoryID uint, search string) ([]entity.Product, int64, error)
	GetProductByID(id uint) (entity.Product, error)
}

type catalogService struct {
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
}

func NewCatalogService(categoryRepo repository.CategoryRepository, productRepo repository.ProductRepository) CatalogService {
	return &catalogService{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
	}
}

func (s *catalogService) GetCategories() ([]entity.Category, error) {
	return s.categoryRepo.FindAll()
}

func (s *catalogService) GetProducts(page, limit int, categoryID uint, search string) ([]entity.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.productRepo.FindAll(page, limit, categoryID, search)
}

func (s *catalogService) GetProductByID(id uint) (entity.Product, error) {
	return s.productRepo.FindByID(id)
}
