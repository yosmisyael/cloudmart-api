package config

import (
	"log"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB) {
	var count int64
	db.Model(&entity.User{}).Count(&count)
	if count > 0 {
		log.Println("database not empty, seeding skipped.")
		return
	}

	log.Println("Seeding database...")

	user := entity.User{
		Name:     "Agung",
		Email:    "agung@cloudmart.com",
		Password: "hashedpassword123",
		Phone:    "08123456789",
		Role:     "customer",
	}
	db.Create(&user)

	store := entity.Store{
		UserID: user.ID,
		Name:   "Studio Tropik",
	}
	db.Create(&store)

	category := entity.Category{
		Name:      "Tas Pria",
		IsDefault: true,
	}
	db.Create(&category)

	product := entity.Product{
		StoreID:     store.ID,
		CategoryID:  category.ID,
		Name:        "Nevada Ransel Backpack Elegan",
		Description: "Tas andalan buat ngantor dan kuliah",
	}
	db.Create(&product)

	variant := entity.ProductVariant{
		ProductID: product.ID,
		SKU:       "SKU-NVDA-BRN-18",
		Color:     "Brown",
		Size:      "18 INCH",
		Price:     450000,
		Stock:     10,
	}
	db.Create(&variant)

	log.Println("Seeder completed")
}
