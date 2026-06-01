package transport

import (
	"api-cultura-conecta/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserProfileHandler struct {
	svc *service.UserProfileService
}

func NewUserProfileHandler(svc *service.UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{
		svc: svc,
	}
}

type CreateProfileRequest struct {
	UserID       int32   `json:"user_id" binding:"required"`
	DepthLevel   string  `json:"depth_level" binding:"required"`
	FocusIDs     []int32 `json:"focus_ids" binding:"required"`
	InterestsIDs []int32 `json:"interests_ids" binding:"required"`
}

func (h *UserProfileHandler) CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request",
			"El cuerpo de la solicitud es inválido.")
		return
	}

	profile, err := h.svc.Create(c.Request.Context(), service.CreateProfileInput{
		UserID:       req.UserID,
		DepthLevel:   req.DepthLevel,
		FocusIDs:     req.FocusIDs,
		InterestsIDs: req.InterestsIDs,
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, "Bad Request",
			"Error al crear el perfil del usuario.")
		return
	}

	OK(c, http.StatusCreated, gin.H{"profile": profile})
}
