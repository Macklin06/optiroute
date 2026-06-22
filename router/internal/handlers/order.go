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

// CreateOrder godoc
// @Summary Create order
// @Description Creates a new order in a delivery zone
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body models.CreateOrderRequest true "Create Order Request"
// @Success 201 {object} models.Order
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/ [post]
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

// GetPendingOrdersByZone godoc
// @Summary Get pending orders by zone
// @Description Returns all pending orders for a specific zone
// @Tags Orders
// @Produce json
// @Param zone_id path string true "Zone ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/zone/{zone_id} [get]
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
