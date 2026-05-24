package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type LogisticServiceInput struct {
	Name      string
	BasePrice float64
}

type SellerLogisticService interface {
	GetLogistics() ([]entity.Logistic, error)
	CreateLogistic(name string) (*entity.Logistic, error)
	UpdateLogistic(id uint, name string) (*entity.Logistic, error)
	DeleteLogistic(id uint) error
	AddService(logisticID uint, req LogisticServiceInput) (*entity.LogisticService, error)
	UpdateService(serviceID uint, req LogisticServiceInput) (*entity.LogisticService, error)
	DeleteService(serviceID uint) error
}

type sellerLogisticService struct {
	logisticRepo repository.LogisticRepository
}

func NewSellerLogisticService(logisticRepo repository.LogisticRepository) SellerLogisticService {
	return &sellerLogisticService{logisticRepo}
}

func (s *sellerLogisticService) GetLogistics() ([]entity.Logistic, error) {
	logistics, err := s.logisticRepo.FindAll()
	if err != nil {
		return nil, errors.New("gagal mengambil data logistik")
	}
	return logistics, nil
}

func (s *sellerLogisticService) CreateLogistic(name string) (*entity.Logistic, error) {
	logistic := &entity.Logistic{
		Name: name,
	}
	if err := s.logisticRepo.Create(logistic); err != nil {
		return nil, errors.New("gagal membuat logistik")
	}
	return logistic, nil
}

func (s *sellerLogisticService) UpdateLogistic(id uint, name string) (*entity.Logistic, error) {
	logistic, err := s.logisticRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("logistik tidak ditemukan")
	}

	logistic.Name = name
	if err := s.logisticRepo.Update(logistic); err != nil {
		return nil, errors.New("gagal mengupdate logistik")
	}
	return logistic, nil
}

func (s *sellerLogisticService) DeleteLogistic(id uint) error {
	_, err := s.logisticRepo.FindByID(id)
	if err != nil {
		return errors.New("logistik tidak ditemukan")
	}

	if err := s.logisticRepo.Delete(id); err != nil {
		return errors.New("gagal menghapus logistik")
	}
	return nil
}

func (s *sellerLogisticService) AddService(logisticID uint, req LogisticServiceInput) (*entity.LogisticService, error) {
	_, err := s.logisticRepo.FindByID(logisticID)
	if err != nil {
		return nil, errors.New("logistik tidak ditemukan")
	}

	svc := &entity.LogisticService{
		LogisticID: logisticID,
		Name:       req.Name,
		BasePrice:  req.BasePrice,
	}

	if err := s.logisticRepo.CreateService(svc); err != nil {
		return nil, errors.New("gagal menambahkan layanan")
	}
	return svc, nil
}

func (s *sellerLogisticService) UpdateService(serviceID uint, req LogisticServiceInput) (*entity.LogisticService, error) {
	svc, err := s.logisticRepo.FindServiceByID(serviceID)
	if err != nil {
		return nil, errors.New("layanan tidak ditemukan")
	}

	svc.Name = req.Name
	svc.BasePrice = req.BasePrice

	if err := s.logisticRepo.UpdateService(svc); err != nil {
		return nil, errors.New("gagal mengupdate layanan")
	}
	return svc, nil
}

func (s *sellerLogisticService) DeleteService(serviceID uint) error {
	_, err := s.logisticRepo.FindServiceByID(serviceID)
	if err != nil {
		return errors.New("layanan tidak ditemukan")
	}

	if err := s.logisticRepo.DeleteService(serviceID); err != nil {
		return errors.New("gagal menghapus layanan")
	}
	return nil
}
