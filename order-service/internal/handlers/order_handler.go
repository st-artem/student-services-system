package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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

// ІЗ №4: Дедуплікація вхідних даних
func CreateOrder(c *gin.Context) {
	var input models.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input. Amount must be > 0 and PaymentMethod is required.")
		return
	}

	// Алгоритм дедуплікації: шукаємо таке ж замовлення за останні 10 секунд
	var count int64
	tenSecondsAgo := time.Now().Add(-10 * time.Second)
	db.DB.Model(&models.Order{}).
		Where("amount = ? AND payment_method = ? AND created_at >= ?", input.Amount, input.PaymentMethod, tenSecondsAgo).
		Count(&count)

	if count > 0 {
		sendError(c, http.StatusConflict, "Duplicate order detected within 10 seconds window")
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

// ІЗ №1 (Сортування), ІЗ №2 (Пагінація), ІЗ №6 (Часові інтервали), ІЗ №8 (Конфліктні параметри)
func GetOrders(c *gin.Context) {
	var orders []models.Order
	query := db.DB

	// 1. Фільтрація за статусом
	if statusFilter := c.Query("status"); statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	// 2. Фільтрація за часовим інтервалом (ІЗ №6 та №8)
	createdAfter := c.Query("created_after")
	createdBefore := c.Query("created_before")

	if createdAfter != "" && createdBefore != "" {
		// Парсимо дати (очікуємо формат RFC3339, напр. "2026-04-20T10:00:00Z")
		tAfter, err1 := time.Parse(time.RFC3339, createdAfter)
		tBefore, err2 := time.Parse(time.RFC3339, createdBefore)

		if err1 == nil && err2 == nil {
			// ІЗ №8: Негативний сценарій (конфлікт параметрів)
			if tAfter.After(tBefore) {
				sendError(c, http.StatusBadRequest, "Invalid interval: created_after cannot be greater than created_before")
				return
			}
			query = query.Where("created_at >= ? AND created_at <= ?", tAfter, tBefore)
		} else {
			sendError(c, http.StatusBadRequest, "Invalid date format. Use RFC3339 (e.g., 2026-04-20T10:00:00Z)")
			return
		}
	} else if createdAfter != "" {
		query = query.Where("created_at >= ?", createdAfter)
	} else if createdBefore != "" {
		query = query.Where("created_at <= ?", createdBefore)
	}

	// 3. Сортування за двома полями (ІЗ №1)
	// Очікуємо формат: ?sort=status:asc,amount:desc
	if sortParam := c.Query("sort"); sortParam != "" {
		sorts := strings.Split(sortParam, ",")
		for _, s := range sorts {
			parts := strings.Split(s, ":")
			if len(parts) == 2 {
				field, order := parts[0], parts[1]
				if (field == "status" || field == "amount" || field == "created_at") && (order == "asc" || order == "desc") {
					query = query.Order(field + " " + order)
				}
			}
		}
	} else {
		query = query.Order("created_at desc") // Дефолтне сортування
	}

	// 4. Пагінація (ІЗ №2)
	limit := 10
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	query.Limit(limit).Offset(offset).Find(&orders)
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

// ІЗ №5: Перевірка інваріанта переходів
func PayOrder(c *gin.Context) {
	var order models.Order
	if err := db.DB.First(&order, c.Param("id")).Error; err != nil {
		sendError(c, http.StatusNotFound, "Order not found")
		return
	}
	if order.Status == models.StatusCancelled {
		sendError(c, http.StatusConflict, "Invariant violation: CANCELLED order cannot be PAID")
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
	if order.Status == models.StatusPaid {
		sendError(c, http.StatusConflict, "Invariant violation: PAID order cannot be CANCELLED")
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

// ІЗ №3: Статистика
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