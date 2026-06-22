package services

import (
	"fmt"
	"math"

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

// bangaloreZones lists the 8 named zones the XGBoost demand model was
// trained on. Order assignment must use these exact names, not arbitrary
// quadrants, or the predictor's LabelEncoder will reject the zone_id.
var bangaloreZones = []struct {
	name   string
	lat    float64
	lng    float64
	radius float64
}{
	{"koramangala", 12.9352, 77.6245, 0.03},
	{"indiranagar", 12.9784, 77.6408, 0.025},
	{"whitefield", 12.9698, 77.7499, 0.04},
	{"marathahalli", 12.9591, 77.6974, 0.03},
	{"hsr_layout", 12.9116, 77.6389, 0.025},
	{"jp_nagar", 12.9102, 77.5856, 0.03},
	{"electronic_city", 12.8399, 77.6770, 0.04},
	{"hebbal", 13.0353, 77.5972, 0.03},
}

// assignZone maps order coordinates to the nearest named zone using a
// simple bounding-box check. Falls back to koramangala (highest base
// demand zone) if coordinates fall outside every defined box.
func (s *OrderService) assignZone(lat, lng float64) string {
	for _, z := range bangaloreZones {
		if math.Abs(lat-z.lat) <= z.radius && math.Abs(lng-z.lng) <= z.radius {
			return z.name
		}
	}
	return "koramangala"
}
