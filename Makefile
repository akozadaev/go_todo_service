.PHONY: run dev test-curl clean help

SERVER_URL = http://localhost:8080

# Запуск в production-режиме (без перезагрузки)
run:
	go run main.go

# Запуск в режиме разработки с Air (hot reload)
dev:
	air

# Выполнение последовательности cURL-запросов
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
	rm -rf tmp/
	rm -f build-errors.log

# Помощь
help:
	@echo "Цели:"
	@echo "  make run        — запустить сервер (обычный режим)"
	@echo "  make dev        — запустить сервер с Air (hot reload, для разработки)"
	@echo "  make test-curl  — выполнить тестовые cURL-запросы к http://localhost:8080"
	@echo "  make clean      — удалить временные файлы Air"
	@echo "  make help       — показать это сообщение"