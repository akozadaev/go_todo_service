В проекте TODO-микросервиса используется веб-фреймворк **Gin** (https://gin-gonic.com/). Ниже приведена справочная информация о методах Gin, которые применяются в коде.

---

## Основные методы Gin, используемые в проекте

### 1. `gin.Default()`
Создаёт экземпляр Gin-роутера с предустановленными middleware:
- `Logger()` — логирует все HTTP-запросы в stdout.
- `Recovery()` — перехватывает паники и возвращает ошибку 500 вместо падения сервера.

```go
r := gin.Default()
```

---

### 2. HTTP-методы роутинга

Эти методы регистрируют обработчики для соответствующих HTTP-глаголов.

#### `r.GET(path string, handlers ...gin.HandlerFunc)`
Регистрирует обработчик для **GET**-запросов.

Пример:
```go
r.GET("/todos", handler.GetTodos)
```

#### `r.POST(path string, handlers ...gin.HandlerFunc)`
Регистрирует обработчик для **POST**-запросов (обычно для создания ресурсов).

Пример:
```go
r.POST("/todos", handler.CreateTodo)
```

#### `r.PUT(path string, handlers ...gin.HandlerFunc)`
Регистрирует обработчик для **PUT**-запросов (полное обновление ресурса).

Пример:
```go
r.PUT("/todos/:id", handler.UpdateTodo)
```

#### `r.DELETE(path string, handlers ...gin.HandlerFunc)`
Регистрирует обработчик для **DELETE**-запросов.

Пример:
```go
r.DELETE("/todos/:id", handler.DeleteTodo)
```

> Все эти методы принимают:
> - путь (может содержать параметры, например `:id`)
> - одну или несколько функций-обработчиков (`gin.HandlerFunc`)

---

### 3. Контекст запроса: `*gin.Context`

Объект `c *gin.Context` передаётся в каждый обработчик и предоставляет доступ к:

#### `c.Param(key string) string`
Получает значение из URL-параметра.

Пример:
```go
id := c.Param("id") // из пути /todos/:id
```

#### `c.ShouldBindJSON(obj interface{}) error`
Десериализует тело запроса из JSON в структуру Go. Возвращает ошибку, если JSON невалиден или не соответствует структуре.

Пример:
```go
var todo models.Todo
if err := c.ShouldBindJSON(&todo); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

> Аналоги: `ShouldBind`, `BindJSON` и др. — `ShouldBindJSON` предпочтителен, так как не прерывает выполнение при ошибке.

#### `c.JSON(code int, obj interface{})`
Отправляет ответ в формате JSON с указанным HTTP-статусом.

Пример:
```go
c.JSON(http.StatusOK, todos)
c.JSON(http.StatusCreated, todo)
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
```

> `gin.H` — это просто псевдоним для `map[string]interface{}`.

#### `c.AbortWithStatus(code int)`
Немедленно прерывает обработку запроса и отправляет HTTP-статус без тела.

Пример:
```go
c.AbortWithStatus(http.StatusNoContent) // 204
```

(В проекте используется `c.JSON(http.StatusNoContent, nil)`, что эквивалентно по эффекту.)

---

## Дополнительно: Тип обработчика

```go
type HandlerFunc func(*Context)
```

Каждый обработчик — это функция, принимающая указатель на `gin.Context`.

---

## Пример полного обработчика

```go
func (h *TodoHandler) CreateTodo(c *gin.Context) {
    var todo models.Todo
    if err := c.ShouldBindJSON(&todo); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.DB.Create(&todo).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
        return
    }
    c.JSON(http.StatusCreated, todo)
}
```

---

## Полезные ссылки

- Официальная документация Gin: https://gin-gonic.com/docs/
- GitHub: https://github.com/gin-gonic/gin
- Примеры: https://github.com/gin-gonic/examples
