package transport

import (
	"context"
	"net/http"

	"api-cultura-conecta/internal/service"

	"github.com/gin-gonic/gin"
)

type CulturalWorksService interface {
	CreateCulturalWork(ctx context.Context, input service.CreateCulturalWorkInput) (service.CreateCulturalWorkOutput, error)
	GetCulturalWorks(ctx context.Context) ([]service.CreateCulturalWorkOutput, error)
}

type CulturalWorksHandler struct {
	culturalWorksService CulturalWorksService
}

func NewCulturalWorksHandler(culturalWorksService CulturalWorksService) *CulturalWorksHandler {
	return &CulturalWorksHandler{
		culturalWorksService: culturalWorksService,
	}
}

type CreateCulturalWorkRequest struct {
	Title      string `json:"title"`
	CategoryID int32  `json:"category_id"`
}

func (h *CulturalWorksHandler) CreateCulturalWork(c *gin.Context) {
	var req CreateCulturalWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	output, err := h.culturalWorksService.CreateCulturalWork(c.Request.Context(), service.CreateCulturalWorkInput{
		Title:      req.Title,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		FailErr(c, http.StatusInternalServerError, err, "Error al crear la obra cultural.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"cultural_work": output})
}

func (h *CulturalWorksHandler) GetCulturalWorks(c *gin.Context) {
	output, err := h.culturalWorksService.GetCulturalWorks(c.Request.Context())
	if err != nil {
		FailErr(c, http.StatusInternalServerError, err, "Error al obtener las obras culturales.")
		return
	}
	OK(c, http.StatusOK, gin.H{"cultural_works": output})
}
