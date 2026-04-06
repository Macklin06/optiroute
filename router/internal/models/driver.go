package models

import "time"

type Driver struct {
	ID        string    `json:"id"        gorm:"primaryKey;type:varchar(50)"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DriverLocation struct {
	ID        uint      `json:"id"        gorm:"primaryKey;autoIncrement"`
	DriverID  string    `json:"driver_id" gorm:"type:varchar(50);not null:index"`
	Latitude  float64   `json:"latitude"  gorm:"not null"`
	Longitude float64   `json:"longitude" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

type LocationUpdate struct {
	DriverID  string  `json:"driver_id"  binding:"required"`
	Latitude  float64 `json:"latitude"   binding:"required"`
	Longitude float64 `json:"longitude"  binding:"required"`
}
