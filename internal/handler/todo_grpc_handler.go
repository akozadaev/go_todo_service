package handler

import (
	"context"
	"errors"
	"io"
	"strconv"

	todopb "github.com/akozadaev/go_todo_service/api/proto"
	"github.com/akozadaev/go_todo_service/internal/model"
	"github.com/akozadaev/go_todo_service/internal/repository"
	"github.com/akozadaev/go_todo_service/internal/service"
	"github.com/akozadaev/go_todo_service/internal/userctx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TodoGRPCHandler реализует gRPC интерфейс для TODO сервиса
type TodoGRPCHandler struct {
	todopb.UnimplementedTodoServiceServer
	service service.TodoService
	logger  *zap.Logger
}

const grpcUserIDMetadataKey = "user-id"

// NewTodoGRPCHandler создает новый экземпляр gRPC handler
func NewTodoGRPCHandler(service service.TodoService) *TodoGRPCHandler {
	return &TodoGRPCHandler{
		service: service,
		logger:  zap.L(),
	}
}

// CreateTodo создает новую задачу
func (h *TodoGRPCHandler) CreateTodo(ctx context.Context, req *todopb.CreateTodoRequest) (*todopb.TodoResponse, error) {
	ctx, userID, err := h.ensureUserContext(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Info("gRPC: Creating todo", zap.Uint64("user_id", userID), zap.String("title", req.Title))

	createReq := &model.TodoCreateRequest{
		Title:       req.Title,
		Description: req.Description,
		Done:        req.Done,
	}

	todo, err := h.service.Create(ctx, createReq)
	if err != nil {
		h.logger.Error("gRPC: Failed to create todo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create todo")
	}

	h.logger.Info("gRPC: Successfully created todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", uint64(todo.ID)))
	return &todopb.TodoResponse{Todo: h.toProtoTodo(todo)}, nil
}

// GetTodo получает задачу по ID
func (h *TodoGRPCHandler) GetTodo(ctx context.Context, req *todopb.GetTodoRequest) (*todopb.TodoResponse, error) {
	ctx, userID, err := h.ensureUserContext(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Info("gRPC: Getting todo by ID", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))

	todo, err := h.service.GetByID(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.Info("gRPC: Todo not found", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		h.logger.Error("gRPC: Failed to get todo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get todo")
	}

	h.logger.Info("gRPC: Successfully fetched todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
	return &todopb.TodoResponse{Todo: h.toProtoTodo(todo)}, nil
}

// ListTodos получает все задачи
func (h *TodoGRPCHandler) ListTodos(ctx context.Context, _ *todopb.Empty) (*todopb.ListTodosResponse, error) {
	ctx, userID, err := h.ensureUserContext(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Info("gRPC: Getting all todos", zap.Uint64("user_id", userID))

	todos, err := h.service.GetAll(ctx)
	if err != nil {
		h.logger.Error("gRPC: Failed to get todos", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get todos")
	}

	protoTodos := make([]*todopb.Todo, len(todos))
	for i, t := range todos {
		protoTodos[i] = h.toProtoTodo(&t)
	}

	h.logger.Info("gRPC: Successfully fetched todos", zap.Uint64("user_id", userID), zap.Int("count", len(todos)))
	return &todopb.ListTodosResponse{Todos: protoTodos}, nil
}

// UpdateTodo обновляет задачу
func (h *TodoGRPCHandler) UpdateTodo(ctx context.Context, req *todopb.UpdateTodoRequest) (*todopb.TodoResponse, error) {
	ctx, userID, err := h.ensureUserContext(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Info("gRPC: Updating todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))

	updateReq := &model.TodoUpdateRequest{
		Title:       req.Title,
		Description: req.Description,
		Done:        req.Done,
	}

	todo, err := h.service.Update(ctx, uint(req.Id), updateReq)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.Info("gRPC: Todo not found for update", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		h.logger.Error("gRPC: Failed to update todo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update todo")
	}

	h.logger.Info("gRPC: Successfully updated todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
	return &todopb.TodoResponse{Todo: h.toProtoTodo(todo)}, nil
}

// DeleteTodo удаляет задачу
func (h *TodoGRPCHandler) DeleteTodo(ctx context.Context, req *todopb.DeleteTodoRequest) (*todopb.Empty, error) {
	ctx, userID, err := h.ensureUserContext(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Info("gRPC: Deleting todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))

	err = h.service.Delete(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.Info("gRPC: Todo not found for delete", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		h.logger.Error("gRPC: Failed to delete todo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete todo")
	}

	h.logger.Info("gRPC: Successfully deleted todo", zap.Uint64("user_id", userID), zap.Uint64("todo_id", req.Id))
	return &todopb.Empty{}, nil
}

// ListTodosStream - серверный поток для получения всех задач
func (h *TodoGRPCHandler) ListTodosStream(_ *todopb.Empty, stream todopb.TodoService_ListTodosStreamServer) error {
	ctx, userID, err := h.ensureUserContext(stream.Context())
	if err != nil {
		return err
	}

	h.logger.Info("gRPC: Starting todo stream", zap.Uint64("user_id", userID))

	todos, err := h.service.GetAll(ctx)
	if err != nil {
		h.logger.Error("gRPC: Failed to get todos for stream", zap.Error(err))
		return status.Error(codes.Internal, "failed to get todos")
	}

	for _, todo := range todos {
		if err := stream.Send(h.toProtoTodo(&todo)); err != nil {
			h.logger.Error("gRPC: Failed to send todo in stream", zap.Error(err))
			return err
		}
	}

	h.logger.Info("gRPC: Completed todo stream", zap.Uint64("user_id", userID), zap.Int("count", len(todos)))
	return nil
}

// BulkCreateTodos - клиентский поток для массового создания задач
func (h *TodoGRPCHandler) BulkCreateTodos(stream todopb.TodoService_BulkCreateTodosServer) error {
	ctx, userID, err := h.ensureUserContext(stream.Context())
	if err != nil {
		return err
	}

	h.logger.Info("gRPC: Starting bulk create todos", zap.Uint64("user_id", userID))

	var createdTodos []*todopb.Todo
	createdCount := 0

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			h.logger.Error("gRPC: Ошибка чтения клиентского потока", zap.Error(err))
			return status.Error(codes.Unavailable, "failed to receive todo")
		}

		createReq := &model.TodoCreateRequest{
			Title:       req.Title,
			Description: req.Description,
			Done:        req.Done,
		}

		todo, err := h.service.Create(ctx, createReq)
		if err != nil {
			h.logger.Error("gRPC: Failed to create todo in bulk", zap.Error(err))
			return status.Error(codes.Internal, "failed to create todo")
		}

		createdTodos = append(createdTodos, h.toProtoTodo(todo))
		createdCount++
	}

	h.logger.Info("gRPC: Completed bulk create", zap.Uint64("user_id", userID), zap.Int("count", createdCount))
	return stream.SendAndClose(&todopb.BulkCreateResponse{
		Todos:        createdTodos,
		CreatedCount: int32(createdCount),
	})
}

// toProtoTodo конвертирует model.Todo в todo.Todo
func (h *TodoGRPCHandler) toProtoTodo(todo *model.Todo) *todopb.Todo {
	return &todopb.Todo{
		Id:          uint64(todo.ID),
		UserId:      uint64(todo.UserID),
		Title:       todo.Title,
		Description: todo.Description,
		Done:        todo.Done,
		CreatedAt:   timestamppb.New(todo.CreatedAt),
		UpdatedAt:   timestamppb.New(todo.UpdatedAt),
	}
}

func (h *TodoGRPCHandler) ensureUserContext(ctx context.Context) (context.Context, uint64, error) {
	if userID, err := userctx.GetUserID(ctx); err == nil && userID != 0 {
		return ctx, uint64(userID), nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		h.logger.Warn("gRPC: Missing metadata with user id")
		return ctx, 0, status.Error(codes.Unauthenticated, "user-id metadata header required")
	}

	values := md.Get(grpcUserIDMetadataKey)
	if len(values) == 0 {
		h.logger.Warn("gRPC: Missing user-id metadata value")
		return ctx, 0, status.Error(codes.Unauthenticated, "user-id metadata header required")
	}

	parsedID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil || parsedID == 0 {
		h.logger.Warn("gRPC: Invalid user-id metadata value", zap.String("value", values[0]), zap.Error(err))
		return ctx, 0, status.Error(codes.InvalidArgument, "invalid user-id metadata value")
	}

	ctxWithUser := userctx.WithUserID(ctx, uint(parsedID))
	return ctxWithUser, parsedID, nil
}
