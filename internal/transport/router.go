package transport

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	r *gin.Engine,
	auth *AuthHandler,
	user *UserProfileHandler,
	catalog *CatalogHandler,
) {
	v1 := r.Group("/api/v1")
	authG := v1.Group("/auth")
	{
		authG.POST("/register", auth.Register)
		authG.POST("/login", auth.Login)
	}
	userG := v1.Group("/users")
	{
		userG.POST("", user.CreateProfile)
	}
	interestG := v1.Group("/interests")
	{
		interestG.GET("", catalog.GetInterests)
		interestG.POST("", catalog.CreateInterest)
	}
	focusTG := v1.Group("/focus-types")
	{
		focusTG.GET("", catalog.GetFocusTypes)
		focusTG.POST("", catalog.CreateFocusType)
	}
}
