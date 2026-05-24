package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type SellerStoreService interface {
	GetStore(userID uint) (*entity.Store, error)
	CreateStore(userID uint, name string) (*entity.Store, error)
	UpdateStore(userID uint, name string, addressID *uint) (*entity.Store, error)
}

type sellerStoreService struct {
	storeRepo repository.StoreRepository
	userRepo  repository.UserRepository
}

func NewSellerStoreService(storeRepo repository.StoreRepository, userRepo repository.UserRepository) SellerStoreService {
	return &sellerStoreService{storeRepo, userRepo}
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
