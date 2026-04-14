package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Macklin06/optiroute/router/internal/models"
	"github.com/Macklin06/optiroute/router/internal/services"
)

type DriverHandler struct {
	DriverService *services.DriverService
}

func NewDriverHandler(driverService *services.DriverService) *DriverHandler {
	return &DriverHandler{
		DriverService: driverService,
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

	if err := h.DriverService.UpdateDriverLocation(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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

func (h *DriverHandler) GetActiveDrivers(c *gin.Context) {
	drivers, err := h.DriverService.GetActiveDrivers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"drivers": drivers,
		"count":   len(drivers),
	})
}
