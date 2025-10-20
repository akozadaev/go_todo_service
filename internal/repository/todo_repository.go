package repository

import (
	"context"
	"errors"

	"github.com/akozadaev/go_todo_service/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	db     *gorm.DB
	tracer trace.Tracer
}

// NewTodoRepository создает новый экземпляр репозитория
func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{
		db:     db,
		tracer: otel.Tracer("go-todo-service"),
	}
}

func (r *todoRepository) Create(ctx context.Context, todo *model.Todo) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateTodo")
	defer span.End()

	span.SetAttributes(
		attribute.String("todo.title", todo.Title),
		attribute.String("todo.description", todo.Description),
		attribute.Bool("todo.done", todo.Done),
	)

	if err := r.db.WithContext(ctx).Create(todo).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database create error")
		return ErrDatabaseError
	}

	span.SetAttributes(attribute.Int64("todo.id", int64(todo.ID)))
	span.SetStatus(codes.Ok, "Todo created in database")
	return nil
}

func (r *todoRepository) GetByID(ctx context.Context, id uint) (*model.Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetTodoByID")
	defer span.End()

	span.SetAttributes(attribute.Int64("todo.id", int64(id)))

	var todo model.Todo
	err := r.db.WithContext(ctx).First(&todo, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "Todo not found")
			return nil, ErrNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database query error")
		return nil, ErrDatabaseError
	}

	span.SetAttributes(
		attribute.String("todo.title", todo.Title),
		attribute.Bool("todo.done", todo.Done),
	)
	span.SetStatus(codes.Ok, "Todo retrieved from database")
	return &todo, nil
}

func (r *todoRepository) GetAll(ctx context.Context) ([]model.Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetAllTodos")
	defer span.End()

	var todos []model.Todo
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database query error")
		return nil, ErrDatabaseError
	}

	span.SetAttributes(attribute.Int("todos.count", len(todos)))
	span.SetStatus(codes.Ok, "All todos retrieved from database")
	return todos, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *model.Todo) error {
	ctx, span := r.tracer.Start(ctx, "repository.UpdateTodo")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("todo.id", int64(todo.ID)),
		attribute.String("todo.title", todo.Title),
		attribute.String("todo.description", todo.Description),
		attribute.Bool("todo.done", todo.Done),
	)

	result := r.db.WithContext(ctx).Model(todo).Updates(map[string]interface{}{
		"title":       todo.Title,
		"description": todo.Description,
		"done":        todo.Done,
	})

	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "Database update error")
		return ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, "Todo not found for update")
		return ErrNotFound
	}

	span.SetStatus(codes.Ok, "Todo updated in database")
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "repository.DeleteTodo")
	defer span.End()

	span.SetAttributes(attribute.Int64("todo.id", int64(id)))

	result := r.db.WithContext(ctx).Delete(&model.Todo{}, id)

	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "Database delete error")
		return ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, "Todo not found for delete")
		return ErrNotFound
	}

	span.SetStatus(codes.Ok, "Todo deleted from database")
	return nil
}
