# gRPC Client Example

Пример клиента для демонстрации работы с gRPC API TODO сервиса.

## Запуск

### 1. Запустите сервер

```bash
make dev
```

Сервер будет доступен на:
- HTTP REST API: `http://localhost:8080`
- gRPC API: `localhost:50051`

### 2. Запустите клиент

```bash
cd examples/grpc_client
go run main.go
```

Или с другим адресом сервера:

```bash
go run main.go -addr localhost:50051
```

## Возможности

Клиент демонстрирует следующие операции:

1. **CreateTodo** - создание новой задачи
2. **GetTodo** - получение задачи по ID
3. **ListTodos** - получение всех задач
4. **UpdateTodo** - обновление задачи
5. **ListTodosStream** - потоковое получение всех задач
6. **BulkCreateTodos** - массовое создание задач через поток
7. **DeleteTodo** - удаление задачи

## Использование grpcurl

Альтернативный способ тестирования gRPC сервиса:

```bash
# Установка grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Список сервисов
grpcurl -plaintext localhost:50051 list

# Список методов TodoService
grpcurl -plaintext localhost:50051 list todo.TodoService

# Создание задачи
grpcurl -plaintext -d '{
  "title": "Тестовая задача",
  "description": "Описание",
  "done": false
}' localhost:50051 todo.TodoService/CreateTodo

# Получение всех задач
grpcurl -plaintext localhost:50051 todo.TodoService/ListTodos

# Получение задачи по ID
grpcurl -plaintext -d '{"id": 1}' localhost:50051 todo.TodoService/GetTodo
```

## Протокол

gRPC сервис использует Protocol Buffers и определен в файле [todo.proto](../../api/proto/todo.proto).

Для генерации Go кода из proto файлов:

```bash
make proto
```

