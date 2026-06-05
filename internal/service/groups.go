package service

import (
	db "api-cultura-conecta/internal/db/generated"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupService struct {
	q    db.Querier
	pool *pgxpool.Pool
}

func NewGroupService(q db.Querier, pool *pgxpool.Pool) *GroupService {
	return &GroupService{q: q, pool: pool}
}

type CreateGroupInput struct {
	Name          string  `json:"name" validate:"required"`
	WorkID        int32   `json:"work_id" validate:"required"`
	CreatedBy     int32   `json:"created_by" validate:"required"`
	Description   string  `json:"description"`
	DepthLevel    string  `json:"depth_level" validate:"required"`
	CategoriesIDs []int32 `json:"interests" validate:"required"`
}

type GroupOutput struct {
	ID          int32            `json:"id"`
	Name        string           `json:"name" validate:"required"`
	WorkID      int32            `json:"work_id" validate:"required"`
	WorkTitle   string           `json:"work_title" validate:"required"`
	CreatedBy   int32            `json:"created_by" validate:"required"`
	Description string           `json:"description"`
	DepthLevel  string           `json:"depth_level" validate:"required"`
	Interests   []InterestOutput `json:"interests" validate:"required"`
}

func (s *GroupService) CreateGroup(ctx context.Context, input CreateGroupInput) (GroupOutput, error) {
	var groupID int32
	err := withTx(ctx, s.pool, func(q db.Querier) error {
		var err error
		groupID, err = q.CreateGroup(ctx, db.CreateGroupParams{
			Name:        input.Name,
			WorkID:      input.WorkID,
			CreatedBy:   input.CreatedBy,
			Description: &input.Description,
			DepthLevel:  input.DepthLevel,
		})
		if err != nil {
			return err
		}
		for _, id := range input.CategoriesIDs {
			err = q.AssignFocusTypeToGroup(ctx, db.AssignFocusTypeToGroupParams{
				GroupID:     groupID,
				FocusTypeID: id,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return GroupOutput{}, err
	}
	return s.GetGroup(ctx, groupID)
}

func (s *GroupService) GetGroup(ctx context.Context, groupID int32) (GroupOutput, error) {
	group, err := s.q.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupOutput{}, err
	}
	categories, err := s.q.GetGroupFocusTypes(ctx, groupID)
	if err != nil {
		return GroupOutput{}, err
	}

	categoriesOutput := make([]InterestOutput, 0, len(categories))
	for _, c := range categories {
		categoriesOutput = append(categoriesOutput, InterestOutput{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	return GroupOutput{
		ID:          group.ID,
		Name:        group.Name,
		WorkID:      group.WorkID,
		WorkTitle:   group.WorkTitle,
		CreatedBy:   group.CreatedBy,
		Description: *group.Description,
		DepthLevel:  group.DepthLevel,
		Interests:   categoriesOutput,
	}, nil
}
