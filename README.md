# TODO Микросервис на Go

RESTful микросервис для управления списком задач (TODO), построенный с использованием лучших практик Go.

## Архитектура

Структура папок в соответствии с [рекомендациями](https://github.com/golang-standards/project-layout/blob/master/README_ru.md)

gitПроект следует **Clean Architecture** и **SOLID** принципам:

### Слои приложения

```
HTTP Request -> Handler -> Service -> Repository -> Database

```

## Технологии

- **Язык**: Go 1.25
- **Веб-фреймворк**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **База данных**: PostgreSQL
- **Конфигурация**: Environment variables
- **Hot Reload**: [Air](https://github.com/cosmtrek/air) (для разработки)

## Требования

### Для локальной разработки:
- Go 1.25
- PostgreSQL 12+
- Make (опционально)
- Air (для hot reload)

## ️ Установка и запуск

### 1. Установка зависимостей

```bash
# Установить Go зависимости
make deps

# Установить Air для hot reload
go install github.com/cosmtrek/air@latest
```

### 2. Настройка базы данных

Создайте базу данных в PostgreSQL:

```sql
CREATE DATABASE todo_db;
```

Создайте `.env` файл:

```bash
make init
```

Отредактируйте `.env`:

```env
SERVER_HOST=localhost
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=todo_db
DB_SSLMODE=disable
```

### 3. Запуск приложения

**Режим разработки с hot reload:**

```bash
make dev
```

**Обычный запуск:**

```bash
make run
```

**Сборка и запуск бинарника:**

```bash
make build
./bin/api
```

##  API Endpoints

### Health & Monitoring

| Метод | Маршрут   | Описание                    |
|-------|-----------|-----------------------------|
| GET   | `/health` | Проверка работы сервиса     |
| GET   | `/ready`  | Проверка готовности (БД)    |

### Todo API (префикс `/api/v1`)

| Метод  | Маршрут            | Описание                |
|--------|--------------------|-------------------------|
| GET    | `/api/v1/todos`    | Получить все задачи     |
| GET    | `/api/v1/todos/:id`| Получить задачу по ID   |
| POST   | `/api/v1/todos`    | Создать новую задачу    |
| PUT    | `/api/v1/todos/:id`| Обновить задачу         |
| DELETE | `/api/v1/todos/:id`| Удалить задачу          |

### Формат данных

**Todo объект:**

```json
{
  "id": 1,
  "title": "Купить продукты",
  "description": "Молоко, хлеб, яйца",
  "done": false,
  "created_at": "2025-10-12T10:00:00Z",
  "updated_at": "2025-10-12T10:00:00Z"
}
```

**Создание/обновление (request body):**

```json
{
  "title": "Название задачи",
  "description": "Описание (опционально)",
  "done": false
}
```

### Примеры запросов

**Получить все задачи:**
```bash
curl http://localhost:8080/api/v1/todos
```

**Создать задачу:**
```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Купить продукты",
    "description": "Молоко, хлеб, яйца",
    "done": false
  }'
```

**Обновить задачу:**
```bash
curl -X PUT http://localhost:8080/api/v1/todos/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Купить продукты",
    "description": "Молоко, хлеб, яйца",
    "done": true
  }'
```

**Удалить задачу:**
```bash
curl -X DELETE http://localhost:8080/api/v1/todos/1
```

## Тестирование

### Автоматическое тестирование API

```bash
make test-curl
```

Эта команда последовательно протестирует все endpoints.

### Запуск unit-тестов

```bash
make test
```

Отчет о покрытии будет сохранен в `coverage.html`.

##  Команды Makefile

```bash
make help              # Показать все доступные команды
make build             # Собрать бинарник
make run               # Запустить приложение
make dev               # Запустить с hot reload (Air)
make deps              # Установить зависимости
make lint              # Запустить линтер
make fmt               # Форматировать код
make test              # Запустить тесты
make test-curl         # Протестировать API
make clean             # Очистить временные файлы
make init              # Создать .env из .env.example
```

## Конфигурация

Все настройки задаются через переменные окружения (`.env` файл):

| Переменная              | Описание                    | По умолчанию |
|-------------------------|-----------------------------| ------------ |
| `SERVER_HOST`           | Хост сервера                | localhost    |
| `SERVER_PORT`           | Порт сервера                | 8080         |
| `SERVER_READ_TIMEOUT`   | Таймаут чтения (сек)        | 10           |
| `SERVER_WRITE_TIMEOUT`  | Таймаут записи (сек)        | 10           |
| `SERVER_SHUTDOWN_TIMEOUT`| Таймаут завершения (сек)   | 5            |
| `DB_HOST`               | Хост PostgreSQL             | localhost    |
| `DB_PORT`               | Порт PostgreSQL             | 5432         |
| `DB_USER`               | Пользователь БД             | postgres     |
| `DB_PASSWORD`           | Пароль БД                   | (обязателен) |
| `DB_NAME`               | Имя базы данных             | todo_db      |
| `DB_SSLMODE`            | SSL режим PostgreSQL        | disable      |
| `DB_MAX_IDLE_CONNS`     | Макс. idle соединений       | 10           |
| `DB_MAX_OPEN_CONNS`     | Макс. открытых соединений   | 100          |


### Code Style

```bash
# Форматирование
make fmt

# Проверка
make vet

# Линтер
make lint
```

Генерация OpenAPI
```bash
swag init -g ./cmd/api/main.go -o ./api 
```

