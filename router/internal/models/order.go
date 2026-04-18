package models

import "time"

type Order struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerID string    `json:"customer_id" gorm:"type:varchar(50);not null"`
	ZoneID     string    `json:"zone_id" gorm:"type:varchar(50);not null; index"`
	Status     string    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Latitude   float64   `json:"latitude" gorm:"not null"`
	Longitude  float64   `json:"longitude" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerID string  `json:"customer_id" binding:"required"`
	Latitude   float64 `json:"latitude"    binding:"required"`
	Longitude  float64 `json:"longitude"   binding:"required"`
}
