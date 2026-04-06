package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/Macklin06/optiroute/router/internal/models"
)

type DriverHandler struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}

func NewDriverHandler(redisClient *redis.Client, db *gorm.DB) *DriverHandler {
	return &DriverHandler{
		RedisClient: redisClient,
		DB:          db,
	}
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	var req models.LocationUpdate

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	locationKey := fmt.Sprintf("driver:location:%s", req.DriverID)
	locationValue := fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	err := h.RedisClient.Set(ctx, locationKey, locationValue, 60*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update location cache",
		})
		return
	}

	locationRecord := models.DriverLocation{
		DriverID:  req.DriverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		CreatedAt: time.Now(),
	}

	if result := h.DB.Create(&locationRecord); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to persist location",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"driver_id": req.DriverID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"message":   "location updated successfully",
	})
}
