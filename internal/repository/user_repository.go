package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id uint) (*entity.User, error)
	FindAddresses(userID uint) ([]entity.Address, error)
	CreateAddress(address *entity.Address) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAddresses(userID uint) ([]entity.Address, error) {
	var addresses []entity.Address
	err := r.db.Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}

func (r *userRepository) CreateAddress(address *entity.Address) error {
	return r.db.Create(address).Error
}
