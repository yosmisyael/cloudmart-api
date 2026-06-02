package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id uint) (*entity.User, error)
	FindAddresses(userID uint) ([]entity.Address, error)
	CreateAddress(address *entity.Address) error
	UpdateAddress(address *entity.Address) error
	DeleteAddress(addressID, userID uint) error
	SetDefaultAddress(addressID, userID uint) error
	UpdateRole(userID uint, role string) error
	UpdateProfile(userID uint, name, phone string) error
	UpdatePassword(userID uint, hashedPassword string) error
	UpdateAvatarURL(userID uint, url string) error
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

func (r *userRepository) UpdateAddress(address *entity.Address) error {
	return r.db.Save(address).Error
}

func (r *userRepository) DeleteAddress(addressID, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", addressID, userID).Delete(&entity.Address{}).Error
}

func (r *userRepository) SetDefaultAddress(addressID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}

		res := tx.Model(&entity.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (r *userRepository) UpdateRole(userID uint, role string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", userID).Update("role", role).Error
}

func (r *userRepository) UpdateProfile(userID uint, name, phone string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"name":  name,
		"phone": phone,
	}).Error
}

func (r *userRepository) UpdatePassword(userID uint, hashedPassword string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *userRepository) UpdateAvatarURL(userID uint, url string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", userID).Update("avatar_url", url).Error
}
