package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/Macklin06/optiroute/router/config"
	"github.com/Macklin06/optiroute/router/internal/handlers"
	"github.com/Macklin06/optiroute/router/internal/models"
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

	driverHandler := handlers.NewDriverHandler(redisClient, db)

	router := gin.Default()

	router.GET("/health", healthHandler)

	v1 := router.Group("/api/v1")
	{
		drivers := v1.Group("/drivers")
		{
			drivers.PUT("/location", driverHandler.UpdateLocation)
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
