package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type OrderEstimate struct {
	Subtotal     float64 `json:"subtotal"`
	ShippingFee  float64 `json:"shipping_fee"`
	Discount     float64 `json:"discount"`
	VoucherName  string  `json:"voucher_name,omitempty"`
	GrandTotal   float64 `json:"grand_total"`
}

type OrderService interface {
	Checkout(userID uint, addressID *uint, address string, logisticServiceID uint, voucherCode string) (*entity.Order, error)
	EstimateOrder(userID uint, logisticServiceID uint, voucherCode string) (*OrderEstimate, error)
	CancelOrder(userID, orderID uint) error
	GetOrders(userID uint) ([]entity.Order, error)
	GetOrderByID(id, userID uint) (*entity.Order, error)
	InitiatePayment(userID, orderID uint) (*entity.Order, error)
}

type orderService struct {
	orderRepo    repository.OrderRepository
	cartRepo     repository.CartRepository
	productRepo  repository.ProductRepository
	authRepo     repository.AuthRepository
	userRepo     repository.UserRepository
	logisticRepo repository.LogisticRepository
	voucherRepo  repository.VoucherRepository
	paymentSvc   PaymentService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
	logisticRepo repository.LogisticRepository,
	voucherRepo repository.VoucherRepository,
	paymentSvc PaymentService,
) OrderService {
	return &orderService{
		orderRepo:    orderRepo,
		cartRepo:     cartRepo,
		productRepo:  productRepo,
		authRepo:     authRepo,
		userRepo:     userRepo,
		logisticRepo: logisticRepo,
		voucherRepo:  voucherRepo,
		paymentSvc:   paymentSvc,
	}
}

func (s *orderService) Checkout(userID uint, addressID *uint, address string, logisticServiceID uint, voucherCode string) (*entity.Order, error) {
	if addressID == nil && address == "" {
		return nil, errors.New("alamat pengiriman harus diisi")
	}

	var shippingAddress string
	if addressID != nil {
		addrRecords, err := s.userRepo.FindAddresses(userID)
		if err != nil {
			return nil, errors.New("gagal mengambil alamat")
		}
		var found *entity.Address
		for _, a := range addrRecords {
			if a.ID == *addressID {
				found = &a
				break
			}
		}
		if found == nil {
			return nil, errors.New("alamat tidak ditemukan atau bukan milik Anda")
		}
		shippingAddress = fmt.Sprintf("%s - %s, %s, %s %s", found.Recipient, found.Address, found.City, found.State, found.PostalCode)
	} else {
		if len(address) < 10 {
			return nil, errors.New("alamat minimal 10 karakter")
		}
		shippingAddress = address
	}

	logisticService, err := s.logisticRepo.FindServiceByID(logisticServiceID)
	if err != nil {
		return nil, errors.New("layanan logistik tidak ditemukan")
	}
	shippingFee := logisticService.BasePrice

	cartItems, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data keranjang")
	}

	if len(cartItems) == 0 {
		return nil, errors.New("keranjang kosong")
	}

	var subTotal float64
	var orderItems []entity.OrderItem

	for _, cart := range cartItems {
		variant, err := s.productRepo.FindVariantByID(cart.VariantID)
		if err != nil {
			return nil, fmt.Errorf("barang dengan ID %d tidak ditemukan", cart.VariantID)
		}

		if variant.Stock < cart.Quantity {
			return nil, fmt.Errorf("stok barang %s tidak mencukupi", variant.SKU)
		}

		itemSub := variant.Price * float64(cart.Quantity)
		subTotal += itemSub

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

	var discount float64
	var voucherID *uint

	if voucherCode != "" {
		voucher, err := s.voucherRepo.FindByCode(voucherCode)
		if err != nil {
			return nil, errors.New("voucher tidak ditemukan")
		}
		if !voucher.ExpiredAt.After(time.Now()) {
			return nil, errors.New("voucher sudah kadaluarsa")
		}
		claimed, _ := s.voucherRepo.IsClaimedByUser(userID, voucher.ID)
		if !claimed {
			return nil, errors.New("voucher belum diklaim")
		}

		switch voucher.Type {
		case "percentage":
			calc := subTotal * (voucher.Amount / 100)
			if voucher.Max > 0 && calc > voucher.Max {
				calc = voucher.Max
			}
			discount = calc
		case "price":
			discount = voucher.Amount
			if discount > subTotal {
				discount = subTotal
			}
		case "free_shipping":
			discount = shippingFee
		}
		voucherID = &voucher.ID
	}

	grandTotal := subTotal + shippingFee - discount
	if grandTotal < 0 {
		grandTotal = 0
	}

	logisticName := logisticService.Name

	order := entity.Order{
		UserID:            userID,
		GrandTotal:        grandTotal,
		ShippingAddress:   shippingAddress,
		LogisticService:   logisticName,
		LogisticVoucherID: voucherID,
		PaymentStatus:     "pending",
		ShippingStatus:    "pending",
	}

	if err := s.orderRepo.CreateOrder(&order, orderItems, userID); err != nil {
		return nil, fmt.Errorf("transaksi gagal: %v", err)
	}

	go s.sendOrderNotification(order.ID, userID)

	return &order, nil
}

func (s *orderService) EstimateOrder(userID uint, logisticServiceID uint, voucherCode string) (*OrderEstimate, error) {
	logisticService, err := s.logisticRepo.FindServiceByID(logisticServiceID)
	if err != nil {
		return nil, errors.New("layanan logistik tidak ditemukan")
	}
	shippingFee := logisticService.BasePrice

	cartItems, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data keranjang")
	}

	if len(cartItems) == 0 {
		return nil, errors.New("keranjang kosong")
	}

	var subTotal float64
	for _, cart := range cartItems {
		variant, err := s.productRepo.FindVariantByID(cart.VariantID)
		if err != nil {
			return nil, fmt.Errorf("barang dengan ID %d tidak ditemukan", cart.VariantID)
		}
		if variant.Stock < cart.Quantity {
			return nil, fmt.Errorf("stok barang %s tidak mencukupi", variant.SKU)
		}
		subTotal += variant.Price * float64(cart.Quantity)
	}

	var discount float64
	var voucherName string

	if voucherCode != "" {
		voucher, err := s.voucherRepo.FindByCode(voucherCode)
		if err != nil {
			return nil, errors.New("voucher tidak ditemukan")
		}
		if !voucher.ExpiredAt.After(time.Now()) {
			return nil, errors.New("voucher sudah kadaluarsa")
		}
		claimed, _ := s.voucherRepo.IsClaimedByUser(userID, voucher.ID)
		if !claimed {
			return nil, errors.New("voucher belum diklaim")
		}

		voucherName = voucher.Name
		switch voucher.Type {
		case "percentage":
			calc := subTotal * (voucher.Amount / 100)
			if voucher.Max > 0 && calc > voucher.Max {
				calc = voucher.Max
			}
			discount = calc
		case "price":
			discount = voucher.Amount
			if discount > subTotal {
				discount = subTotal
			}
		case "free_shipping":
			discount = shippingFee
		}
	}

	grandTotal := subTotal + shippingFee - discount
	if grandTotal < 0 {
		grandTotal = 0
	}

	return &OrderEstimate{
		Subtotal:    subTotal,
		ShippingFee: shippingFee,
		Discount:    discount,
		VoucherName: voucherName,
		GrandTotal:  grandTotal,
	}, nil
}

func (s *orderService) CancelOrder(userID, orderID uint) error {
	order, err := s.orderRepo.FindByID(orderID, userID)
	if err != nil {
		return errors.New("pesanan tidak ditemukan")
	}

	if order.PaymentStatus != "pending" {
		return errors.New("pesanan tidak dapat dibatalkan karena tidak dalam status pending")
	}

	if err := s.orderRepo.CancelOrder(orderID); err != nil {
		return errors.New("gagal membatalkan pesanan")
	}
	return nil
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
