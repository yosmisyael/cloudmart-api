package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type LogisticRepository interface {
	FindAll() ([]entity.Logistic, error)
	FindByID(id uint) (*entity.Logistic, error)
	Create(l *entity.Logistic) error
	Update(l *entity.Logistic) error
	Delete(id uint) error
	CreateService(s *entity.LogisticService) error
	UpdateService(s *entity.LogisticService) error
	DeleteService(id uint) error
	FindServiceByID(id uint) (*entity.LogisticService, error)
}

type logisticRepository struct {
	db *gorm.DB
}

func NewLogisticRepository(db *gorm.DB) LogisticRepository {
	return &logisticRepository{db}
}

func (r *logisticRepository) FindAll() ([]entity.Logistic, error) {
	var logistics []entity.Logistic
	err := r.db.Preload("Services").Find(&logistics).Error
	return logistics, err
}

func (r *logisticRepository) FindByID(id uint) (*entity.Logistic, error) {
	var logistic entity.Logistic
	err := r.db.Preload("Services").First(&logistic, id).Error
	if err != nil {
		return nil, err
	}
	return &logistic, nil
}

func (r *logisticRepository) Create(l *entity.Logistic) error {
	return r.db.Create(l).Error
}

func (r *logisticRepository) Update(l *entity.Logistic) error {
	return r.db.Save(l).Error
}

func (r *logisticRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Logistic{}, id).Error
}

func (r *logisticRepository) CreateService(s *entity.LogisticService) error {
	return r.db.Create(s).Error
}

func (r *logisticRepository) UpdateService(s *entity.LogisticService) error {
	return r.db.Save(s).Error
}

func (r *logisticRepository) DeleteService(id uint) error {
	return r.db.Delete(&entity.LogisticService{}, id).Error
}

func (r *logisticRepository) FindServiceByID(id uint) (*entity.LogisticService, error) {
	var svc entity.LogisticService
	err := r.db.First(&svc, id).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}
