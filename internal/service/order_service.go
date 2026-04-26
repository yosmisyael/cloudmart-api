package service

import (
	"errors"
	"fmt"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type OrderService interface {
	Checkout(userID uint, address string) (*entity.Order, error)
	GetOrders(userID uint) ([]entity.Order, error)
	GetOrderByID(id, userID uint) (*entity.Order, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *orderService) Checkout(userID uint, address string) (*entity.Order, error) {
	cartItems, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data keranjang")
	}

	if len(cartItems) == 0 {
		return nil, errors.New("keranjang kosong")
	}

	var grandTotal float64
	var orderItems []entity.OrderItem

	for _, cart := range cartItems {
		variant, err := s.productRepo.FindVariantByID(cart.VariantID)
		if err != nil {
			return nil, fmt.Errorf("barang dengan ID %d tidak ditemukan", cart.VariantID)
		}

		if variant.Stock < cart.Quantity {
			return nil, fmt.Errorf("stok barang %s tidak mencukupi", variant.SKU)
		}

		subTotal := variant.Price * float64(cart.Quantity)
		grandTotal += subTotal

		orderItems = append(orderItems, entity.OrderItem{
			VariantID: variant.ID,
			VariantDetails: fmt.Sprintf("%s (%s) - %s %s",
				variant.Product.Name,
				variant.Product.Category.Name,
				variant.Color,
				variant.Size,
			),
			Price:    variant.Price,
			Quantity: cart.Quantity,
		})
	}

	order := entity.Order{
		UserID:          userID,
		GrandTotal:      grandTotal,
		ShippingAddress: address,
		PaymentStatus:   "pending",
	}

	if err := s.orderRepo.CreateOrder(&order, orderItems, userID); err != nil {
		return nil, fmt.Errorf("transaksi gagal: %v", err)
	}

	return &order, nil
}

func (s *orderService) GetOrders(userID uint) ([]entity.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderService) GetOrderByID(id, userID uint) (*entity.Order, error) {
	return s.orderRepo.FindByID(id, userID)
}
