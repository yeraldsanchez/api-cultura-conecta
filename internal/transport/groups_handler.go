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
	JoinGroup(ctx context.Context, groupID int32, userID int32) error
	CreatePost(ctx context.Context, input service.CreatePostInput) (service.PostOutput, error)
	GetSuggestedGroups(ctx context.Context, input service.SuggestGroupsInput) (service.ListGroupsOutput, error)
	GetGroupsByMember(ctx context.Context, userID int32) ([]service.UserGroupOutput, error)
	GetGroupMembers(ctx context.Context, groupID int32) ([]service.GroupMemberOutput, error)
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
		RespondError(c, err, "Error al obtener los grupos.")
		return
	}

	OK(c, http.StatusOK, PaginatedOutput[service.GroupOutput]{
		Items:      result.Groups,
		TotalCount: result.Total,
		Page:       int32(page),
		Limit:      int32(limit),
	})
}

func (h *GroupHandler) JoinGroup(c *gin.Context) {
	groupID, err := parsePathInt32(c, "group_id")
	if err != nil {
		return
	}
	userID := c.MustGet(UserIDKey).(int32)

	if err := h.svc.JoinGroup(c.Request.Context(), groupID, userID); err != nil {
		RespondError(c, err, "Error al unirse al grupo.")
		return
	}
	c.Status(http.StatusNoContent)
}

type CreatePostRequest struct {
	Content         string  `json:"content" binding:"required"`
	HasSpoiler      bool    `json:"has_spoiler"`
	SpoilerProgress *string `json:"spoiler_progress"`
}

func (h *GroupHandler) CreatePost(c *gin.Context) {
	groupID, err := parsePathInt32(c, "group_id")
	if err != nil {
		return
	}
	userID := c.MustGet(UserIDKey).(int32)

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	post, err := h.svc.CreatePost(c.Request.Context(), service.CreatePostInput{
		GroupID:         groupID,
		UserID:          userID,
		Content:         req.Content,
		HasSpoiler:      req.HasSpoiler,
		SpoilerProgress: req.SpoilerProgress,
	})
	if err != nil {
		RespondError(c, err, "Error al publicar el mensaje.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"post": post})
}

func (h *GroupHandler) GetSuggestedGroups(c *gin.Context) {
	userID := c.MustGet(UserIDKey).(int32)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	result, err := h.svc.GetSuggestedGroups(c.Request.Context(), service.SuggestGroupsInput{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		RespondError(c, err, "Error al obtener sugerencias de grupos.")
		return
	}
	OK(c, http.StatusOK, PaginatedOutput[service.GroupOutput]{
		Items:      result.Groups,
		TotalCount: result.Total,
		Page:       int32(page),
		Limit:      int32(limit),
	})
}

func (h *GroupHandler) GetGroupsByMember(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}

	groups, err := h.svc.GetGroupsByMember(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err, "Error al obtener los grupos del usuario.")
		return
	}
	OK(c, http.StatusOK, gin.H{"groups": groups})
}

func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupID, err := parsePathInt32(c, "group_id")
	if err != nil {
		return
	}

	members, err := h.svc.GetGroupMembers(c.Request.Context(), groupID)
	if err != nil {
		RespondError(c, err, "Error al obtener los miembros del grupo.")
		return
	}
	OK(c, http.StatusOK, gin.H{"members": members})
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
		RespondError(c, err, "Error al crear el grupo.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"group": group})
}
