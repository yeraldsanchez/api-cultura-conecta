package transport

import (
	"api-cultura-conecta/internal/service"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserProfileService interface {
	Create(ctx context.Context, input service.CreateProfileInput) (service.ProfileOutput, error)
	GetProfile(ctx context.Context, userID int32) (service.ProfileOutput, error)
	UpdateProfile(ctx context.Context, input service.UpdateProfileInput) (service.ProfileOutput, error)
	AddInterest(ctx context.Context, userID int32, categoryID int32) (service.ProfileOutput, error)
	RemoveInterest(ctx context.Context, userID int32, categoryID int32) (service.ProfileOutput, error)
	AddFocusType(ctx context.Context, userID int32, focusTypeID int32) (service.ProfileOutput, error)
	RemoveFocusType(ctx context.Context, userID int32, focusTypeID int32) (service.ProfileOutput, error)
}

type UserProfileHandler struct {
	svc UserProfileService
}

func NewUserProfileHandler(svc UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{
		svc: svc,
	}
}

type CreateProfileRequest struct {
	UserID       int32   `json:"user_id" binding:"required"`
	DepthLevel   string  `json:"depth_level" binding:"required"`
	FocusIDs     []int32 `json:"focus_ids" binding:"required"`
	Name         string  `json:"name" binding:"required"`
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
		Name:         req.Name,
		UserID:       req.UserID,
		DepthLevel:   req.DepthLevel,
		FocusIDs:     req.FocusIDs,
		InterestsIDs: req.InterestsIDs,
	})
	if err != nil {
		RespondError(c, err, "Error al crear el perfil del usuario.")
		return
	}

	OK(c, http.StatusCreated, gin.H{"profile": profile})
}

type PatchProfileRequest struct {
	Name       *string `json:"name"`
	DepthLevel *string `json:"depth_level"`
}

func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err, "Error al obtener el perfil del usuario.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *UserProfileHandler) PatchProfile(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	var req PatchProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}
	if req.Name == nil && req.DepthLevel == nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "Debe enviar al menos un campo a actualizar.")
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), service.UpdateProfileInput{
		UserID:     userID,
		Name:       req.Name,
		DepthLevel: req.DepthLevel,
	})
	if err != nil {
		RespondError(c, err, "Error al actualizar el perfil.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

type AddInterestRequest struct {
	CategoryID int32 `json:"category_id" binding:"required"`
}

func (h *UserProfileHandler) AddInterest(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	var req AddInterestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}
	profile, err := h.svc.AddInterest(c.Request.Context(), userID, req.CategoryID)
	if err != nil {
		RespondError(c, err, "Error al agregar el interés.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *UserProfileHandler) RemoveInterest(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	categoryID, err := parsePathInt32(c, "category_id")
	if err != nil {
		return
	}
	profile, err := h.svc.RemoveInterest(c.Request.Context(), userID, categoryID)
	if err != nil {
		RespondError(c, err, "Error al eliminar el interés.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

type AddFocusTypeRequest struct {
	FocusTypeID int32 `json:"focus_type_id" binding:"required"`
}

func (h *UserProfileHandler) AddFocusType(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	var req AddFocusTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}
	profile, err := h.svc.AddFocusType(c.Request.Context(), userID, req.FocusTypeID)
	if err != nil {
		RespondError(c, err, "Error al agregar el tipo de enfoque.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *UserProfileHandler) RemoveFocusType(c *gin.Context) {
	userID, err := parsePathInt32(c, "user_id")
	if err != nil {
		return
	}
	focusTypeID, err := parsePathInt32(c, "focus_type_id")
	if err != nil {
		return
	}
	profile, err := h.svc.RemoveFocusType(c.Request.Context(), userID, focusTypeID)
	if err != nil {
		RespondError(c, err, "Error al eliminar el tipo de enfoque.")
		return
	}
	OK(c, http.StatusOK, gin.H{"profile": profile})
}

func parsePathInt32(c *gin.Context, param string) (int32, error) {
	raw := c.Param(param)
	val, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El parámetro '"+param+"' debe ser un número entero.")
		return 0, err
	}
	return int32(val), nil
}
