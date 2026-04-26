package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/yosmisyael/cloudmart-web-service/internal/config"
	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"github.com/yosmisyael/cloudmart-web-service/internal/handler"
	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
	"github.com/yosmisyael/cloudmart-web-service/internal/service"

	_ "github.com/yosmisyael/cloudmart-web-service/docs"
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
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Cart{},
		&entity.Store{},
		&entity.Product{},
		&entity.ProductVariant{},
		&entity.Category{},
		&entity.Order{},
		&entity.OrderItem{},
		&entity.Address{},
		&entity.Logistic{},
		&entity.LogisticService{},
		&entity.PaymentConfiguration{},
		&entity.PaymentBank{},
		&entity.Voucher{},
	)
	if err != nil {
		log.Fatalf("[Migration] Failed: %v", err)
	}
	config.RunSeeder(db)

	authRepo := repository.NewAuthRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(authRepo, cfg)
	catalogService := service.NewCatalogService(categoryRepo, productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)
	webhookService := service.NewWebhookService(orderRepo, cfg)
	userService := service.NewUserService(userRepo)

	app := fiber.New()
	app.Use(logger.New())

	handler.NewAuthHandler(app, authService)
	handler.NewCatalogHandler(app, catalogService)
	handler.NewCartHandler(app, cartService, cfg)
	handler.NewOrderHandler(app, orderService, cfg)
	handler.NewWebhookHandler(app, webhookService)
	handler.NewProfileHandler(app, userService, cfg)
	app.Get("/swagger/*", swagger.HandlerDefault)

	log.Fatal("[server] ", app.Listen(":"+cfg.Port))
}
