package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(userID uint) (*entity.User, error)
	GetAddresses(userID uint) ([]entity.Address, error)
	CreateAddress(address *entity.Address) error
	UpdateAddress(address *entity.Address) error
	DeleteAddress(addressID, userID uint) error
	SetDefaultAddress(addressID, userID uint) error
	UpdateProfile(userID uint, name, phone string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
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

func (s *userService) UpdateAddress(address *entity.Address) error {
	return s.userRepo.UpdateAddress(address)
}

func (s *userService) DeleteAddress(addressID, userID uint) error {
	return s.userRepo.DeleteAddress(addressID, userID)
}

func (s *userService) SetDefaultAddress(addressID, userID uint) error {
	return s.userRepo.SetDefaultAddress(addressID, userID)
}

func (s *userService) UpdateProfile(userID uint, name, phone string) error {
	return s.userRepo.UpdateProfile(userID, name, phone)
}

func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("password saat ini salah")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return errors.New("gagal memproses password baru")
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}
