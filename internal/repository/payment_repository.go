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
	GetConfigByIDAndStoreID(configID, storeID uint) (*entity.PaymentConfiguration, error)
	UpdateConfig(config *entity.PaymentConfiguration) error
	GetBankByIDAndStoreID(bankID, storeID uint) (*entity.PaymentBank, error)
	UpdateBank(bank *entity.PaymentBank) error
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

func (r *paymentRepository) GetConfigByIDAndStoreID(configID, storeID uint) (*entity.PaymentConfiguration, error) {
	var config entity.PaymentConfiguration
	result := r.db.Where("id = ? AND store_id = ?", configID, storeID).First(&config)
	if result.Error != nil {
		return nil, result.Error
	}
	return &config, nil
}

func (r *paymentRepository) UpdateConfig(config *entity.PaymentConfiguration) error {
	return r.db.Save(config).Error
}

func (r *paymentRepository) GetBankByIDAndStoreID(bankID, storeID uint) (*entity.PaymentBank, error) {
	var bank entity.PaymentBank
	result := r.db.
		Joins("JOIN payment_configurations ON payment_configurations.id = payment_banks.payment_configuration_id").
		Where("payment_banks.id = ? AND payment_configurations.store_id = ?", bankID, storeID).
		First(&bank)
	if result.Error != nil {
		return nil, result.Error
	}
	return &bank, nil
}

func (r *paymentRepository) UpdateBank(bank *entity.PaymentBank) error {
	return r.db.Save(bank).Error
}
