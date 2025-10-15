package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PProfAuth middleware для защиты pprof endpoints
// В продакшене должен быть заменен на более строгую аутентификацию
func PProfAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// В режиме разработки разрешаем доступ без аутентификации
		if gin.Mode() == gin.DebugMode {
			c.Next()
			return
		}

		// В продакшене можно добавить проверку токена или IP
		// Пока просто блокируем доступ в ReleaseMode
		c.JSON(http.StatusForbidden, gin.H{
			"error": "pprof endpoints are disabled in production mode",
		})
		c.Abort()
	}
}

// PProfAuthWithToken middleware с проверкой токена для pprof endpoints
func PProfAuthWithToken(secretToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем токен из заголовка Authorization
		token := c.GetHeader("Authorization")
		if token != "Bearer "+secretToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized access to pprof endpoints",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

