package service

import (
	"context"
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
	UploadAvatar(ctx context.Context, userID uint, data []byte, filename, contentType string) (string, error)
	DeleteAvatar(ctx context.Context, userID uint) error
}

type userService struct {
	userRepo repository.UserRepository
	s3Svc    S3Service
}

func NewUserService(userRepo repository.UserRepository, s3Svc S3Service) UserService {
	return &userService{userRepo: userRepo, s3Svc: s3Svc}
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

func (s *userService) UploadAvatar(ctx context.Context, userID uint, data []byte, filename, contentType string) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	if user.AvatarURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(user.AvatarURL))
	}

	url, err := s.s3Svc.UploadFile(ctx, "avatars", filename, data, contentType)
	if err != nil {
		return "", errors.New("gagal upload avatar")
	}

	if err := s.userRepo.UpdateAvatarURL(userID, url); err != nil {
		// DB failed — roll back the S3 upload
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(url))
		return "", errors.New("gagal update url avatar")
	}

	return url, nil
}

func (s *userService) DeleteAvatar(ctx context.Context, userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	if user.AvatarURL != "" {
		_ = s.s3Svc.DeleteFile(ctx, s.s3Svc.ExtractKeyFromURL(user.AvatarURL))
		_ = s.userRepo.UpdateAvatarURL(userID, "")
	}

	return nil
}
