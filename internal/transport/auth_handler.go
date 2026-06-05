package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(ctx context.Context, input service.CreateUserInput) (*int32, error)
	Login(ctx context.Context, input service.LoginInput) (string, error)
}

type AuthHandler struct {
	svc AuthService
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request",
			"El cuerpo de la solicitud es inválido.")
		return
	}

	userID, err := h.svc.Register(c.Request.Context(), service.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request",
			"Error al registrar el usuario.",
			FieldError{Field: "email", Message: "Este correo electrónico ya está registrado."})
		return
	}

	OK(c, http.StatusCreated, gin.H{"user_id": userID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	token, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		Fail(c, http.StatusUnauthorized, "Unauthorized", "Credenciales inválidas.")
		return
	}
	OK(c, http.StatusOK, gin.H{"token": token})
}
