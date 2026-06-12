package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(ctx context.Context, input service.CreateUserInput) (*int32, error)
	Login(ctx context.Context, input service.LoginInput) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateAccessToken(tokenStr string) (int32, error)
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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
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
		RespondError(c, err, "Error al registrar el usuario.")
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

	accessToken, refreshToken, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		RespondError(c, err, "Credenciales inválidas.")
		return
	}

	OK(c, http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	accessToken, err := h.svc.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		RespondError(c, err, "Refresh token inválido o expirado.")
		return
	}

	OK(c, http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		RespondError(c, err, "Error al cerrar sesión.")
		return
	}

	c.Status(http.StatusNoContent)
}
