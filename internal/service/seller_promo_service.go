package service

import (
	"errors"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type VoucherInput struct {
	Code      string
	Name      string
	Type      string
	Amount    float64
	Max       float64
	ExpiredAt time.Time
}

type BankInput struct {
	Name        string
	AccountID   string
	AccountName string
}

type SellerPromoService interface {
	GetVouchers(userID uint) ([]entity.Voucher, error)
	CreateVoucher(userID uint, req VoucherInput) (*entity.Voucher, error)
	UpdateVoucher(userID, voucherID uint, req VoucherInput) (*entity.Voucher, error)
	DeleteVoucher(userID, voucherID uint) error

	GetPaymentConfigs(userID uint) ([]entity.PaymentConfiguration, error)
	CreatePaymentConfig(userID uint, name string) (*entity.PaymentConfiguration, error)
	DeletePaymentConfig(userID, configID uint) error
	UpdatePaymentConfig(configID, sellerUserID uint, name string) (*entity.PaymentConfiguration, error)
	AddBank(userID, configID uint, req BankInput) (*entity.PaymentBank, error)
	DeleteBank(userID, bankID uint) error
	UpdateBank(bankID, sellerUserID uint, name, accountID, accountName string) (*entity.PaymentBank, error)
}

type sellerPromoService struct {
	storeRepo   repository.StoreRepository
	voucherRepo repository.VoucherRepository
	paymentRepo repository.PaymentRepository
}

func NewSellerPromoService(storeRepo repository.StoreRepository, voucherRepo repository.VoucherRepository, paymentRepo repository.PaymentRepository) SellerPromoService {
	return &sellerPromoService{storeRepo, voucherRepo, paymentRepo}
}

func (s *sellerPromoService) GetVouchers(userID uint) ([]entity.Voucher, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	vouchers, err := s.voucherRepo.FindByStoreID(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data voucher")
	}
	return vouchers, nil
}

func (s *sellerPromoService) CreateVoucher(userID uint, req VoucherInput) (*entity.Voucher, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	voucher := &entity.Voucher{
		StoreID:   store.ID,
		Code:      req.Code,
		Name:      req.Name,
		Type:      req.Type,
		Amount:    req.Amount,
		Max:       req.Max,
		ExpiredAt: req.ExpiredAt,
	}

	if err := s.voucherRepo.Create(voucher); err != nil {
		return nil, errors.New("gagal membuat voucher")
	}
	return voucher, nil
}

func (s *sellerPromoService) UpdateVoucher(userID, voucherID uint, req VoucherInput) (*entity.Voucher, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	voucher, err := s.voucherRepo.FindByID(voucherID)
	if err != nil {
		return nil, errors.New("voucher tidak ditemukan")
	}

	if voucher.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	voucher.Code = req.Code
	voucher.Name = req.Name
	voucher.Type = req.Type
	voucher.Amount = req.Amount
	voucher.Max = req.Max
	voucher.ExpiredAt = req.ExpiredAt

	if err := s.voucherRepo.Update(voucher); err != nil {
		return nil, errors.New("gagal mengupdate voucher")
	}
	return voucher, nil
}

func (s *sellerPromoService) DeleteVoucher(userID, voucherID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	voucher, err := s.voucherRepo.FindByID(voucherID)
	if err != nil {
		return errors.New("voucher tidak ditemukan")
	}

	if voucher.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if err := s.voucherRepo.Delete(voucherID); err != nil {
		return errors.New("gagal menghapus voucher")
	}
	return nil
}

func (s *sellerPromoService) GetPaymentConfigs(userID uint) ([]entity.PaymentConfiguration, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	configs, err := s.paymentRepo.FindByStoreID(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data konfigurasi pembayaran")
	}
	return configs, nil
}

func (s *sellerPromoService) CreatePaymentConfig(userID uint, name string) (*entity.PaymentConfiguration, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	config := &entity.PaymentConfiguration{
		StoreID: store.ID,
		Name:    name,
	}

	if err := s.paymentRepo.Create(config); err != nil {
		return nil, errors.New("gagal membuat konfigurasi pembayaran")
	}
	return config, nil
}

func (s *sellerPromoService) DeletePaymentConfig(userID, configID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	config, err := s.paymentRepo.FindByID(configID)
	if err != nil {
		return errors.New("konfigurasi pembayaran tidak ditemukan")
	}

	if config.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if err := s.paymentRepo.Delete(configID); err != nil {
		return errors.New("gagal menghapus konfigurasi pembayaran")
	}
	return nil
}

func (s *sellerPromoService) AddBank(userID, configID uint, req BankInput) (*entity.PaymentBank, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	config, err := s.paymentRepo.FindByID(configID)
	if err != nil {
		return nil, errors.New("konfigurasi pembayaran tidak ditemukan")
	}

	if config.StoreID != store.ID {
		return nil, errors.New("akses ditolak")
	}

	bank := &entity.PaymentBank{
		PaymentConfigID: configID,
		Name:            req.Name,
		AccountID:       req.AccountID,
		AccountName:     req.AccountName,
	}

	if err := s.paymentRepo.CreateBank(bank); err != nil {
		return nil, errors.New("gagal menambahkan bank")
	}
	return bank, nil
}

func (s *sellerPromoService) DeleteBank(userID, bankID uint) error {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("toko tidak ditemukan")
	}

	bank, err := s.paymentRepo.FindBankByID(bankID)
	if err != nil {
		return errors.New("bank tidak ditemukan")
	}

	config, err := s.paymentRepo.FindByID(bank.PaymentConfigID)
	if err != nil {
		return errors.New("konfigurasi pembayaran tidak ditemukan")
	}

	if config.StoreID != store.ID {
		return errors.New("akses ditolak")
	}

	if err := s.paymentRepo.DeleteBank(bankID); err != nil {
		return errors.New("gagal menghapus bank")
	}
	return nil
}

func (s *sellerPromoService) UpdatePaymentConfig(configID, sellerUserID uint, name string) (*entity.PaymentConfiguration, error) {
	store, err := s.storeRepo.FindByUserID(sellerUserID)
	if err != nil {
		return nil, errors.New("store not found")
	}

	config, err := s.paymentRepo.GetConfigByIDAndStoreID(configID, store.ID)
	if err != nil {
		return nil, errors.New("payment config not found or not owned by seller")
	}

	config.Name = name
	if err := s.paymentRepo.UpdateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *sellerPromoService) UpdateBank(bankID, sellerUserID uint, name, accountID, accountName string) (*entity.PaymentBank, error) {
	store, err := s.storeRepo.FindByUserID(sellerUserID)
	if err != nil {
		return nil, errors.New("store not found")
	}

	bank, err := s.paymentRepo.GetBankByIDAndStoreID(bankID, store.ID)
	if err != nil {
		return nil, errors.New("bank not found or not owned by seller")
	}

	bank.Name = name
	bank.AccountID = accountID
	bank.AccountName = accountName

	if err := s.paymentRepo.UpdateBank(bank); err != nil {
		return nil, err
	}
	return bank, nil
}
