package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret         string
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3BucketName       string
	Port              string
	FrontendPaymentRedirect string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "cloudmart"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
		AWSRegion:          getEnv("AWS_REGION", "ap-southeast-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:       getEnv("S3_BUCKET_NAME", ""),
		Port:              getEnv("APP_PORT", "8080"),
		FrontendPaymentRedirect: getEnv("FRONTEND_PAYMENT_REDIRECT", "http://localhost:5173/payment-success"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
