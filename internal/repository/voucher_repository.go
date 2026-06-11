package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type VoucherRepository interface {
	FindByStoreID(storeID uint) ([]entity.Voucher, error)
	FindByID(id uint) (*entity.Voucher, error)
	Create(v *entity.Voucher) error
	Update(v *entity.Voucher) error
	Delete(id uint) error
	FindByCode(code string) (*entity.Voucher, error)
	ClaimVoucher(userID, voucherID uint) error
	FindClaimedByUserID(userID uint) ([]entity.Voucher, error)
	IsClaimedByUser(userID, voucherID uint) (bool, error)
}

type voucherRepository struct {
	db *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) VoucherRepository {
	return &voucherRepository{db}
}

func (r *voucherRepository) FindByStoreID(storeID uint) ([]entity.Voucher, error) {
	var vouchers []entity.Voucher
	err := r.db.Where("store_id = ?", storeID).Find(&vouchers).Error
	return vouchers, err
}

func (r *voucherRepository) FindByID(id uint) (*entity.Voucher, error) {
	var voucher entity.Voucher
	err := r.db.First(&voucher, id).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *voucherRepository) Create(v *entity.Voucher) error {
	return r.db.Create(v).Error
}

func (r *voucherRepository) Update(v *entity.Voucher) error {
	return r.db.Save(v).Error
}

func (r *voucherRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Voucher{}, id).Error
}

func (r *voucherRepository) FindByCode(code string) (*entity.Voucher, error) {
	var voucher entity.Voucher
	err := r.db.Where("code ILIKE ?", code).First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *voucherRepository) ClaimVoucher(userID, voucherID uint) error {
	return r.db.Exec("INSERT INTO user_vouchers (user_id, voucher_id) VALUES (?, ?)", userID, voucherID).Error
}

func (r *voucherRepository) FindClaimedByUserID(userID uint) ([]entity.Voucher, error) {
	var vouchers []entity.Voucher
	err := r.db.Joins("JOIN user_vouchers ON user_vouchers.voucher_id = vouchers.id").
		Where("user_vouchers.user_id = ?", userID).
		Find(&vouchers).Error
	return vouchers, err
}

func (r *voucherRepository) IsClaimedByUser(userID, voucherID uint) (bool, error) {
	var count int64
	err := r.db.Table("user_vouchers").
		Where("user_id = ? AND voucher_id = ?", userID, voucherID).
		Count(&count).Error
	return count > 0, err
}
