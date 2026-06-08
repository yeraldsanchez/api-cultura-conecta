package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GroupService interface {
	CreateGroup(ctx context.Context, input service.CreateGroupInput) (service.GroupOutput, error)
	ListGroups(ctx context.Context, input service.ListGroupsInput) (service.ListGroupsOutput, error)
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

func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	input := service.ListGroupsInput{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if v := c.Query("work_id"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			Fail(c, http.StatusBadRequest, "Bad Request", "work_id debe ser un número entero.")
			return
		}
		input.WorkID = &n
	}
	if v := c.Query("name"); v != "" {
		input.Name = &v
	}
	if v := c.Query("depth_level"); v != "" {
		input.DepthLevel = &v
	}
	if v := c.Query("categories_ids"); v != "" {
		for _, part := range strings.Split(v, ",") {
			n, err := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
			if err != nil {
				Fail(c, http.StatusBadRequest, "Bad Request", "categories_ids debe ser una lista de enteros separados por coma.")
				return
			}
			input.FocusTypeIDs = append(input.FocusTypeIDs, int32(n))
		}
	}

	result, err := h.svc.ListGroups(c.Request.Context(), input)
	if err != nil {
		FailErr(c, http.StatusInternalServerError, err, "Error al obtener los grupos.")
		return
	}

	OK(c, http.StatusOK, PaginatedOutput[service.GroupOutput]{
		Items:      result.Groups,
		TotalCount: result.Total,
		Page:       int32(page),
		Limit:      int32(limit),
	})
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
		FailErr(c, http.StatusInternalServerError, err, "Error al crear el grupo.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"group": group})
}
