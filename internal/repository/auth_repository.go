package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type AuthRepository interface {
	Create(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	FindByID(id uint) (*entity.User, error)
	UpdateRefreshToken(userID uint, token string) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateRefreshToken(userID uint, token string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", userID).Update("refresh_token", token).Error
}
