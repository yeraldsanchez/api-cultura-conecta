package transport

import (
	"bytes"
	"io"
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// bodyCapture reads the request body for POST/PUT/PATCH requests and attaches
// it to the Sentry scope so it's visible on error events.
func bodyCapture() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 10_000))
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.Scope().SetRequestBody(bodyBytes)
				}
			}
		}
		c.Next()
	}
}

func RegisterRoutes(
	r *gin.Engine,
	auth *AuthHandler,
	user *UserProfileHandler,
	catalog *CatalogHandler,
	culturalWorks *CulturalWorksHandler,
	group *GroupHandler,
) {
	v1 := r.Group("/api/v1")
	v1.Use(bodyCapture())
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
	culturalG := v1.Group("/cultural-works")
	{
		culturalG.GET("", culturalWorks.GetCulturalWorks)
		culturalG.POST("", culturalWorks.CreateCulturalWork)
	}
	groupG := v1.Group("/groups")
	{
		groupG.GET("", group.ListGroups)
		groupG.POST("", group.CreateGroup)
	}
}
