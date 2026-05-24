package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	FindByStoreID(storeID uint) ([]entity.PaymentConfiguration, error)
	FindByID(id uint) (*entity.PaymentConfiguration, error)
	Create(p *entity.PaymentConfiguration) error
	Delete(id uint) error
	CreateBank(b *entity.PaymentBank) error
	DeleteBank(id uint) error
	FindBankByID(id uint) (*entity.PaymentBank, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db}
}

func (r *paymentRepository) FindByStoreID(storeID uint) ([]entity.PaymentConfiguration, error) {
	var configs []entity.PaymentConfiguration
	err := r.db.Preload("Banks").Where("store_id = ?", storeID).Find(&configs).Error
	return configs, err
}

func (r *paymentRepository) FindByID(id uint) (*entity.PaymentConfiguration, error) {
	var config entity.PaymentConfiguration
	err := r.db.Preload("Banks").First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *paymentRepository) Create(p *entity.PaymentConfiguration) error {
	return r.db.Create(p).Error
}

func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&entity.PaymentConfiguration{}, id).Error
}

func (r *paymentRepository) CreateBank(b *entity.PaymentBank) error {
	return r.db.Create(b).Error
}

func (r *paymentRepository) DeleteBank(id uint) error {
	return r.db.Delete(&entity.PaymentBank{}, id).Error
}

func (r *paymentRepository) FindBankByID(id uint) (*entity.PaymentBank, error) {
	var bank entity.PaymentBank
	err := r.db.First(&bank, id).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}
