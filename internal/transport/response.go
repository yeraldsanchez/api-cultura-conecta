package transport

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SuccessResponse wraps any successful API payload.
type SuccessResponse[T any] struct {
	Data T `json:"data"`
}

// PaginatedOutput wraps a paginated list result.
type PaginatedOutput[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Page       int32 `json:"page"`
	Limit      int32 `json:"limit"`
}

// FieldError describes a validation error on a specific request field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse is the standard error envelope returned on all failures.
type ErrorResponse struct {
	Status  int          `json:"status"`
	Error   string       `json:"error"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// OK writes a 2xx JSON response with the given data wrapped in SuccessResponse.
func OK[T any](c *gin.Context, statusCode int, data T) {
	c.JSON(statusCode, SuccessResponse[T]{Data: data})
}

// Fail writes an error JSON response using the standard ErrorResponse envelope.
func Fail(c *gin.Context, status int, errText, message string, details ...FieldError) {
	c.JSON(status, ErrorResponse{
		Status:  status,
		Error:   errText,
		Message: message,
		Details: details,
	})
}

// FailErr is like Fail but accepts the real Go error so Sentry can capture the
// full error message and stack trace. The client still receives only the generic
// HTTP status text — the internal error never leaves the server.
func FailErr(c *gin.Context, status int, err error, message string, details ...FieldError) {
	if status >= 500 {
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
				scope.SetTag("endpoint", c.FullPath())
				scope.SetTag("method", c.Request.Method)
				hub.CaptureException(err)
			})
		}
	}

	c.JSON(status, ErrorResponse{
		Status:  status,
		Error:   http.StatusText(status),
		Message: message,
		Details: details,
	})
}
