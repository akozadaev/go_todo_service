# Установка grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Список сервисов
grpcurl -plaintext localhost:50051 list

# Создание задачи
grpcurl -plaintext -d '{
  "title": "Купить продукты",
  "description": "Молоко, хлеб, яйца",
  "done": false
}' localhost:50051 todo.TodoService/CreateTodo

# Получение всех задач
grpcurl -plaintext localhost:50051 todo.TodoService/ListTodos
