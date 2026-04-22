package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"order-service/internal/db"
	"order-service/internal/models"

	"github.com/gin-gonic/gin"
)

func sendError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": gin.H{"code": code, "message": message, "timestamp": time.Now().Unix()}})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "order-service"})
}

func CreateOrder(c *gin.Context) {
	var input models.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input. Amount must be > 0 and PaymentMethod is required.")
		return
	}

	order := models.Order{
		Amount:        input.Amount,
		PaymentMethod: input.PaymentMethod,
		Description:   input.Description,
		Status:        models.StatusNew,
	}

	db.DB.Create(&order)
	db.InvalidateCache()
	c.JSON(http.StatusCreated, order)
}

func GetOrders(c *gin.Context) {
	var orders []models.Order
	statusFilter := c.Query("status")

	query := db.DB
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	query.Find(&orders)
	c.JSON(http.StatusOK, orders)
}

func GetOrder(c *gin.Context) {
	var order models.Order
	if err := db.DB.First(&order, c.Param("id")).Error; err != nil {
		sendError(c, http.StatusNotFound, "Order not found")
		return
	}
	c.JSON(http.StatusOK, order)
}

func UpdateDescription(c *gin.Context) {
	var order models.Order
	if err := db.DB.First(&order, c.Param("id")).Error; err != nil {
		sendError(c, http.StatusNotFound, "Order not found")
		return
	}

	var input models.UpdateDescInput
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Description is required")
		return
	}

	db.DB.Model(&order).Update("description", input.Description)
	c.JSON(http.StatusOK, order)
}

func PayOrder(c *gin.Context) {
	var order models.Order
	if err := db.DB.First(&order, c.Param("id")).Error; err != nil {
		sendError(c, http.StatusNotFound, "Order not found")
		return
	}

	if order.Status != models.StatusNew {
		sendError(c, http.StatusConflict, "Only NEW orders can be paid")
		return
	}

	db.DB.Model(&order).Update("status", models.StatusPaid)
	db.InvalidateCache()
	c.JSON(http.StatusOK, gin.H{"message": "Order paid successfully", "order": order})
}

func CancelOrder(c *gin.Context) {
	var order models.Order
	if err := db.DB.First(&order, c.Param("id")).Error; err != nil {
		sendError(c, http.StatusNotFound, "Order not found")
		return
	}

	if order.Status != models.StatusNew {
		sendError(c, http.StatusConflict, "Only NEW orders can be cancelled")
		return
	}

	db.DB.Model(&order).Update("status", models.StatusCancelled)
	db.InvalidateCache()
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully", "order": order})
}

func GetStats(c *gin.Context) {
	cachedStats, err := db.RDB.Get(db.Ctx, "orders_stats").Result()
	if err == nil {
		var stats []map[string]interface{}
		json.Unmarshal([]byte(cachedStats), &stats)
		c.JSON(http.StatusOK, gin.H{"source": "redis", "data": stats})
		return
	}

	var stats []map[string]interface{}
	db.DB.Model(&models.Order{}).Select("status, count(id) as count").Group("status").Scan(&stats)

	statsJSON, _ := json.Marshal(stats)
	db.RDB.Set(db.Ctx, "orders_stats", statsJSON, time.Hour)

	c.JSON(http.StatusOK, gin.H{"source": "database", "data": stats})
}