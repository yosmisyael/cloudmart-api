package repository

import (
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

type StoreRepository interface {
	FindByUserID(userID uint) (*entity.Store, error)
	Create(store *entity.Store) error
	Update(store *entity.Store) error
	UpdateLogoURL(storeID uint, url string) error
}

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{db}
}

func (r *storeRepository) FindByUserID(userID uint) (*entity.Store, error) {
	var store entity.Store
	err := r.db.Where("user_id = ?", userID).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) Create(store *entity.Store) error {
	return r.db.Create(store).Error
}

func (r *storeRepository) Update(store *entity.Store) error {
	return r.db.Save(store).Error
}

func (r *storeRepository) UpdateLogoURL(storeID uint, url string) error {
	return r.db.Model(&entity.Store{}).Where("id = ?", storeID).Update("logo_url", url).Error
}
