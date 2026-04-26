package config

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatalf("[Database] Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		log.Fatalf("[Database] Failed to prepare pool connection: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("[Database] Connection pool initialized")
	return db
}
