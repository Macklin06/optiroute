// Handlers for driver endpoints
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Macklin06/optiroute/router/internal/models"
	"github.com/Macklin06/optiroute/router/internal/services"
)

// DriverHandler holds the service layer reference
type DriverHandler struct {
	DriverService *services.DriverService
}

// NewDriverHandler creates a new handler
func NewDriverHandler(driverService *services.DriverService) *DriverHandler {
	return &DriverHandler{
		DriverService: driverService,
	}
}

// UpdateLocation updates a driver's location (PUT /api/v1/drivers/location)
func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	// Parse and validate JSON request
	var req models.LocationUpdate

	if err := c.ShouldBindJSON(&req); err != nil {
		// Invalid request
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Call service to update location
	if err := h.DriverService.UpdateDriverLocation(req); err != nil {
		// Server error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"driver_id": req.DriverID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"message":   "location updated successfully",
	})
}

// GetActiveDrivers retrieves all active drivers from cache (GET /api/v1/drivers/active)
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
