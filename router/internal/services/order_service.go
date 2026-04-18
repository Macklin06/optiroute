package services

import (
	"fmt"

	"github.com/Macklin06/optiroute/router/internal/models"
	"gorm.io/gorm"
)

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{DB: db}
}

func (s *OrderService) CreateOrder(req models.CreateOrderRequest) (*models.Order, error) {
	zoneID := s.assignZone(req.Latitude, req.Longitude)

	order := models.Order{
		CustomerID: req.CustomerID,
		ZoneID:     zoneID,
		Status:     "pending",
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	if result := s.DB.Create(&order); result.Error != nil {
		return nil, fmt.Errorf("failed to create order: %w", result.Error)
	}

	return &order, nil

}

func (s *OrderService) GetPendingOrdersByZone(zoneID string) ([]models.Order, error) {
	var orders []models.Order

	result := s.DB.
		Where("zone_id = ? AND status = ?", zoneID, "pending").
		Order("created_at ASC").
		Find(&orders)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch orders for zone %s: %w", zoneID, result.Error)
	}

	return orders, nil
}

func (s *OrderService) assignZone(lat, lng float64) string {
	if lat >= 12.97 && lng >= 77.59 {
		return "zone_northeast"
	} else if lat >= 12.97 && lng < 77.59 {
		return "zone_northwest"
	} else if lat < 12.97 && lng >= 77.59 {
		return "zone_southeast"
	}
	return "zone_southwest"
}
