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
	dbConfig := config.NewDatabaseConfig()
	db := config.ConnectDatabase(dbConfig)

	db.AutoMigrate(
		&models.Driver{},
		&models.DriverLocation{},
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis connected successfully")

	driverService := services.NewDriverService(redisClient, db)
	driverHandler := handlers.NewDriverHandler(driverService)

	router := gin.Default()

	router.GET("/health", healthHandler)

	v1 := router.Group("/api/v1")
	{
		drivers := v1.Group("/drivers")
		{
			drivers.PUT("/location", driverHandler.UpdateLocation)
			drivers.GET("/active", driverHandler.GetActiveDrivers)
		}
	}

	fmt.Println("OptiRoute router starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "optiroute-router",
	})
}
