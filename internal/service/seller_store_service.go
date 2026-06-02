package service

import (
	"context"
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type SellerStoreService interface {
	GetStore(userID uint) (*entity.Store, error)
	CreateStore(userID uint, name string) (*entity.Store, error)
	UpdateStore(userID uint, name string, addressID *uint) (*entity.Store, error)
	UploadStoreLogo(ctx context.Context, userID uint, data []byte, filename, contentType string) (string, error)
	DeleteStoreLogo(ctx context.Context, userID uint) error
}

type sellerStoreService struct {
	storeRepo repository.StoreRepository
	userRepo  repository.UserRepository
	s3Svc     S3Service
}

func NewSellerStoreService(storeRepo repository.StoreRepository, userRepo repository.UserRepository, s3Svc S3Service) SellerStoreService {
	return &sellerStoreService{storeRepo, userRepo, s3Svc}
}

func (s *sellerStoreService) GetStore(userID uint) (*entity.Store, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}
	return store, nil
}

func (s *sellerStoreService) CreateStore(userID uint, name string) (*entity.Store, error) {
	existing, _ := s.storeRepo.FindByUserID(userID)
	if existing != nil {
		return nil, errors.New("toko sudah ada")
	}

	store := &entity.Store{
		UserID: userID,
		Name:   name,
	}

	if err := s.storeRepo.Create(store); err != nil {
		return nil, errors.New("gagal membuat toko")
	}

	if err := s.userRepo.UpdateRole(userID, "seller"); err != nil {
		return nil, errors.New("gagal mengupdate role user")
	}

	return store, nil
}

func (s *sellerStoreService) UpdateStore(userID uint, name string, addressID *uint) (*entity.Store, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	store.Name = name
	store.AddressID = addressID

	if err := s.storeRepo.Update(store); err != nil {
		return nil, errors.New("gagal mengupdate toko")
	}

	return store, nil
}

func (s *sellerStoreService) UploadStoreLogo(ctx context.Context, userID uint, data []byte, filename, contentType string) (string, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("toko tidak ditemukan")
	}

	if store.LogoURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(store.LogoURL))
	}

	url, err := s.s3Svc.UploadFile(ctx, "stores", filename, data, contentType)
	if err != nil {
		return "", errors.New("gagal upload logo toko")
	}

	if err := s.storeRepo.UpdateLogoURL(store.ID, url); err != nil {
		// DB failed — roll back the S3 upload
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(url))
		return "", errors.New("gagal update url logo toko")
	}

	return url, nil
}

func (s *sellerStoreService) DeleteStoreLogo(ctx context.Context, userID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	if store.LogoURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(store.LogoURL))
		_ = s.storeRepo.UpdateLogoURL(store.ID, "")
	}

	return nil
}
