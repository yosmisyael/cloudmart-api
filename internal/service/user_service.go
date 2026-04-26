package service

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type UserService interface {
	GetProfile(userID uint) (*entity.User, error)
	GetAddresses(userID uint) ([]entity.Address, error)
	CreateAddress(address *entity.Address) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetProfile(userID uint) (*entity.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) GetAddresses(userID uint) ([]entity.Address, error) {
	return s.userRepo.FindAddresses(userID)
}

func (s *userService) CreateAddress(address *entity.Address) error {
	return s.userRepo.CreateAddress(address)
}
