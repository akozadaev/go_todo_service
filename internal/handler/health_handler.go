package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler создает экземпляр обработчика работоспособности
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// RegisterRoutes регистрирует маршруты для проверки работоспособности
func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)
}

// HealthResponse представляет ответ о состоянии работоспособности
type HealthResponse struct {
	Status string `json:"status"`
}

// Health проверяет, что сервис запущен
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

// Ready проверяет готовность сервиса (включая БД)
func (h *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status: "not ready",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status: "not ready",
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status: "ready",
	})
}
