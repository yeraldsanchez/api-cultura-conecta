package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"api-cultura-conecta/internal/service"
)

type EventService interface {
	CreateEvent(ctx context.Context, input service.CreateEventInput) (service.EventOutput, error)
}

type EventHandler struct {
	svc EventService
}

func NewEventHandler(svc EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

type CreateEventRequest struct {
	Title       string    `json:"title"       binding:"required"`
	Description *string   `json:"description"`
	EventDate   time.Time `json:"event_date"  binding:"required"`
	Modality    string    `json:"modality"    binding:"required"`
	Link        *string   `json:"link"`
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	groupID, err := parsePathInt32(c, "group_id")
	if err != nil {
		return
	}
	userID := c.MustGet(UserIDKey).(int32)

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "Bad Request", "El cuerpo de la solicitud es inválido.")
		return
	}

	event, err := h.svc.CreateEvent(c.Request.Context(), service.CreateEventInput{
		GroupID:     groupID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
		Modality:    req.Modality,
		Link:        req.Link,
	})
	if err != nil {
		RespondError(c, err, "Error al crear el evento.")
		return
	}
	OK(c, http.StatusCreated, gin.H{"event": event})
}
