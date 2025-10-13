package service

import (
	"context"
	"fmt"

	"github.com/akozadaev/go_todo_service/internal/model"
	"github.com/akozadaev/go_todo_service/internal/repository"
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
	repo repository.TodoRepository
}

// NewTodoService создает новый экземпляр сервиса
func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{
		repo: repo,
	}
}

func (s *todoService) Create(ctx context.Context, req *model.TodoCreateRequest) (*model.Todo, error) {
	todo := req.ToTodo()

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

func (s *todoService) GetByID(ctx context.Context, id uint) (*model.Todo, error) {
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return todo, nil
}

func (s *todoService) GetAll(ctx context.Context) ([]model.Todo, error) {
	todos, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *todoService) Update(ctx context.Context, id uint, req *model.TodoUpdateRequest) (*model.Todo, error) {
	// Проверяем существование
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Done = req.Done

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return existing, nil
}

func (s *todoService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
