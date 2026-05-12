// Business logic for driver operations
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/Macklin06/optiroute/router/internal/models"
)

// DriverService has Redis cache and database
type DriverService struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}

// NewDriverService creates a new service
func NewDriverService(redisClient *redis.Client, db *gorm.DB) *DriverService {
	return &DriverService{
		RedisClient: redisClient,
		DB:          db,
	}
}

// UpdateDriverLocation updates a driver's location in both Redis cache and PostgreSQL database.
func (s *DriverService) UpdateDriverLocation(req models.LocationUpdate) error {
	// context.Background() = no timeout, infinite wait.
	ctx := context.Background()

	// Build Redis key using pattern: "driver:location:<driver_id>"
	locationKey := fmt.Sprintf("driver:location:%s", req.DriverID)

	// Build Redis value: comma-separated lat,lng
	locationValue := fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	// Write to Redis with 60-second TTL.
	// After 60s, Redis auto-deletes the key (keeps only fresh data).
	err := s.RedisClient.Set(ctx, locationKey, locationValue, 60*time.Second).Err()
	if err != nil {
		// Error wrapping with %w preserves the original error chain for debugging.
		return fmt.Errorf("redis write failed: %w", err)
	}

	// Create database record struct filled with location data.
	// CreatedAt is set to current timestamp (server time).
	record := models.DriverLocation{
		DriverID:  req.DriverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		CreatedAt: time.Now(),
	}

	// Insert record into PostgreSQL.
	if result := s.DB.Create(&record); result.Error != nil {
		// Database write failed: wrap error and return.
		// Handler will catch this and return HTTP 500.
		return fmt.Errorf("postgresql write failed: %w", result.Error)
	}

	if err := s.PublishLocationUpdate(req); err != nil {
		log.Printf("Warning: pub/sub publish failed: %v", err)
		// NOT returning error here — core work is done
		// Location is saved in Redis and PostgreSQL
		// Pub/sub is a notification, not core functionality
	}

	// Both writes succeeded: return nil (no error).
	return nil
}

// GetActiveDrivers retrieves all currently active drivers from Redis cache.
func (s *DriverService) GetActiveDrivers() ([]models.DriverLocationResponse, error) {
	ctx := context.Background()

	// Query Redis for all keys matching pattern "driver:location:*".
	keys, err := s.RedisClient.Keys(ctx, "driver:location:*").Result()
	if err != nil {
		// Redis connection failed or other critical error.
		return nil, fmt.Errorf("failed to fetch driver keys from redis: %w", err)
	}

	// Slice to accumulate results.
	// Started empty and grows with each driver appended.
	var drivers []models.DriverLocationResponse

	// Iterate through each driver key.
	// _ discards the index (we don't need it; only the key value matters).
	for _, key := range keys {
		// Get the value (coordinates) for this driver from Redis.
		// Example: key="driver:location:driver_001" -> val="12.971600,77.594600"
		val, err := s.RedisClient.Get(ctx, key).Result()
		if err != nil {
			// Key might have expired or other read error.
			// continue skips this driver and moves to next (graceful degradation).
			continue
		}

		// Parse comma-separated coordinates string back into floats.
		// fmt.Sscanf = "scan formatted string" (inverse of fmt.Sprintf).
		var lat, lng float64
		fmt.Sscanf(val, "%f,%f", &lat, &lng)

		// Extract driver ID from the key.
		driverID := key[len("driver:location:"):]

		// Append driver to results slice.
		// Each iteration grows the slice by 1 element.
		drivers = append(drivers, models.DriverLocationResponse{
			DriverID:  driverID,
			Latitude:  lat,
			Longitude: lng,
		})
	}

	return drivers, nil
}

func (s *DriverService) PublishLocationUpdate(req models.LocationUpdate) error {
	ctx := context.Background()

	message := fmt.Sprintf(
		`{"driver_id":"%s","latitude":%f,"longitude":%f}`,
		req.DriverID, req.Latitude, req.Longitude,
	)

	err := s.RedisClient.Publish(ctx, "driver:updates", message).Err()
	if err != nil {
		return fmt.Errorf("failed to publish location update: %w", err)
	}

	return nil
}
