package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type OrderService interface {
	Checkout(userID uint, address string) (*entity.Order, error)
	GetOrders(userID uint) ([]entity.Order, error)
	GetOrderByID(id, userID uint) (*entity.Order, error)
	InitiatePayment(userID, orderID uint) (*entity.Order, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	authRepo    repository.AuthRepository
	paymentSvc  PaymentService
}

func NewOrderService(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository, authRepo repository.AuthRepository, paymentSvc PaymentService) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		authRepo:    authRepo,
		paymentSvc:  paymentSvc,
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

	go s.sendOrderNotification(order.ID, userID)

	return &order, nil
}

func (s *orderService) GetOrders(userID uint) ([]entity.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

func (s *orderService) GetOrderByID(id, userID uint) (*entity.Order, error) {
	return s.orderRepo.FindByID(id, userID)
}

func (s *orderService) sendOrderNotification(orderID, userID uint) {
	log.Println("Simulate notification send")

	time.Sleep(2 * time.Second)

	log.Println("Notification sent successfully")
}

func (s *orderService) InitiatePayment(userID, orderID uint) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(orderID, userID)
	if err != nil {
		return nil, errors.New("pesanan tidak ditemukan")
	}

	if order.PaymentStatus != "pending" {
		return nil, errors.New("pembayaran sudah diproses")
	}

	user, err := s.authRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	snapToken, paymentURL, err := s.paymentSvc.CreateSnapTransaction(order, user)
	if err != nil {
		return nil, errors.New("gagal membuat transaksi pembayaran")
	}

	if err := s.orderRepo.UpdateSnapToken(orderID, snapToken, paymentURL); err != nil {
		return nil, errors.New("gagal menyimpan token pembayaran")
	}

	order.SnapToken = &snapToken
	order.PaymentURL = &paymentURL

	return order, nil
}
