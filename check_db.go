package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=127.0.0.1 user=postgres password=123 dbname=cloudmart port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var orders []entity.Order
	db.Where("user_id = ?", 3).
		Preload("OrderItems").
		Preload("OrderItems.Variant").
		Preload("OrderItems.Variant.Product").
		Preload("OrderItems.Variant.Product.Store").
		Order("created_at DESC").
		Find(&orders)

	b, _ := json.MarshalIndent(orders[0], "", "  ")
	fmt.Println(string(b))
}
