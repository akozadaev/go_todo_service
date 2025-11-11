.PHONY: run dev test-curl clean help build lint test profile profile-web profile-load swag swag-serve swag-local proto proto-gen proto-clean jaeger-up jaeger-stop jaeger-rm jaeger-down

SERVER_URL ?= http://localhost:8080/api/v1
USER_ID ?= 1
REST_USER_HEADER := -H "X-User-ID: $(USER_ID)"
CURL_JSON_HEADER := -H "Content-Type: application/json"

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

# Запуск в режиме разработки с stdout логгированием
dev-stdout:
	@echo "Running in development mode with stdout logging..."
	@LOG_ENABLE_STDOUT=true LOG_FORMAT=text go run cmd/api/main.go

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
	@echo "Testing API with user ID: $(USER_ID)"
	@echo "1. Получить все задачи (ожидается пустой список)..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos
	@echo

	@echo "2. Создать задачу 'Купить продукты'..."
	curl -s -X POST $(SERVER_URL)/todos \
		$(CURL_JSON_HEADER) \
		$(REST_USER_HEADER) \
		-d '{"title": "Купить продукты", "description": "Молоко, хлеб, яйца", "done": false}'
	@echo

	@echo "3. Создать задачу 'Прогуляться с собакой'..."
	curl -s -X POST $(SERVER_URL)/todos \
		$(CURL_JSON_HEADER) \
		$(REST_USER_HEADER) \
		-d '{"title": "Прогуляться с собакой", "description": "30 минут в парке", "done": false}'
	@echo

	@echo "4. Получить все задачи (должно быть 2)..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos
	@echo

	@echo "5. Получить задачу с ID=1..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos/1
	@echo

	@echo "6. Обновить задачу ID=1 (отметить как выполненную)..."
	curl -s -X PUT $(SERVER_URL)/todos/1 \
		$(CURL_JSON_HEADER) \
		$(REST_USER_HEADER) \
		-d '{"title": "Купить продукты", "description": "Молоко, хлеб, яйца", "done": true}'
	@echo

	@echo "7. Получить обновлённую задачу ID=1..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos/1
	@echo

	@echo "8. Удалить задачу ID=2..."
	curl -s -X DELETE $(REST_USER_HEADER) $(SERVER_URL)/todos/2
	@echo " (удалено)"
	@echo

	@echo "9. Попытаться получить удалённую задачу ID=2 (должна быть ошибка)..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos/2
	@echo

	@echo "10. Финальный список задач (должна остаться только ID=1)..."
	curl -s -X GET $(REST_USER_HEADER) $(SERVER_URL)/todos
	@echo

# Профилирование
profile:
	@echo "Running profiling test..."
	@./scripts/test_profiling.sh

profile-web:
	@echo "Opening pprof web interface..."
	@echo "Available endpoints:"
	@echo "  http://localhost:8080/debug/pprof/"
	@echo "  go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap"
	@echo "  go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile"

profile-load:
	@echo "Running full profiling test with load..."
	@./scripts/run_profiling_test.sh

# Генерация proto файлов
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p api/proto/gen/todo
	@which protoc > /dev/null || (echo "Please install protoc: https://grpc.io/docs/protoc-installation/" && exit 1)
	@protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		api/proto/todo.proto
	@echo "Proto files generated successfully in api/proto/gen/todo/"

# Очистка сгенерированных proto файлов
proto-clean:
	@echo "Cleaning generated proto files..."
	rm -rf api/proto/gen
	@echo "Proto files cleaned"

# Очистка временных файлов (обновлено)
clean:
	@echo "Cleaning up..."
	rm -f tmp/main tmp/build-errors.log
	rm -f build-errors.log
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf profiles/
	rm -rf logs/*.log
	rm -rf api/proto/gen
	@echo "Note: api/ folder preserved (contains swagger files)"

# Создание .env из примера
init:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created. Please update DB_PASSWORD!"; \
	else \
		echo ".env file already exists"; \
	fi
# Генерация OpenAPI документации
swag:
	@echo "Generating OpenAPI documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@v1.8.12)
	@swag init -g ./cmd/api/main.go -o ./api --parseDependency --parseInternal
	@echo "OpenAPI documentation generated in ./api/"
	@echo "Files created:"
	@ls -la ./api/

# Просмотр OpenAPI документации
swag-serve:
	@echo "OpenAPI Documentation Viewing Options:"
	@echo ""
	@echo "1. Swagger Editor (Recommended):"
	@echo "   - Open: https://editor.swagger.io/"
	@echo "   - Copy content from: ./api/swagger.json"
	@echo "   - Paste into Swagger Editor"
	@echo ""
	@echo "2. Swagger UI Online:"
	@echo "   - Open: http://petstore.swagger.io/"
	@echo "   - Click 'Try it out' -> 'Execute'"
	@echo "   - Or use: http://petstore.swagger.io/?url=http://localhost:8081/swagger.json"
	@echo "   - (requires local server: make swag-local)"
	@echo ""
	@echo "3. Local file server:"
	@echo "   make swag-local"
	@echo ""
	@echo "4. Direct file access:"
	@echo "   - JSON: ./api/swagger.json"
	@echo "   - YAML: ./api/swagger.yaml"

# Локальный файловый сервер для просмотра OpenAPI
swag-local:
	@echo "Starting local file server for OpenAPI documentation..."
	@echo "Open http://localhost:8081/swagger.json in your browser"
	@echo "Or use Swagger UI: http://petstore.swagger.io/?url=http://localhost:8081/swagger.json"
	@echo "Press Ctrl+C to stop"
	@python3 -m http.server 8081 --directory ./api

# Jaeger — это система распределенной трассировки
jaeger-up:
	docker run -d --name jaeger \
		-p 4318:4318 \
		-p 16686:16686 \
		jaegertracing/all-in-one

jaeger-stop:
	docker stop jaeger || true

jaeger-rm:
	docker rm jaeger || true

jaeger-down: jaeger-stop jaeger-rm

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build          - Собрать бинарный файл приложения"
	@echo "  make run            - Запустить приложение"
	@echo "  make dev            - Запустить с Air (hot reload)"
	@echo "  make dev-stdout     - Запустить с stdout логгированием для разработки"
	@echo "  make deps           - Скачать и обновить зависимости"
	@echo "  make lint           - Запустить линтер (golangci-lint)"
	@echo "  make fmt            - Форматировать код"
	@echo "  make vet            - Запустить go vet"
	@echo "  make test           - Запустить тесты с покрытием"
	@echo "  make test-curl      - Протестировать API через cURL"
	@echo "  make profile        - Запустить тест профилирования"
	@echo "  make profile-web    - Показать информацию о веб-интерфейсе pprof"
	@echo "  make profile-load   - Запустить полный тест профилирования с нагрузкой"
	@echo "  make swag           - Сгенерировать OpenAPI документацию"
	@echo "  make swag-serve     - Показать способы просмотра OpenAPI документации"
	@echo "  make swag-local     - Запустить локальный файловый сервер для просмотра OpenAPI"
	@echo "  make proto          - Сгенерировать Go код из proto файлов"
	@echo "  make proto-clean    - Очистить сгенерированные proto файлы"
	@echo "  make clean          - Очистить временные файлы"
	@echo "  make init           - Создать .env из .env.example"
	@echo "  make jaeger-up      - Запускает контейнер Jaeger в фоновом режиме с нужными портами."
	@echo "  make jaeger-stop    - Останавливает контейнер (если он запущен)"
	@echo "  make jaeger-rm      - Удаляет контейнер (если он существует)"
	@echo "  make jaeger-down    - Последовательно останавливает и удаляет контейнер"
	@echo ""
	@echo "Переменные окружения:"
	@echo "  USER_ID             - Идентификатор пользователя для HTTP запросов (по умолчанию 1)"
	@echo "  SERVER_URL          - Базовый URL API (по умолчанию http://localhost:8080/api/v1)"
	@echo "  make help           - Показать это сообщение"
