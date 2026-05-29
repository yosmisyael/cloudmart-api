package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type SellerOrderService interface {
	GetOrders(userID uint, status string) ([]entity.Order, error)
	GetOrderByID(userID, orderID uint) (*entity.Order, error)
	UpdateOrderStatus(userID, orderID uint, status string) error
}

type sellerOrderService struct {
	storeRepo repository.StoreRepository
	orderRepo repository.SellerOrderRepository
}

func NewSellerOrderService(storeRepo repository.StoreRepository, orderRepo repository.SellerOrderRepository) SellerOrderService {
	return &sellerOrderService{storeRepo, orderRepo}
}

func (s *sellerOrderService) GetOrders(userID uint, status string) ([]entity.Order, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	orders, err := s.orderRepo.FindByStoreID(store.ID, status)
	if err != nil {
		return nil, errors.New("gagal mengambil data pesanan")
	}
	return orders, nil
}

func (s *sellerOrderService) GetOrderByID(userID, orderID uint) (*entity.Order, error) {
	_, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	order, err := s.orderRepo.FindOrderByID(orderID)
	if err != nil {
		return nil, errors.New("pesanan tidak ditemukan")
	}
	return order, nil
}

func (s *sellerOrderService) UpdateOrderStatus(userID, orderID uint, status string) error {
	_, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	validStatuses := map[string]bool{
		"processing": true,
		"shipped":    true,
		"delivered":  true,
	}

	if !validStatuses[status] {
		return errors.New("status tidak valid")
	}

	_, err = s.orderRepo.FindOrderByID(orderID)
	if err != nil {
		return errors.New("pesanan tidak ditemukan")
	}

	if err := s.orderRepo.UpdateShippingStatus(orderID, status); err != nil {
		return errors.New("gagal mengupdate status pesanan")
	}
	return nil
}
