package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	todopb "github.com/akozadaev/go_todo_service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	defaultAddress = "localhost:50051"
)

func main() {
	address := flag.String("addr", defaultAddress, "gRPC server address")
	userID := flag.Uint("user", 1, "User identifier used for multi-user mode")
	flag.Parse()

	// Подключаемся к gRPC серверу
	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := todopb.NewTodoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.Pairs("user-id", fmt.Sprintf("%d", *userID))
	ctx = metadata.NewOutgoingContext(ctx, md)

	fmt.Println("=== gRPC TODO Client Demo ===")

	// 1. Создание задачи
	fmt.Println("1. Creating a new todo...")
	createReq := &todopb.CreateTodoRequest{
		Title:       "Купить продукты",
		Description: "Молоко, хлеб, яйца",
		Done:        false,
	}
	createResp, err := client.CreateTodo(ctx, createReq)
	if err != nil {
		log.Fatalf("Failed to create todo: %v", err)
	}
	fmt.Printf("Created: ID=%d, Title=%s\n\n", createResp.Todo.Id, createResp.Todo.Title)

	// 2. Получение задачи по ID
	fmt.Println("2. Getting todo by ID...")
	getReq := &todopb.GetTodoRequest{Id: createResp.Todo.Id}
	getResp, err := client.GetTodo(ctx, getReq)
	if err != nil {
		log.Fatalf("Failed to get todo: %v", err)
	}
	fmt.Printf("Got: ID=%d, Title=%s, Done=%v\n\n", getResp.Todo.Id, getResp.Todo.Title, getResp.Todo.Done)

	// 3. Получение всех задач
	fmt.Println("3. Listing all todos...")
	listResp, err := client.ListTodos(ctx, &todopb.Empty{})
	if err != nil {
		log.Fatalf("Failed to list todos: %v", err)
	}
	fmt.Printf("Found %d todos:\n", len(listResp.Todos))
	for _, todo := range listResp.Todos {
		fmt.Printf("  - ID=%d, Title=%s, Done=%v\n", todo.Id, todo.Title, todo.Done)
	}
	fmt.Println()

	// 4. Обновление задачи
	fmt.Println("4. Updating todo...")
	updateReq := &todopb.UpdateTodoRequest{
		Id:          createResp.Todo.Id,
		Title:       "Купить продукты",
		Description: "Молоко, хлеб, яйца",
		Done:        true,
	}
	updateResp, err := client.UpdateTodo(ctx, updateReq)
	if err != nil {
		log.Fatalf("Failed to update todo: %v", err)
	}
	fmt.Printf("Updated: ID=%d, Title=%s, Done=%v\n\n", updateResp.Todo.Id, updateResp.Todo.Title, updateResp.Todo.Done)

	// 5. Стриминг: получение всех задач через поток
	fmt.Println("5. Streaming todos...")
	stream, err := client.ListTodosStream(ctx, &todopb.Empty{})
	if err != nil {
		log.Fatalf("Failed to stream todos: %v", err)
	}
	fmt.Println("Streaming todos:")
	for {
		todo, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Printf("  Received: ID=%d, Title=%s\n", todo.Id, todo.Title)
	}
	fmt.Println()

	// 6. Массовое создание задач
	fmt.Println("6. Bulk creating todos...")
	bulkStream, err := client.BulkCreateTodos(ctx)
	if err != nil {
		log.Fatalf("Failed to start bulk create: %v", err)
	}

	todosToCreate := []*todopb.CreateTodoRequest{
		{Title: "Задача 1", Description: "Описание 1", Done: false},
		{Title: "Задача 2", Description: "Описание 2", Done: false},
		{Title: "Задача 3", Description: "Описание 3", Done: false},
	}

	for _, todo := range todosToCreate {
		if err := bulkStream.Send(todo); err != nil {
			log.Fatalf("Failed to send todo: %v", err)
		}
	}

	bulkResp, err := bulkStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive bulk response: %v", err)
	}
	fmt.Printf("Bulk created %d todos\n\n", bulkResp.CreatedCount)

	// 7. Удаление задачи
	fmt.Println("7. Deleting todo...")
	deleteReq := &todopb.DeleteTodoRequest{Id: createResp.Todo.Id}
	_, err = client.DeleteTodo(ctx, deleteReq)
	if err != nil {
		log.Fatalf("Failed to delete todo: %v", err)
	}
	fmt.Printf("Deleted todo ID=%d\n\n", createResp.Todo.Id)

	// 8. Финальный список задач
	fmt.Println("8. Final list of todos...")
	finalResp, err := client.ListTodos(ctx, &todopb.Empty{})
	if err != nil {
		log.Fatalf("Failed to list todos: %v", err)
	}
	fmt.Printf("Found %d todos:\n", len(finalResp.Todos))
	for _, todo := range finalResp.Todos {
		fmt.Printf("  - ID=%d, Title=%s, Done=%v\n", todo.Id, todo.Title, todo.Done)
	}

	fmt.Println("\n=== Demo completed successfully ===")
}
