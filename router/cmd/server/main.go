// Main router service
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/Macklin06/optiroute/router/config"
	"github.com/Macklin06/optiroute/router/internal/handlers"
	"github.com/Macklin06/optiroute/router/internal/models"
	"github.com/Macklin06/optiroute/router/internal/services"
)

func main() {
	// Connect to database and run migrations
	dbConfig := config.NewDatabaseConfig()
	db := config.ConnectDatabase(dbConfig)

	db.AutoMigrate(
		&models.Driver{},
		&models.DriverLocation{},
		&models.Order{},
	)

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Check if Redis is running
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis connected successfully")

	// Setup service and handler layers
	driverService := services.NewDriverService(redisClient, db)
	driverHandler := handlers.NewDriverHandler(driverService)
	orderService := services.NewOrderService(db)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Create router and setup routes
	router := gin.Default()

	router.GET("/health", healthHandler)

	v1 := router.Group("/api/v1")
	{
		drivers := v1.Group("/drivers")
		{
			drivers.PUT("/location", driverHandler.UpdateLocation)
			drivers.GET("/active", driverHandler.GetActiveDrivers)
		}

		orders := v1.Group("/orders")
		{
			orders.POST("/", orderHandler.CreateOrder)
			orders.GET("/zone/:zone_id", orderHandler.GetPendingOrdersByZone)
		}
	}

	// Start server on port 8080
	fmt.Println("OptiRoute router starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// Health check endpoint
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "optiroute-router",
	})
}
