package service

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type WebhookService interface {
	HandleMidtransNotification(orderIDStr string, statusCode, grossAmount, signatureKey, transactionStatus, paymentType string) error
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

func (s *webhookService) HandleMidtransNotification(orderIDStr string, statusCode, grossAmount, signatureKey, transactionStatus, paymentType string) error {
	raw := orderIDStr + statusCode + grossAmount + s.config.MidtransServerKey
	hash := sha512.Sum512([]byte(raw))
	expectedSignature := hex.EncodeToString(hash[:])

	if expectedSignature != signatureKey {
		return fmt.Errorf("signature tidak valid")
	}

	parts := strings.Split(orderIDStr, "-")
	orderID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("order id invalid")
	}

	if _, err := s.orderRepo.FindByOrderID(uint(orderID)); err != nil {
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

	if paymentType != "" {
		_ = s.orderRepo.UpdatePaymentMethod(uint(orderID), paymentType)
	}

	return s.orderRepo.UpdatePaymentStatus(uint(orderID), paymentStatus)
}
