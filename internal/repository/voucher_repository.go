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
