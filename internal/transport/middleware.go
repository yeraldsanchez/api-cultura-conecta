package transport

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func AuthMiddleware(svc AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			Fail(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized),
				"Se requiere autenticación.")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := svc.ValidateAccessToken(tokenStr)
		if err != nil {
			Fail(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized),
				"Token inválido o expirado.")
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}
