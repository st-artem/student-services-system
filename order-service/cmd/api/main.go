package main

import (
	"order-service/internal/db"
	"order-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ініціалізація баз даних
	db.Init()

	r := gin.Default()

	// Реєстрація маршрутів
	r.GET("/health", handlers.HealthCheck)
	
	orders := r.Group("/orders")
	{
		orders.POST("", handlers.CreateOrder)
		orders.GET("", handlers.GetOrders)
		orders.GET("/stats/summary", handlers.GetStats)
		orders.GET("/:id", handlers.GetOrder)
		orders.PATCH("/:id/description", handlers.UpdateDescription)
		
		orders.POST("/:id/pay", handlers.PayOrder)
		orders.POST("/:id/cancel", handlers.CancelOrder)
	}

	r.Run(":8080")
}