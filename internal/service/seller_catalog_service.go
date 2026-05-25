package service

import (
	"context"
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type CreateProductInput struct {
	CategoryID  uint
	Name        string
	Description string
}

type UpdateProductInput struct {
	CategoryID  uint
	Name        string
	Description string
}

type VariantInput struct {
	SKU   string
	Color string
	Size  string
	Price float64
	Stock int
}

type SellerCatalogService interface {
	GetMyCategories(userID uint) ([]entity.Category, error)
	CreateCategory(userID uint, name string) (*entity.Category, error)
	UpdateCategory(userID, categoryID uint, name string) (*entity.Category, error)
	DeleteCategory(userID, categoryID uint) error

	GetMyProducts(userID uint) ([]entity.Product, error)
	CreateProduct(userID uint, req CreateProductInput) (*entity.Product, error)
	UpdateProduct(userID, productID uint, req UpdateProductInput) (*entity.Product, error)
	DeleteProduct(userID, productID uint) error

	GetVariants(userID, productID uint) ([]entity.ProductVariant, error)
	CreateVariant(userID, productID uint, req VariantInput) (*entity.ProductVariant, error)
	UpdateVariant(userID, variantID uint, req VariantInput) (*entity.ProductVariant, error)
	DeleteVariant(userID, variantID uint) error

	UploadProductImage(ctx context.Context, userID, productID uint, data []byte, filename, contentType string) (string, error)
	DeleteProductImage(ctx context.Context, userID, productID uint) error
	UploadVariantImage(ctx context.Context, userID, variantID uint, data []byte, filename, contentType string) (string, error)
	DeleteVariantImage(ctx context.Context, userID, variantID uint) error
}

type sellerCatalogService struct {
	storeRepo    repository.StoreRepository
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
	s3Svc        S3Service
}

func NewSellerCatalogService(storeRepo repository.StoreRepository, categoryRepo repository.CategoryRepository, productRepo repository.ProductRepository, s3Svc S3Service) SellerCatalogService {
	return &sellerCatalogService{storeRepo, categoryRepo, productRepo, s3Svc}
}

func (s *sellerCatalogService) GetMyCategories(userID uint) ([]entity.Category, error) {
	categories, err := s.categoryRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data kategori")
	}
	return categories, nil
}

func (s *sellerCatalogService) CreateCategory(userID uint, name string) (*entity.Category, error) {
	category := &entity.Category{
		Name:   name,
		UserID: &userID,
	}
	if err := s.categoryRepo.Create(category); err != nil {
		return nil, errors.New("gagal membuat kategori")
	}
	return category, nil
}

func (s *sellerCatalogService) UpdateCategory(userID, categoryID uint, name string) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}

	if category.UserID == nil || *category.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	category.Name = name
	if err := s.categoryRepo.Update(category); err != nil {
		return nil, errors.New("gagal mengupdate kategori")
	}
	return category, nil
}

func (s *sellerCatalogService) DeleteCategory(userID, categoryID uint) error {
	category, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		return errors.New("kategori tidak ditemukan")
	}

	if category.UserID == nil || *category.UserID != userID {
		return errors.New("akses ditolak")
	}

	if err := s.categoryRepo.Delete(categoryID); err != nil {
		return errors.New("gagal menghapus kategori")
	}
	return nil
}

func (s *sellerCatalogService) GetMyProducts(userID uint) ([]entity.Product, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	products, err := s.productRepo.FindByStoreID(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data produk")
	}
	return products, nil
}

func (s *sellerCatalogService) CreateProduct(userID uint, req CreateProductInput) (*entity.Product, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	product := &entity.Product{
		StoreID:     store.ID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.productRepo.CreateProduct(product); err != nil {
		return nil, errors.New("gagal membuat produk")
	}
	return product, nil
}

func (s *sellerCatalogService) UpdateProduct(userID, productID uint, req UpdateProductInput) (*entity.Product, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = req.Description

	if err := s.productRepo.UpdateProduct(&product); err != nil {
		return nil, errors.New("gagal mengupdate produk")
	}
	return &product, nil
}

func (s *sellerCatalogService) DeleteProduct(userID, productID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if err := s.productRepo.DeleteProduct(productID); err != nil {
		return errors.New("gagal menghapus produk")
	}
	return nil
}

func (s *sellerCatalogService) GetVariants(userID, productID uint) ([]entity.ProductVariant, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	variants, err := s.productRepo.FindVariantsByProductID(productID)
	if err != nil {
		return nil, errors.New("gagal mengambil data varian")
	}
	return variants, nil
}

func (s *sellerCatalogService) CreateVariant(userID, productID uint, req VariantInput) (*entity.ProductVariant, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	variant := &entity.ProductVariant{
		ProductID: productID,
		SKU:       req.SKU,
		Color:     req.Color,
		Size:      req.Size,
		Price:     req.Price,
		Stock:     req.Stock,
	}

	if err := s.productRepo.CreateVariant(variant); err != nil {
		return nil, errors.New("SKU sudah digunakan")
	}
	return variant, nil
}

func (s *sellerCatalogService) UpdateVariant(userID, variantID uint, req VariantInput) (*entity.ProductVariant, error) {
	variant, err := s.productRepo.FindVariantByID(variantID)
	if err != nil {
		return nil, errors.New("varian tidak ditemukan")
	}

	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	variant.SKU = req.SKU
	variant.Color = req.Color
	variant.Size = req.Size
	variant.Price = req.Price
	variant.Stock = req.Stock

	if err := s.productRepo.UpdateVariant(&variant); err != nil {
		return nil, errors.New("gagal mengupdate varian")
	}
	return &variant, nil
}

func (s *sellerCatalogService) DeleteVariant(userID, variantID uint) error {
	variant, err := s.productRepo.FindVariantByID(variantID)
	if err != nil {
		return errors.New("varian tidak ditemukan")
	}

	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if err := s.productRepo.DeleteVariant(variantID); err != nil {
		return errors.New("gagal menghapus varian")
	}
	return nil
}

func (s *sellerCatalogService) UploadProductImage(ctx context.Context, userID, productID uint, data []byte, filename, contentType string) (string, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return "", errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return "", errors.New("akses ditolak")
	}

	if product.ImageURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(product.ImageURL))
	}

	url, err := s.s3Svc.UploadFile(ctx, "products", filename, data, contentType)
	if err != nil {
		return "", errors.New("gagal upload gambar")
	}

	if err := s.productRepo.UpdateImageURL(productID, url); err != nil {
		return "", errors.New("gagal update url gambar")
	}

	return url, nil
}

func (s *sellerCatalogService) DeleteProductImage(ctx context.Context, userID, productID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if product.ImageURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(product.ImageURL))
		_ = s.productRepo.UpdateImageURL(productID, "")
	}

	return nil
}

func (s *sellerCatalogService) UploadVariantImage(ctx context.Context, userID, variantID uint, data []byte, filename, contentType string) (string, error) {
	variant, err := s.productRepo.FindVariantByID(variantID)
	if err != nil {
		return "", errors.New("varian tidak ditemukan")
	}

	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return "", errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return "", errors.New("akses ditolak")
	}

	if variant.ImageURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(variant.ImageURL))
	}

	url, err := s.s3Svc.UploadFile(ctx, "variants", filename, data, contentType)
	if err != nil {
		return "", errors.New("gagal upload gambar varian")
	}

	if err := s.productRepo.UpdateVariantImageURL(variantID, url); err != nil {
		return "", errors.New("gagal update url gambar varian")
	}

	return url, nil
}

func (s *sellerCatalogService) DeleteVariantImage(ctx context.Context, userID, variantID uint) error {
	variant, err := s.productRepo.FindVariantByID(variantID)
	if err != nil {
		return errors.New("varian tidak ditemukan")
	}

	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	product, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if variant.ImageURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(variant.ImageURL))
		_ = s.productRepo.UpdateVariantImageURL(variantID, "")
	}

	return nil
}

