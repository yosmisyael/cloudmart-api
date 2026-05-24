package service

import (
	"errors"

	"github.com/yosmisyael/cloudmart-web-service/internal/repository"
)

type DashboardSummary struct {
	TotalOrders   int64   `json:"total_orders"`
	PendingOrders int64   `json:"pending_orders"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalProducts int64   `json:"total_products"`
	LowStockItems int64   `json:"low_stock_items"`
}

type SellerDashboardService interface {
	GetSummary(userID uint) (*DashboardSummary, error)
}

type sellerDashboardService struct {
	storeRepo     repository.StoreRepository
	dashboardRepo repository.DashboardRepository
}

func NewSellerDashboardService(storeRepo repository.StoreRepository, dashboardRepo repository.DashboardRepository) SellerDashboardService {
	return &sellerDashboardService{storeRepo, dashboardRepo}
}

func (s *sellerDashboardService) GetSummary(userID uint) (*DashboardSummary, error) {
	store, err := s.storeRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	totalOrders, err := s.dashboardRepo.CountOrdersByStore(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data dashboard")
	}

	pendingOrders, err := s.dashboardRepo.CountPendingOrdersByStore(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data dashboard")
	}

	totalRevenue, err := s.dashboardRepo.SumRevenueByStore(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data dashboard")
	}

	totalProducts, err := s.dashboardRepo.CountProductsByStore(store.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data dashboard")
	}

	lowStockItems, err := s.dashboardRepo.CountLowStockByStore(store.ID, 10)
	if err != nil {
		return nil, errors.New("gagal mengambil data dashboard")
	}

	return &DashboardSummary{
		TotalOrders:   totalOrders,
		PendingOrders: pendingOrders,
		TotalRevenue:  totalRevenue,
		TotalProducts: totalProducts,
		LowStockItems: lowStockItems,
	}, nil
}
