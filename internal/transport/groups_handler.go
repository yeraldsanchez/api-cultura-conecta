package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupService interface {
	CreateGroup(ctx context.Context, input service.CreateGroupInput) (service.GroupOutput, error)
}

type GroupHandler struct {
	svc GroupService
}

func NewGroupHandler(svc GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

type CreateGroupRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"omitempty"`
	WorkID        int32   `json:"work_id" binding:"required"`
	CreatedBy     int32   `json:"created_by" binding:"required"`
	DepthLevel    string  `json:"depth_level" binding:"required"`
	CategoriesIDs []int32 `json:"categories_ids" binding:"required"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	group, err := h.svc.CreateGroup(c.Request.Context(), service.CreateGroupInput{
		Name:          req.Name,
		Description:   req.Description,
		WorkID:        req.WorkID,
		CreatedBy:     req.CreatedBy,
		DepthLevel:    req.DepthLevel,
		CategoriesIDs: req.CategoriesIDs,
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Internal Server Error", "Error al crear el grupo.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"group": group})

}
