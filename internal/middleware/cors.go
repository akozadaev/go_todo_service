package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS - middleware для настройки CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Accept-Language, Content-Type, Content-Length, Accept-Encoding, "+
				"X-CSRF-Token, Authorization, X-Requested-With, Origin, Cache-Control, "+
				"Pragma, Date, If-Modified-Since")
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Expose-Headers",
			"Content-Length, Content-Type, Date, Server")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 часа

		// Обработка preflight OPTIONS запросов
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
