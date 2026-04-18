package handlers

import (
	"net/http"

	"github.com/Macklin06/optiroute/router/internal/models"
	"github.com/Macklin06/optiroute/router/internal/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	OrderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		OrderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	order, err := h.OrderService.CreateOrder(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetPendingOrdersByZone(c *gin.Context) {
	zoneID := c.Param("zone_id")

	orders, err := h.OrderService.GetPendingOrdersByZone(zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"zone_id": zoneID,
		"orders":  orders,
		"count":   len(orders),
	})
}
