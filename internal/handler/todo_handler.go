package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"todo-service/internal/model"
	"todo-service/internal/repository"
	"todo-service/internal/service"

	"github.com/gin-gonic/gin"
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
// @Success 200 {array} model.Todo
// @Failure 500 {object} ErrorResponse
// @Router /todos [get]
func (h *TodoHandler) GetAll(c *gin.Context) {
	todos, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		// Ошибка получения списка из БД - проблема сервера (500)
		log.Printf("ERROR: Failed to fetch todos: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch todos"))
		return
	}

	c.JSON(http.StatusOK, todos)
}

// GetByID godoc
// @Summary Получить задачу по ID
// @Description Возвращает задачу по указанному ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [get]
func (h *TodoHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	todo, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - это нормальное поведение, не логируем как ошибку
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - реальная проблема с сервером (БД недоступна и т.д.)
		log.Printf("ERROR: Failed to fetch todo %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch todo"))
		return
	}

	c.JSON(http.StatusOK, todo)
}

// Create godoc
// @Summary Создать новую задачу
// @Description Создает новую задачу
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body model.TodoCreateRequest true "Todo data"
// @Success 201 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	var req model.TodoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 400 - проблема с запросом клиента
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	todo, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		// 500 - не смогли создать в БД (проблема сервера)
		log.Printf("ERROR: Failed to create todo: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create todo"))
		return
	}

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
// @Success 200 {object} model.Todo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [put]
func (h *TodoHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	var req model.TodoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	todo, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - задача не найдена (нормальное поведение)
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - проблема с обновлением в БД
		log.Printf("ERROR: Failed to update todo %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update todo"))
		return
	}

	c.JSON(http.StatusOK, todo)
}

// Delete godoc
// @Summary Удалить задачу
// @Description Удаляет задачу по ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /todos/{id} [delete]
func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 404 - задача не найдена (нормальное поведение)
			c.JSON(http.StatusNotFound, NewErrorResponse("Todo not found"))
			return
		}
		// 500 - проблема с удалением из БД
		log.Printf("ERROR: Failed to delete todo %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete todo"))
		return
	}

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
