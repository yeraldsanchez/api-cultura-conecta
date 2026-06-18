package apperrors

import (
	"errors"
	"fmt"
)

// Sentinels de categoría — se usan con errors.Is para determinar el HTTP status.
var (
	ErrNotFound     = errors.New("record not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation error")
)

// Errores de dominio tipados — cada uno wrappea su sentinel de categoría,
// por lo que errors.Is(err, ErrConflict) es true para todos los errores de conflicto,
// y err.Error() devuelve el mensaje específico del dominio.
var (
	ErrDuplicateEmail       = newConflict("email already exists")
	ErrDuplicateName        = newConflict("name already exists")
	ErrProfileDuplicateUser = newConflict("profile with this user already exists")

	ErrCategoryNotFound     = newNotFound("category does not exist")
	ErrUserNotFound         = newNotFound("user not found")
	ErrCulturalWorkNotFound = newNotFound("cultural work does not exist")
	ErrFocusTypeNotFound    = newNotFound("focus type does not exist")
	ErrGroupNotFound        = newNotFound("group not found")

	ErrAlreadyMember  = newConflict("user is already a member of this group")
	ErrNotGroupMember = newUnauthorized("user is not a member of this group")

	ErrInvalidCredentials = newUnauthorized("invalid email or password")
)

type typedError struct {
	msg    string
	parent error
}

func (e *typedError) Error() string { return e.msg }
func (e *typedError) Unwrap() error { return e.parent }

func newNotFound(msg string) error     { return &typedError{msg, ErrNotFound} }
func newConflict(msg string) error     { return &typedError{msg, ErrConflict} }
func newUnauthorized(msg string) error { return &typedError{msg, ErrUnauthorized} }

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

func NewValidationError(msg string) error {
	return &ValidationError{Message: msg}
}
