package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/akozadaev/go_todo_service/internal/logger"
	"github.com/akozadaev/go_todo_service/internal/model"
	"github.com/akozadaev/go_todo_service/internal/repository"
	"github.com/akozadaev/go_todo_service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TodoHandler struct {
	service service.TodoService
}

// NewTodoHandler создает новый экземпляр обработчика
func NewTodoHandler(service service.TodoService) *TodoHandler {
	return &TodoHandler{
		service: service,
	}
}

// RegisterRoutes регистрирует все маршруты для TODO
func (h *TodoHandler) RegisterRoutes(r *gin.RouterGroup) {
	todos := r.Group("/todos")
	{
		todos.GET("", h.GetAll)
		todos.GET("/:id", h.GetByID)
		todos.POST("", h.Create)
		todos.PUT("/:id", h.Update)
		todos.DELETE("/:id", h.Delete)
	}
}

// GetAll godoc
// @Summary Получить все задачи
// @Description Возвращает список всех задач
// @Tags todos
// @Accept json
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Success 200 {array} model.Todo
// @Failure 500 {object} ErrorResponse
// @Router /todos [get]
func (h *TodoHandler) GetAll(c *gin.Context) {
	reqLogger := logger.NewRequestLogger(c.Request.Context())

	reqLogger.Info("Getting all todos")

	todos, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		// Ошибка получения списка из БД - проблема сервера (500)
		reqLogger.Error("Failed to fetch todos", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch todos"))
		return
	}

	reqLogger.Info("Successfully fetched todos", zap.Int("count", len(todos)))
	c.JSON(http.StatusOK, todos)
}

// GetByID godoc
// @Summary Получить задачу по ID
// @Description Возвращает задачу по указанному ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param X-User-ID header int true "User ID"
// @Success 200 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [get]
func (h *TodoHandler) GetByID(c *gin.Context) {
	reqLogger := logger.NewRequestLogger(c.Request.Context())

	id, err := parseID(c)
	if err != nil {
		reqLogger.Warn("Invalid ID provided", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	reqLogger.Info("Getting todo by ID", zap.Uint("todo_id", id))

	todo, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - это нормальное поведение, не логируем как ошибку
			reqLogger.Info("Todo not found", zap.Uint("todo_id", id))
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - реальная проблема с сервером (БД недоступна и т.д.)
		reqLogger.Error("Failed to fetch todo", zap.Uint("todo_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch todo"))
		return
	}

	reqLogger.Info("Successfully fetched todo", zap.Uint("todo_id", id))
	c.JSON(http.StatusOK, todo)
}

// Create godoc
// @Summary Создать новую задачу
// @Description Создает новую задачу
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body model.TodoCreateRequest true "Todo data"
// @Param X-User-ID header int true "User ID"
// @Success 201 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	reqLogger := logger.NewRequestLogger(c.Request.Context())

	var req model.TodoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 400 - проблема с запросом клиента
		reqLogger.Warn("Invalid JSON in create request", zap.Error(err))
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	reqLogger.Info("Creating new todo", zap.String("title", req.Title))

	todo, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		// 500 - не смогли создать в БД (проблема сервера)
		reqLogger.Error("Failed to create todo", zap.String("title", req.Title), zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create todo"))
		return
	}

	reqLogger.Info("Successfully created todo", zap.Uint("todo_id", todo.ID), zap.String("title", todo.Title))
	c.JSON(http.StatusCreated, todo)
}

// Update godoc
// @Summary Обновить задачу
// @Description Обновляет существующую задачу
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body model.TodoUpdateRequest true "Updated todo data"
// @Param X-User-ID header int true "User ID"
// @Success 200 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [put]
func (h *TodoHandler) Update(c *gin.Context) {
	reqLogger := logger.NewRequestLogger(c.Request.Context())

	id, err := parseID(c)
	if err != nil {
		reqLogger.Warn("Invalid ID provided for update", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	var req model.TodoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("Invalid JSON in update request", zap.Uint("todo_id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	reqLogger.Info("Updating todo", zap.Uint("todo_id", id))

	todo, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - задача не найдена (нормальное поведение)
			reqLogger.Info("Todo not found for update", zap.Uint("todo_id", id))
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - проблема с обновлением в БД
		reqLogger.Error("Failed to update todo", zap.Uint("todo_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update todo"))
		return
	}

	reqLogger.Info("Successfully updated todo", zap.Uint("todo_id", id))
	c.JSON(http.StatusOK, todo)
}

// Delete godoc
// @Summary Удалить задачу
// @Description Удаляет задачу по ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param X-User-ID header int true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [delete]
func (h *TodoHandler) Delete(c *gin.Context) {
	reqLogger := logger.NewRequestLogger(c.Request.Context())

	id, err := parseID(c)
	if err != nil {
		reqLogger.Warn("Invalid ID provided for delete", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	reqLogger.Info("Deleting todo", zap.Uint("todo_id", id))

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - задача не найдена (нормальное поведение)
			reqLogger.Info("Todo not found for delete", zap.Uint("todo_id", id))
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - проблема с удалением из БД
		reqLogger.Error("Failed to delete todo", zap.Uint("todo_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete todo"))
		return
	}

	reqLogger.Info("Successfully deleted todo", zap.Uint("todo_id", id))
	c.Status(http.StatusNoContent)
}

// parseID парсит ID из параметра пути
func parseID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
