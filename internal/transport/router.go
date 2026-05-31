package transport

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	r *gin.Engine,
	auth *AuthHandler,
) {
	v1 := r.Group("/api/v1")
	authG := v1.Group("/auth")
	{
		authG.POST("/register", auth.Register)
		authG.POST("/login", auth.Login)
	}
}
