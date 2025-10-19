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
- **Логгирование**: [Zap](https://github.com/uber-go/zap) с ротацией через [Lumberjack](https://github.com/natefinch/lumberjack)
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
| `LOG_LEVEL`             | Уровень логирования (debug, info, warn, error, fatal) | info         |
| `LOG_FORMAT`            | Формат логов (json, text)   | json         |
| `LOG_FILENAME`          | Путь к файлу лога (пустая строка = только stdout) | logs/app.log |
| `LOG_MAX_SIZE`          | Макс. размер файла (MB)     | 100          |
| `LOG_MAX_AGE`           | Макс. возраст файла (дни)   | 30           |
| `LOG_MAX_BACKUPS`       | Макс. количество бэкапов    | 5            |
| `LOG_COMPRESS`          | Сжимать старые файлы        | true         |
| `LOG_LOCAL_TIME`        | Использовать локальное время| true         |
| `LOG_ROTATE_DAILY`      | Ротация по дням             | false        |
| `LOG_ENABLE_STDOUT`     | Дублировать в stdout        | false        |
| `TRACE_ENABLED`         | Включить трассировку        | false        |
| `TRACE_URL`             | URL Jaeger коллектора       | http://localhost:14268/api/traces |
| `TRACE_SERVICE_NAME`    | Имя сервиса для трассировки | go-todo-service |
| `TRACE_HTTP_BODY_ENABLED` | Логировать HTTP body       | false        |

## Логгирование

Проект использует структурированное логгирование на основе библиотеки [Zap](https://github.com/uber-go/zap) с поддержкой ротации логов через [Lumberjack](https://github.com/natefinch/lumberjack).

### Особенности

- **Структурированные логи**: JSON формат для production, текстовый для разработки
- **Ротация логов**: По размеру файла и времени с настраиваемыми параметрами
- **Request ID**: Уникальный идентификатор для каждого HTTP запроса
- **Уровни логгирования**: debug, info, warn, error, fatal
- **Специализированные логгеры**: Отдельные файлы для HTTP, ошибок, доступа и аудита
- **Контекстное логгирование**: Автоматическое добавление request ID и других полей

### Режимы работы

#### Production режим
```bash
# Логи только в файл (JSON формат)
LOG_FORMAT=json
LOG_FILENAME=logs/app.log
LOG_ENABLE_STDOUT=false
```

#### Development режим
```bash
# Логи в stdout (текстовый формат)
LOG_FORMAT=text
LOG_FILENAME=
LOG_ENABLE_STDOUT=true
```

#### Гибридный режим
```bash
# Логи и в файл, и в stdout
LOG_FORMAT=json
LOG_FILENAME=logs/app.log
LOG_ENABLE_STDOUT=true
```

### Структура логов

#### HTTP запросы
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/v1/todos",
  "status": 200,
  "latency": "15ms",
  "client_ip": "127.0.0.1",
  "user_agent": "curl/7.68.0",
  "body_size": 1024
}
```

#### Ошибки приложения
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "ERROR",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Failed to fetch todo",
  "todo_id": 123,
  "error": "database connection failed"
}
```

### Файлы логов

- `logs/app.log` - Основные логи приложения
- `logs/http.log` - HTTP запросы и ответы
- `logs/error.log` - Ошибки приложения
- `logs/access.log` - События доступа
- `logs/audit.log` - Аудит действий

### Команды для разработки

```bash
# Запуск с stdout логгированием
make dev-stdout

# Запуск с Air (hot reload)
make dev

# Обычный запуск
make run
```

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

