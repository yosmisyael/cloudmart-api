package service

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type WebhookService interface {
	HandleMidtransNotification(orderID uint, statusCode, grossAmount, signatureKey, transactionStatus string) error
}

type webhookService struct {
	orderRepo repository.OrderRepository
	config    *config.Config
}

func NewWebhookService(orderRepo repository.OrderRepository, cfg *config.Config) WebhookService {
	return &webhookService{
		orderRepo: orderRepo,
		config:    cfg,
	}
}

func (s *webhookService) HandleMidtransNotification(orderID uint, statusCode, grossAmount, signatureKey, transactionStatus string) error {
	raw := strconv.FormatUint(uint64(orderID), 10) + statusCode + grossAmount + s.config.MidtransServerKey
	hash := sha512.Sum512([]byte(raw))
	expectedSignature := hex.EncodeToString(hash[:])

	if expectedSignature != signatureKey {
		return fmt.Errorf("signature tidak valid")
	}

	if _, err := s.orderRepo.FindByOrderID(orderID); err != nil {
		return fmt.Errorf("order tidak ditemukan")
	}

	var paymentStatus string
	switch transactionStatus {
	case "capture", "settlement":
		paymentStatus = "settlement"
	case "deny", "cancel":
		paymentStatus = "cancel"
	case "expire":
		paymentStatus = "expire"
	case "pending":
		paymentStatus = "pending"
	default:
		paymentStatus = transactionStatus
	}

	return s.orderRepo.UpdatePaymentStatus(orderID, paymentStatus)
}
