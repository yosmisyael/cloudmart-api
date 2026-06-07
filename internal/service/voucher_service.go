package service

import (
	"errors"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type VoucherService interface {
	GetAvailableVouchers(storeID uint) ([]entity.Voucher, error)
	ClaimVoucher(userID uint, code string) error
	GetMyVouchers(userID uint) ([]entity.Voucher, error)
}

type voucherService struct {
	voucherRepo repository.VoucherRepository
	storeRepo   repository.StoreRepository
}

func NewVoucherService(voucherRepo repository.VoucherRepository, storeRepo repository.StoreRepository) VoucherService {
	return &voucherService{
		voucherRepo: voucherRepo,
		storeRepo:   storeRepo,
	}
}

func (s *voucherService) GetAvailableVouchers(storeID uint) ([]entity.Voucher, error) {
	vouchers, err := s.voucherRepo.FindByStoreID(storeID)
	if err != nil {
		return nil, errors.New("gagal mengambil data voucher")
	}

	var activeVouchers []entity.Voucher
	now := time.Now()
	for _, v := range vouchers {
		if v.ExpiredAt.After(now) {
			activeVouchers = append(activeVouchers, v)
		}
	}

	return activeVouchers, nil
}

func (s *voucherService) ClaimVoucher(userID uint, code string) error {
	voucher, err := s.voucherRepo.FindByCode(code)
	if err != nil {
		return errors.New("voucher tidak ditemukan")
	}

	if !voucher.ExpiredAt.After(time.Now()) {
		return errors.New("voucher kadaluarsa")
	}

	claimed, _ := s.voucherRepo.IsClaimedByUser(userID, voucher.ID)
	if claimed {
		return nil // idempotent success if already claimed
	}

	if err := s.voucherRepo.ClaimVoucher(userID, voucher.ID); err != nil {
		return errors.New("gagal mengklaim voucher")
	}

	return nil
}

func (s *voucherService) GetMyVouchers(userID uint) ([]entity.Voucher, error) {
	return s.voucherRepo.FindClaimedByUserID(userID)
}
