// Data structures for the app
package models

import "time"

// Driver represents a fleet member
type Driver struct {
	// ID: unique driver identifier (e.g., "driver_001")
	ID string `json:"id" gorm:"primaryKey;type:varchar(50)"`

	// IsActive: whether driver is available
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DriverLocation stores a location snapshot (one row per update)
type DriverLocation struct {
	// ID: auto-generated row ID
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// DriverID: which driver this location is for
	DriverID  string    `json:"driver_id" gorm:"type:varchar(50);not null;index"`
	Latitude  float64   `json:"latitude" gorm:"not null"`
	Longitude float64   `json:"longitude" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// LocationUpdate is what the API request should contain
type LocationUpdate struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// DriverLocationResponse is what we return to the client
type DriverLocationResponse struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
