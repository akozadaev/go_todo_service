package middleware

import (
	"net/http"
	"strconv"

	"github.com/akozadaev/go_todo_service/internal/userctx"
	"github.com/gin-gonic/gin"
)

const userIDHeader = "X-User-ID"

// RequireUser ensures that each request carries a valid user identifier.
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue := c.GetHeader(userIDHeader)
		if userIDValue == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-User-ID header",
			})
			return
		}

		userID, err := strconv.ParseUint(userIDValue, 10, 64)
		if err != nil || userID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid X-User-ID header",
			})
			return
		}

		ctx := userctx.WithUserID(c.Request.Context(), uint(userID))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
