package transport

import "github.com/gin-gonic/gin"

// SuccessResponse wraps any successful API payload.
type SuccessResponse[T any] struct {
	Data T `json:"data"`
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
