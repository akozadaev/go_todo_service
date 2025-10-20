package service

import (
	"context"
	"fmt"

	"github.com/akozadaev/go_todo_service/internal/model"
	"github.com/akozadaev/go_todo_service/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TodoService определяет интерфейс бизнес-логики для TODO
type TodoService interface {
	Create(ctx context.Context, req *model.TodoCreateRequest) (*model.Todo, error)
	GetByID(ctx context.Context, id uint) (*model.Todo, error)
	GetAll(ctx context.Context) ([]model.Todo, error)
	Update(ctx context.Context, id uint, req *model.TodoUpdateRequest) (*model.Todo, error)
	Delete(ctx context.Context, id uint) error
}

type todoService struct {
	repo   repository.TodoRepository
	tracer trace.Tracer
}

// NewTodoService создает новый экземпляр сервиса
func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{
		repo:   repo,
		tracer: otel.Tracer("go-todo-service"),
	}
}

func (s *todoService) Create(ctx context.Context, req *model.TodoCreateRequest) (*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.CreateTodo")
	defer span.End()

	span.SetAttributes(
		attribute.String("todo.title", req.Title),
		attribute.String("todo.description", req.Description),
	)

	todo := req.ToTodo()

	if err := s.repo.Create(ctx, todo); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create todo")
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	span.SetAttributes(attribute.Int64("todo.id", int64(todo.ID)))
	span.SetStatus(codes.Ok, "Todo created successfully")
	return todo, nil
}

func (s *todoService) GetByID(ctx context.Context, id uint) (*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetTodoByID")
	defer span.End()

	span.SetAttributes(attribute.Int64("todo.id", int64(id)))

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get todo")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("todo.title", todo.Title),
		attribute.Bool("todo.done", todo.Done),
	)
	span.SetStatus(codes.Ok, "Todo retrieved successfully")
	return todo, nil
}

func (s *todoService) GetAll(ctx context.Context) ([]model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetAllTodos")
	defer span.End()

	todos, err := s.repo.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get all todos")
		return nil, err
	}

	span.SetAttributes(attribute.Int("todos.count", len(todos)))
	span.SetStatus(codes.Ok, "All todos retrieved successfully")
	return todos, nil
}

func (s *todoService) Update(ctx context.Context, id uint, req *model.TodoUpdateRequest) (*model.Todo, error) {
	ctx, span := s.tracer.Start(ctx, "service.UpdateTodo")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("todo.id", int64(id)),
		attribute.String("todo.title", req.Title),
		attribute.String("todo.description", req.Description),
		attribute.Bool("todo.done", req.Done),
	)

	// Проверяем существование
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get existing todo")
		return nil, err
	}

	// Обновляем поля
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Done = req.Done

	if err := s.repo.Update(ctx, existing); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update todo")
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	span.SetStatus(codes.Ok, "Todo updated successfully")
	return existing, nil
}

func (s *todoService) Delete(ctx context.Context, id uint) error {
	ctx, span := s.tracer.Start(ctx, "service.DeleteTodo")
	defer span.End()

	span.SetAttributes(attribute.Int64("todo.id", int64(id)))

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete todo")
		return err
	}

	span.SetStatus(codes.Ok, "Todo deleted successfully")
	return nil
}
