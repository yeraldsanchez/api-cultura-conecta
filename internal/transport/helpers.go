package transport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"api-cultura-conecta/internal/apperrors"
)

func RespondValidationError(c *gin.Context, err error) bool {
	if ve, ok := errors.AsType[*apperrors.ValidationError](err); ok {
		Fail(c, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), ve.Message)
		return true
	}
	return false
}

func RespondError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, apperrors.ErrValidation):
		RespondValidationError(c, err)
	case errors.Is(err, apperrors.ErrNotFound):
		Fail(c, http.StatusNotFound, http.StatusText(http.StatusNotFound), err.Error())
	case errors.Is(err, apperrors.ErrConflict):
		Fail(c, http.StatusConflict, http.StatusText(http.StatusConflict), err.Error())
	case errors.Is(err, apperrors.ErrUnauthorized):
		Fail(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err.Error())
	default:
		FailErr(c, http.StatusInternalServerError, err, fallback)
	}
}
