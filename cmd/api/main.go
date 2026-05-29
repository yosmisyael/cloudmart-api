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
		&entity.Review{},
		&entity.ReviewImage{},
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
	storeRepo := repository.NewStoreRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	logisticRepo := repository.NewLogisticRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	sellerOrderRepo := repository.NewSellerOrderRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	s3Client := config.NewS3Client(cfg)
	s3Service := service.NewS3Service(s3Client, cfg)
	paymentService := service.NewPaymentService(cfg)
	reviewService := service.NewReviewService(reviewRepo, orderRepo, productRepo, storeRepo, s3Service)

	authService := service.NewAuthService(authRepo, cfg)
	catalogService := service.NewCatalogService(categoryRepo, productRepo, logisticRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, authRepo, userRepo, logisticRepo, voucherRepo, paymentService)
	webhookService := service.NewWebhookService(orderRepo, cfg)
	userService := service.NewUserService(userRepo)
	voucherService := service.NewVoucherService(voucherRepo, storeRepo)
	sellerStoreService := service.NewSellerStoreService(storeRepo, userRepo)
	sellerCatalogService := service.NewSellerCatalogService(storeRepo, categoryRepo, productRepo, s3Service)
	sellerPromoService := service.NewSellerPromoService(storeRepo, voucherRepo, paymentRepo)
	sellerLogisticService := service.NewSellerLogisticService(logisticRepo)
	sellerOrderService := service.NewSellerOrderService(storeRepo, sellerOrderRepo)
	sellerDashboardService := service.NewSellerDashboardService(storeRepo, dashboardRepo)

	app := fiber.New()
	app.Use(logger.New())

	handler.NewAuthHandler(app, authService)
	handler.NewCatalogHandler(app, catalogService)
	handler.NewLogisticHandler(app, catalogService)
	handler.NewVoucherHandler(app, voucherService, cfg)
	handler.NewCartHandler(app, cartService, cfg)
	handler.NewOrderHandler(app, orderService, cfg)
	handler.NewWebhookHandler(app, webhookService)
	handler.NewProfileHandler(app, userService, cfg)
	handler.NewSellerStoreHandler(app, sellerStoreService, userRepo, cfg)
	handler.NewSellerCatalogHandler(app, sellerCatalogService, userRepo, cfg)
	handler.NewSellerPromoHandler(app, sellerPromoService, userRepo, cfg)
	handler.NewSellerLogisticHandler(app, sellerLogisticService, userRepo, cfg)
	handler.NewSellerOrderHandler(app, sellerOrderService, userRepo, cfg)
	handler.NewSellerDashboardHandler(app, sellerDashboardService, userRepo, cfg)
	handler.NewReviewHandler(app, reviewService, userRepo, cfg)
	app.Get("/swagger/*", swagger.HandlerDefault)

	log.Fatal("[server] ", app.Listen(":"+cfg.Port))
}
