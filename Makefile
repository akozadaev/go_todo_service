.PHONY: run dev test-curl clean help build lint test

SERVER_URL = http://localhost:8080/api/v1

# Сборка бинарника
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

# Запуск в обычном режиме
run:
	@echo "Running application..."
	go run cmd/api/main.go

# Запуск в режиме разработки с Air (hot reload)
dev:
	@echo "Running in development mode with Air..."
	@mkdir -p tmp
	@air -c air.toml

# Установка зависимостей
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Запуск линтера
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Форматирование кода
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Проверка кода
vet:
	@echo "Running go vet..."
	go vet ./...

# Запуск тестов
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Тестирование API через cURL
test-curl:
	@echo "1. Получить все задачи (ожидается пустой список)..."
	curl -s -X GET $(SERVER_URL)/todos
	@echo

	@echo "2. Создать задачу 'Купить продукты'..."
	curl -s -X POST $(SERVER_URL)/todos \
		-H "Content-Type: application/json" \
		-d '{"title": "Купить продукты", "description": "Молоко, хлеб, яйца", "done": false}'
	@echo

	@echo "3. Создать задачу 'Прогуляться с собакой'..."
	curl -s -X POST $(SERVER_URL)/todos \
		-H "Content-Type: application/json" \
		-d '{"title": "Прогуляться с собакой", "description": "30 минут в парке", "done": false}'
	@echo

	@echo "4. Получить все задачи (должно быть 2)..."
	curl -s -X GET $(SERVER_URL)/todos
	@echo

	@echo "5. Получить задачу с ID=1..."
	curl -s -X GET $(SERVER_URL)/todos/1
	@echo

	@echo "6. Обновить задачу ID=1 (отметить как выполненную)..."
	curl -s -X PUT $(SERVER_URL)/todos/1 \
		-H "Content-Type: application/json" \
		-d '{"title": "Купить продукты", "description": "Молоко, хлеб, яйца", "done": true}'
	@echo

	@echo "7. Получить обновлённую задачу ID=1..."
	curl -s -X GET $(SERVER_URL)/todos/1
	@echo

	@echo "8. Удалить задачу ID=2..."
	curl -s -X DELETE $(SERVER_URL)/todos/2
	@echo " (удалено)"
	@echo

	@echo "9. Попытаться получить удалённую задачу ID=2 (должна быть ошибка)..."
	curl -s -X GET $(SERVER_URL)/todos/2
	@echo

	@echo "10. Финальный список задач (должна остаться только ID=1)..."
	curl -s -X GET $(SERVER_URL)/todos
	@echo

# Очистка временных файлов
clean:
	@echo "Cleaning up..."
	rm -f tmp/main tmp/build-errors.log
	rm -f build-errors.log
	rm -rf bin/
	rm -f coverage.out coverage.html

# Создание .env из примера
init:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created. Please update DB_PASSWORD!"; \
	else \
		echo ".env file already exists"; \
	fi

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build          - Собрать бинарный файл приложения"
	@echo "  make run            - Запустить приложение"
	@echo "  make dev            - Запустить с Air (hot reload)"
	@echo "  make deps           - Скачать и обновить зависимости"
	@echo "  make lint           - Запустить линтер (golangci-lint)"
	@echo "  make fmt            - Форматировать код"
	@echo "  make vet            - Запустить go vet"
	@echo "  make test           - Запустить тесты с покрытием"
	@echo "  make test-curl      - Протестировать API через cURL"
	@echo "  make clean          - Очистить временные файлы"
	@echo "  make init           - Создать .env из .env.example"
	@echo "  make help           - Показать это сообщение"
