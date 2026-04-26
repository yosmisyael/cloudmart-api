package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type CartService interface {
	GetCart(userID uint) ([]CartItemResponse, float64, error)
	AddToCart(userID, variantID uint, quantity int) error
	UpdateQuantity(cartID, userID uint, quantity int) error
	RemoveItem(cartID, userID uint) error
}

type CartItemResponse struct {
	ID           uint    `json:"id"`
	VariantID    uint    `json:"variant_id"`
	VariantName  string  `json:"variant_name"`
	VariantColor string  `json:"variant_color"`
	VariantSize  string  `json:"variant_size"`
	VariantImage string  `json:"variant_image"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	Subtotal     float64 `json:"subtotal"`
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) GetCart(userID uint) ([]CartItemResponse, float64, error) {
	carts, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	var items []CartItemResponse
	var grandTotal float64

	for _, cart := range carts {
		subtotal := cart.Variant.Price * float64(cart.Quantity)
		grandTotal += subtotal

		items = append(items, CartItemResponse{
			ID:           cart.ID,
			VariantID:    cart.VariantID,
			VariantName:  cart.Variant.Product.Name,
			VariantColor: cart.Variant.Color,
			VariantSize:  cart.Variant.Size,
			Price:        cart.Variant.Price,
			Quantity:     cart.Quantity,
			Subtotal:     subtotal,
		})
	}

	return items, grandTotal, nil
}

func (s *cartService) AddToCart(userID, variantID uint, quantity int) error {
	if quantity < 1 {
		return errors.New("quantity minimal 1")
	}

	if _, err := s.productRepo.FindVariantByID(variantID); err != nil {
		return errors.New("variant tidak ditemukan")
	}

	existing, err := s.cartRepo.FindByUserAndVariant(userID, variantID)
	if err == nil && existing != nil {
		existing.Quantity += quantity
		return s.cartRepo.Update(existing)
	}

	cart := &entity.Cart{
		UserID:    userID,
		VariantID: variantID,
		Quantity:  quantity,
	}
	return s.cartRepo.Create(cart)
}

func (s *cartService) UpdateQuantity(cartID, userID uint, quantity int) error {
	if quantity < 1 {
		return errors.New("quantity minimal 1")
	}

	cart, err := s.cartRepo.FindByID(cartID, userID)
	if err != nil {
		return errors.New("item tidak ditemukan di keranjang")
	}

	cart.Quantity = quantity
	return s.cartRepo.Update(cart)
}

func (s *cartService) RemoveItem(cartID, userID uint) error {
	_, err := s.cartRepo.FindByID(cartID, userID)
	if err != nil {
		return errors.New("item tidak ditemukan di keranjang")
	}
	return s.cartRepo.Delete(cartID, userID)
}
