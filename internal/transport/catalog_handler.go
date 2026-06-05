package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CatalogService interface {
	GetFocusTypes(ctx context.Context) ([]service.FocusTypeOutput, error)
	CreateFocusType(ctx context.Context, input service.FocusTypeInput) (service.FocusTypeOutput, error)
	GetInterests(ctx context.Context) ([]service.InterestOutput, error)
	CreateInterest(ctx context.Context, input service.InterestInput) (service.InterestOutput, error)
}

type CatalogHandler struct {
	svc CatalogService
}

func NewCatalogHandler(svc CatalogService) *CatalogHandler {
	return &CatalogHandler{
		svc: svc,
	}
}

type CreateFocusTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateInterestRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *CatalogHandler) CreateFocusType(c *gin.Context) {
	var req CreateFocusTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	focusType, err := h.svc.CreateFocusType(c.Request.Context(), service.FocusTypeInput{Name: req.Name})
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Internal Server Error", "Error al crear el tipo de enfoque.")
		return
	}

	OK(c, http.StatusCreated, gin.H{"focus_type": focusType})
}

func (h *CatalogHandler) GetFocusTypes(c *gin.Context) {
	focusTypes, err := h.svc.GetFocusTypes(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Internal Server Error", "Error al obtener los tipos de enfoque.")
		return
	}
	OK(c, http.StatusOK, gin.H{"focus_types": focusTypes})
}

func (h *CatalogHandler) CreateInterest(c *gin.Context) {
	var req CreateInterestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	interest, err := h.svc.CreateInterest(c.Request.Context(), service.InterestInput{Name: req.Name})
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Internal Server Error", "Error al crear el interés.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"interest": interest})
}

func (h *CatalogHandler) GetInterests(c *gin.Context) {
	interests, err := h.svc.GetInterests(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Internal Server Error", "Error al obtener los intereses.")
		return
	}
	OK(c, http.StatusOK, gin.H{"interests": interests})
}
