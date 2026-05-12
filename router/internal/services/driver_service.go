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
	CB          *CircuitBreaker
}

func NewDriverService(redisClient *redis.Client, db *gorm.DB) *DriverService {
	return &DriverService{
		RedisClient: redisClient,
		DB:          db,
		CB:          NewCircuitBreaker(3, 30*time.Second),
	}
}

// UpdateDriverLocation updates driver location data.
// Flow:
// 1. Try Redis write if circuit breaker allows it
// 2. Always write to PostgreSQL
// 3. Publish location event to Redis pub/sub
func (s *DriverService) UpdateDriverLocation(req models.LocationUpdate) error {

	// Create PostgreSQL record struct from incoming request data.
	// This becomes one permanent row in driver_locations table.
	record := models.DriverLocation{
		DriverID:  req.DriverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		CreatedAt: time.Now(), // Current server timestamp
	}

	// Ask circuit breaker:
	// "Is Redis healthy enough to attempt requests?"
	if s.CB.CanRequest() {

		// Create base context for Redis operation.
		// Context carries timeout/cancellation metadata.
		ctx := context.Background()

		// Redis key pattern:
		// driver:location:<driver_id>
		// Example: driver:location:d1
		locationKey := fmt.Sprintf("driver:location:%s", req.DriverID)

		// Redis value:
		// latitude,longitude
		// Example: 12.971600,77.594600
		locationValue := fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

		// Write latest location to Redis with 60-second TTL.
		// TTL auto-removes stale offline drivers.
		err := s.RedisClient.Set(
			ctx,
			locationKey,
			locationValue,
			60*time.Second,
		).Err()

		// Redis write failed.
		if err != nil {

			// Tell circuit breaker:
			// "Redis failed one more time."
			s.CB.RecordFailure()

			// Log warning but DO NOT fail request.
			// PostgreSQL fallback still keeps system working.
			log.Printf(
				"Redis failed, fallback to PostgreSQL only (failures: %d)",
				s.CB.failureCount,
			)

		} else {

			// Redis succeeded.
			// Reset failure counter and keep circuit closed.
			s.CB.RecordSuccess()
		}

	} else {

		// Circuit breaker is OPEN.
		// Skip Redis entirely to avoid hammering failing service.
		// Directly continue with PostgreSQL write.
		log.Println(
			"Circuit OPEN: skipping Redis, writing to PostgreSQL only",
		)
	}

	// Always write to PostgreSQL.
	// PostgreSQL is the permanent source of truth.
	if result := s.DB.Create(&record); result.Error != nil {

		// Database failure is CRITICAL.
		// Unlike Redis failure, this should fail the request.
		return fmt.Errorf(
			"postgresql write failed: %w",
			result.Error,
		)
	}

	// Publish location update event to Redis pub/sub.
	// Python ML service subscribes to these events asynchronously.
	if err := s.PublishLocationUpdate(req); err != nil {

		// Pub/sub failure is NON-FATAL.
		// Core writes already succeeded.
		// Only the notification failed.
		log.Printf(
			"Warning: pub/sub publish failed: %v",
			err,
		)
	}

	// Everything important succeeded.
	// Return nil = no error.
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
