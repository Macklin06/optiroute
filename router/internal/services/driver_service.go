// Business logic for driver operations
package services

import (
	"context"
	"fmt"
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
// This implements the Write-Through Cache Pattern:
//  1. Write to cache (fast, latest state available instantly)
//  2. Write to database (durable, historical record for analytics)
//
// If either write fails, the entire operation fails and error is returned.
//
// Parameters:
//
//	req: LocationUpdate with DriverID, Latitude, Longitude
//
// Returns:
//
//	error: nil on success, or wrapped error from cache/db failures
func (s *DriverService) UpdateDriverLocation(req models.LocationUpdate) error {
	// context.Background() = no timeout, infinite wait.
	// In production, use context.WithTimeout() to prevent hanging requests.
	ctx := context.Background()

	// Build Redis key using pattern: "driver:location:<driver_id>"
	// Example: "driver:location:driver_001"
	// This pattern allows querying all drivers with KEYS "driver:location:*"
	locationKey := fmt.Sprintf("driver:location:%s", req.DriverID)

	// Build Redis value: comma-separated lat,lng
	// Example: "12.971600,77.594600"
	// Simple format that's fast to parse: fmt.Sscanf(val, "%f,%f", &lat, &lng)
	locationValue := fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	// Write to Redis with 60-second TTL.
	// After 60s, Redis auto-deletes the key (keeps only fresh data).
	// TTL ensures stale driver locations don't stay in cache forever.
	err := s.RedisClient.Set(ctx, locationKey, locationValue, 60*time.Second).Err()
	if err != nil {
		// Error wrapping with %w preserves the original error chain for debugging.
		// Example error message: "redis write failed: connection refused"
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
	// GORM.Create() runs: INSERT INTO driver_locations (driver_id, latitude, longitude, created_at) VALUES  (...)
	// result.Error contains any database error (constraint violation, connection lost, etc.).
	if result := s.DB.Create(&record); result.Error != nil {
		// Database write failed: wrap error and return.
		// Handler will catch this and return HTTP 500.
		return fmt.Errorf("postgresql write failed: %w", result.Error)
	}

	// Both writes succeeded: return nil (no error).
	return nil
}

// GetActiveDrivers retrieves all currently active drivers from Redis cache.
// A driver is "active" if their location key exists in Redis (within 60-second TTL).
// This is fast compared to querying the full database (no WHERE clause scan).
//
// Returns:
//
//	slice of DriverLocationResponse: list of active drivers with coordinates
//	error: if Redis Keys() or Get() fails
func (s *DriverService) GetActiveDrivers() ([]models.DriverLocationResponse, error) {
	ctx := context.Background()

	// Query Redis for all keys matching pattern "driver:location:*".
	// Examples returned: ["driver:location:driver_001", "driver:location:driver_002", ...]
	// This pattern scan is O(N) where N = total Redis keys, but fast enough for typical scale.
	// In production, use SCAN for incremental iteration (doesn't block Redis).
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
		// Input: "12.971600,77.594600"
		// Output: lat=12.9716, lng=77.5946
		var lat, lng float64
		fmt.Sscanf(val, "%f,%f", &lat, &lng)

		// Extract driver ID from the key.
		// key = "driver:location:driver_001"
		// len("driver:location:") = 16 characters
		// key[16:] = "driver_001" (slice from position 16 to end)
		driverID := key[len("driver:location:"):]

		// Append driver to results slice.
		// Each iteration grows the slice by 1 element.
		drivers = append(drivers, models.DriverLocationResponse{
			DriverID:  driverID,
			Latitude:  lat,
			Longitude: lng,
		})
	}

	// Return populated slice and nil error (success).
	return drivers, nil
}
