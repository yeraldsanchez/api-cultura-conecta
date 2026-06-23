package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-cultura-conecta/internal/apperrors"
	db "api-cultura-conecta/internal/db/generated"
)

type EventService struct {
	pool *pgxpool.Pool
}

func NewEventService(pool *pgxpool.Pool) *EventService {
	return &EventService{pool: pool}
}

type CreateEventInput struct {
	GroupID     int32
	UserID      int32
	Title       string
	Description *string
	EventDate   time.Time
	Modality    string
	Link        *string
}

type EventOutput struct {
	ID          int32     `json:"id"`
	GroupID     int32     `json:"group_id"`
	CreatedBy   int32     `json:"created_by"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	EventDate   time.Time `json:"event_date"`
	Modality    string    `json:"modality"`
	Link        *string   `json:"link"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *EventService) CreateEvent(ctx context.Context, input CreateEventInput) (EventOutput, error) {
	if input.Modality != "in-person" && input.Modality != "virtual" {
		return EventOutput{}, apperrors.NewValidationError("modality must be 'in-person' or 'virtual'")
	}
	if input.Modality == "virtual" && input.Link == nil {
		return EventOutput{}, apperrors.NewValidationError("link is required for virtual events")
	}

	var out EventOutput
	err := withTx(ctx, s.pool, func(q db.Querier) error {
		isMember, err := q.IsGroupMember(ctx, db.IsGroupMemberParams{
			GroupID: input.GroupID,
			UserID:  input.UserID,
		})
		if err != nil {
			return err
		}
		if !isMember {
			return apperrors.ErrNotGroupMember
		}

		event, err := q.CreateEvent(ctx, db.CreateEventParams{
			GroupID:     input.GroupID,
			CreatedBy:   input.UserID,
			Title:       input.Title,
			Description: input.Description,
			EventDate:   input.EventDate,
			Modality:    input.Modality,
			Link:        input.Link,
		})
		if err != nil {
			return apperrors.FromPgx(err, apperrors.EventsConstraints)
		}

		out = EventOutput{
			ID:          event.ID,
			GroupID:     event.GroupID,
			CreatedBy:   event.CreatedBy,
			Title:       event.Title,
			Description: event.Description,
			EventDate:   event.EventDate,
			Modality:    event.Modality,
			Link:        event.Link,
			CreatedAt:   event.CreatedAt,
		}
		return nil
	})
	return out, err
}
