package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
)

func main() {
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBName,
		cfg.DBPassword,
	)
	db := config.InitDatabase(dsn)

	// Create a test voucher
	v := entity.Voucher{
		StoreID:   1, // Assign to store 1 (Nike)
		Code:      "DISKON10",
		Type:      "percentage",
		Amount:    10, // 10%
		Max:       50000,
		ExpiredAt: time.Now().AddDate(0, 1, 0),
	}
	
	if err := db.Create(&v).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created voucher: %s\n", v.Code)
}
