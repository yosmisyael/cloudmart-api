package service

import (
	"fmt"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

type PaymentService interface {
	CreateSnapTransaction(order *entity.Order, user *entity.User) (snapToken string, paymentURL string, err error)
}

type paymentService struct {
	serverKey    string
	clientKey    string
	isProduction bool
}

func NewPaymentService(cfg *config.Config) PaymentService {
	return &paymentService{
		serverKey:    cfg.MidtransServerKey,
		clientKey:    cfg.MidtransClientKey,
		isProduction: cfg.MidtransEnv == "production",
	}
}

func (s *paymentService) CreateSnapTransaction(order *entity.Order, user *entity.User) (string, string, error) {
	var snapClient snap.Client
	env := midtrans.Sandbox
	if s.isProduction {
		env = midtrans.Production
	}
	snapClient.New(s.serverKey, env)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("%d-%d", order.ID, time.Now().Unix()),
			GrossAmt: int64(order.GrandTotal),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
		Callbacks: &snap.Callbacks{
			Finish: "http://localhost:5173/payment-success",
		},
	}

	resp, err := snapClient.CreateTransaction(snapReq)
	if err != nil {
		return "", "", err
	}

	return resp.Token, resp.RedirectURL, nil
}
