package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/Macklin06/optiroute/router/internal/models"
)

type DriverService struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}

func NewDriverService(redisClient *redis.Client, db *gorm.DB) *DriverService {
	return &DriverService{
		RedisClient: redisClient,
		DB:          db,
	}
}

func (s *DriverService) UpdateDriverLocation(req models.LocationUpdate) error {
	ctx := context.Background()

	locationKey := fmt.Sprintf("driver:location:%s", req.DriverID)
	locationValue := fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	err := s.RedisClient.Set(ctx, locationKey, locationValue, 60*time.Second).Err()
	if err != nil {
		return fmt.Errorf("redis write failed: %w", err)
	}

	record := models.DriverLocation{
		DriverID:  req.DriverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		CreatedAt: time.Now(),
	}

	if result := s.DB.Create(&record); result.Error != nil {
		return fmt.Errorf("postgresql write failed: %w", result.Error)
	}

	return nil
}

func (s *DriverService) GetActiveDrivers() ([]models.DriverLocationResponse, error) {
	ctx := context.Background()

	keys, err := s.RedisClient.Keys(ctx, "driver:location:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch driver keys from redis: %w", err)
	}

	var drivers []models.DriverLocationResponse

	for _, key := range keys {
		val, err := s.RedisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var lat, lng float64
		fmt.Sscanf(val, "%f,%f", &lat, &lng)

		driverID := key[len("driver:location:"):]

		drivers = append(drivers, models.DriverLocationResponse{
			DriverID:  driverID,
			Latitude:  lat,
			Longitude: lng,
		})
	}

	return drivers, nil
}
