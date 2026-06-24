package service

import (
	"api-cultura-conecta/internal/apperrors"
	db "api-cultura-conecta/internal/db/generated"
	"context"
	"encoding/json"
	"time"

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
	Description *string          `json:"description"`
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
			return apperrors.FromPgx(err, apperrors.GroupsConstraints)
		}
		for _, id := range input.CategoriesIDs {
			err = q.AssignFocusTypeToGroup(ctx, db.AssignFocusTypeToGroupParams{
				GroupID:     groupID,
				FocusTypeID: id,
			})
			if err != nil {
				return apperrors.FromPgx(err, apperrors.GroupsFocusTypesConstraints)
			}
		}
		err = q.AddGroupMember(ctx, db.AddGroupMemberParams{
			GroupID: groupID,
			UserID:  input.CreatedBy,
			Role:    "admin",
		})
		if err != nil {
			return apperrors.FromPgx(err, apperrors.GroupMembersConstraints)
		}
		return nil
	})
	if err != nil {
		return GroupOutput{}, err
	}
	return s.GetGroup(ctx, groupID)
}

type ListGroupsInput struct {
	WorkID       *int64  `json:"work_id"`
	Name         *string `json:"name"`
	DepthLevel   *string `json:"depth_level"`
	FocusTypeIDs []int32 `json:"focus_type_ids"`
	Limit        int32   `json:"limit"`
	Offset       int32   `json:"offset"`
}

type ListGroupsOutput struct {
	Groups []GroupOutput
	Total  int64
}

func (s *GroupService) ListGroups(ctx context.Context, input ListGroupsInput) (ListGroupsOutput, error) {
	countParams := db.CountGroupsParams{
		WorkID:       input.WorkID,
		Name:         input.Name,
		DepthLevel:   input.DepthLevel,
		FocusTypeIds: input.FocusTypeIDs,
	}
	total, err := s.q.CountGroups(ctx, countParams)
	if err != nil {
		return ListGroupsOutput{}, err
	}

	rows, err := s.q.ListGroups(ctx, db.ListGroupsParams{
		WorkID:       input.WorkID,
		Name:         input.Name,
		DepthLevel:   input.DepthLevel,
		FocusTypeIds: input.FocusTypeIDs,
		Limit:        input.Limit,
		Offset:       input.Offset,
	})
	if err != nil {
		return ListGroupsOutput{}, err
	}

	groups := make([]GroupOutput, 0, len(rows))
	for _, row := range rows {
		var interests []InterestOutput
		if len(row.FocusTypes) > 0 {
			if err := json.Unmarshal(row.FocusTypes, &interests); err != nil {
				return ListGroupsOutput{}, err
			}
		}
		groups = append(groups, GroupOutput{
			ID:          row.ID,
			Name:        row.Name,
			WorkID:      row.WorkID,
			WorkTitle:   row.WorkTitle,
			CreatedBy:   row.CreatedBy,
			Description: row.Description,
			DepthLevel:  row.DepthLevel,
			Interests:   interests,
		})
	}
	return ListGroupsOutput{Groups: groups, Total: total}, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, groupID int32, userID int32) error {
	return withTx(ctx, s.pool, func(q db.Querier) error {
		err := q.AddGroupMember(ctx, db.AddGroupMemberParams{
			GroupID: groupID,
			UserID:  userID,
			Role:    "member",
		})
		return apperrors.FromPgx(err, apperrors.GroupMembersConstraints)
	})
}

type CreatePostInput struct {
	GroupID         int32
	UserID          int32
	Content         string
	HasSpoiler      bool
	SpoilerProgress *string
}

type PostOutput struct {
	ID              int32     `json:"id"`
	GroupID         int32     `json:"group_id"`
	UserID          int32     `json:"user_id"`
	Content         string    `json:"content"`
	HasSpoiler      bool      `json:"has_spoiler"`
	SpoilerProgress *string   `json:"spoiler_progress"`
	CreatedAt       time.Time `json:"created_at"`
}

func (s *GroupService) CreatePost(ctx context.Context, input CreatePostInput) (PostOutput, error) {
	var out PostOutput
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

		post, err := q.CreatePost(ctx, db.CreatePostParams{
			GroupID:         input.GroupID,
			UserID:          input.UserID,
			Content:         input.Content,
			HasSpoiler:      input.HasSpoiler,
			SpoilerProgress: input.SpoilerProgress,
		})
		if err != nil {
			return apperrors.FromPgx(err, apperrors.PostsConstraints)
		}

		out = PostOutput{
			ID:              post.ID,
			GroupID:         post.GroupID,
			UserID:          post.UserID,
			Content:         post.Content,
			HasSpoiler:      post.HasSpoiler,
			SpoilerProgress: post.SpoilerProgress,
			CreatedAt:       post.CreatedAt,
		}
		return nil
	})
	return out, err
}

type ListGroupPostsInput struct {
	GroupID int32
	UserID  int32
	Limit   int32
	Offset  int32
}

type PostWithAuthorOutput struct {
	ID              int32     `json:"id"`
	GroupID         int32     `json:"group_id"`
	UserID          int32     `json:"user_id"`
	AuthorName      *string   `json:"author_name"`
	Content         string    `json:"content"`
	HasSpoiler      bool      `json:"has_spoiler"`
	SpoilerProgress *string   `json:"spoiler_progress"`
	CreatedAt       time.Time `json:"created_at"`
}

type ListGroupPostsOutput struct {
	Posts []PostWithAuthorOutput
}

func (s *GroupService) ListGroupPosts(ctx context.Context, input ListGroupPostsInput) (ListGroupPostsOutput, error) {
	isMember, err := s.q.IsGroupMember(ctx, db.IsGroupMemberParams{
		GroupID: input.GroupID,
		UserID:  input.UserID,
	})
	if err != nil {
		return ListGroupPostsOutput{}, err
	}
	if !isMember {
		return ListGroupPostsOutput{}, apperrors.ErrNotGroupMember
	}

	rows, err := s.q.ListGroupPosts(ctx, db.ListGroupPostsParams{
		GroupID: input.GroupID,
		Offset:  input.Offset,
		Limit:   input.Limit,
	})
	if err != nil {
		return ListGroupPostsOutput{}, err
	}

	posts := make([]PostWithAuthorOutput, len(rows))
	for i, r := range rows {
		posts[i] = PostWithAuthorOutput{
			ID:              r.ID,
			GroupID:         r.GroupID,
			UserID:          r.UserID,
			AuthorName:      r.AuthorName,
			Content:         r.Content,
			HasSpoiler:      r.HasSpoiler,
			SpoilerProgress: r.SpoilerProgress,
			CreatedAt:       r.CreatedAt,
		}
	}
	return ListGroupPostsOutput{Posts: posts}, nil
}

type SuggestGroupsInput struct {
	UserID int32
	Limit  int32
	Offset int32
}

func (s *GroupService) GetSuggestedGroups(ctx context.Context, input SuggestGroupsInput) (ListGroupsOutput, error) {
	total, err := s.q.CountSuggestedGroups(ctx, input.UserID)
	if err != nil {
		return ListGroupsOutput{}, err
	}

	rows, err := s.q.ListSuggestedGroups(ctx, db.ListSuggestedGroupsParams{
		UserID: input.UserID,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return ListGroupsOutput{}, err
	}

	groups := make([]GroupOutput, 0, len(rows))
	for _, row := range rows {
		var interests []InterestOutput
		if len(row.FocusTypes) > 0 {
			if err := json.Unmarshal(row.FocusTypes, &interests); err != nil {
				return ListGroupsOutput{}, err
			}
		}
		groups = append(groups, GroupOutput{
			ID:          row.ID,
			Name:        row.Name,
			WorkID:      row.WorkID,
			WorkTitle:   row.WorkTitle,
			CreatedBy:   row.CreatedBy,
			Description: row.Description,
			DepthLevel:  row.DepthLevel,
			Interests:   interests,
		})
	}
	return ListGroupsOutput{Groups: groups, Total: total}, nil
}

type UserGroupOutput struct {
	ID          int32            `json:"id"`
	Name        string           `json:"name"`
	WorkID      int32            `json:"work_id"`
	WorkTitle   string           `json:"work_title"`
	CreatedBy   int32            `json:"created_by"`
	Description *string          `json:"description"`
	DepthLevel  string           `json:"depth_level"`
	Role        string           `json:"role"`
	JoinedAt    time.Time        `json:"joined_at"`
	Interests   []InterestOutput `json:"interests"`
}

func (s *GroupService) GetGroupsByMember(ctx context.Context, userID int32) ([]UserGroupOutput, error) {
	rows, err := s.q.ListGroupsByMember(ctx, userID)
	if err != nil {
		return nil, err
	}

	groups := make([]UserGroupOutput, 0, len(rows))
	for _, row := range rows {
		var interests []InterestOutput
		if len(row.FocusTypes) > 0 {
			if err := json.Unmarshal(row.FocusTypes, &interests); err != nil {
				return nil, err
			}
		}
		groups = append(groups, UserGroupOutput{
			ID:          row.ID,
			Name:        row.Name,
			WorkID:      row.WorkID,
			WorkTitle:   row.WorkTitle,
			CreatedBy:   row.CreatedBy,
			Description: row.Description,
			DepthLevel:  row.DepthLevel,
			Role:        row.Role,
			JoinedAt:    row.JoinedAt,
			Interests:   interests,
		})
	}
	return groups, nil
}

type GroupMemberOutput struct {
	UserID   int32     `json:"user_id"`
	Name     *string   `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupID int32) ([]GroupMemberOutput, error) {
	rows, err := s.q.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, apperrors.FromPgx(err, nil)
	}

	members := make([]GroupMemberOutput, 0, len(rows))
	for _, row := range rows {
		members = append(members, GroupMemberOutput{
			UserID:   row.UserID,
			Name:     row.Name,
			Role:     row.Role,
			JoinedAt: row.JoinedAt,
		})
	}
	return members, nil
}

func (s *GroupService) GetGroup(ctx context.Context, groupID int32) (GroupOutput, error) {
	group, err := s.q.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupOutput{}, apperrors.FromPgx(err, nil)
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
		Description: group.Description,
		DepthLevel:  group.DepthLevel,
		Interests:   categoriesOutput,
	}, nil
}
