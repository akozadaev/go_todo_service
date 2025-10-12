package repository

import (
	"context"
	"errors"

	"todo-service/internal/model"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("todo not found")
	ErrInvalidID     = errors.New("invalid todo ID")
	ErrDatabaseError = errors.New("database error")
)

// TodoRepository определяет интерфейс для работы с TODO
type TodoRepository interface {
	Create(ctx context.Context, todo *model.Todo) error
	GetByID(ctx context.Context, id uint) (*model.Todo, error)
	GetAll(ctx context.Context) ([]model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository создает новый экземпляр репозитория
func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{
		db: db,
	}
}

func (r *todoRepository) Create(ctx context.Context, todo *model.Todo) error {
	if err := r.db.WithContext(ctx).Create(todo).Error; err != nil {
		return ErrDatabaseError
	}
	return nil
}

func (r *todoRepository) GetByID(ctx context.Context, id uint) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.WithContext(ctx).First(&todo, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrDatabaseError
	}
	return &todo, nil
}

func (r *todoRepository) GetAll(ctx context.Context) ([]model.Todo, error) {
	var todos []model.Todo
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		return nil, ErrDatabaseError
	}
	return todos, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *model.Todo) error {
	result := r.db.WithContext(ctx).Model(todo).Updates(map[string]interface{}{
		"title":       todo.Title,
		"description": todo.Description,
		"done":        todo.Done,
	})
	
	if result.Error != nil {
		return ErrDatabaseError
	}
	
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Todo{}, id)
	
	if result.Error != nil {
		return ErrDatabaseError
	}
	
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	
	return nil
}

