package service

import (
	db "api-cultura-conecta/internal/db/generated"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfileService struct {
	q    db.Querier
	pool *pgxpool.Pool
}

func NewUserProfileService(q db.Querier, pool *pgxpool.Pool) *UserProfileService {
	return &UserProfileService{q: q, pool: pool}
}

type CreateProfileInput struct {
	UserID       int32   `json:"user_id" validate:"required"`
	DepthLevel   string  `json:"depth_level" validate:"required"`
	FocusIDs     []int32 `json:"focus_ids" validate:"required"`
	InterestsIDs []int32 `json:"interests" validate:"required"`
}

type FocusTypesOutput struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type UserInterestsOutput struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}
type CreateProfileOutput struct {
	UserID     int32                 `json:"user_id"`
	Email      string                `json:"email"`
	ProfileID  int32                 `json:"profile_id"`
	DepthLevel string                `json:"depth_level"`
	FocusTypes []FocusTypesOutput    `json:"focus_types"`
	Interests  []UserInterestsOutput `json:"interests"`
}

func (s *UserProfileService) Create(ctx context.Context, input CreateProfileInput) (CreateProfileOutput, error) {
	err := withTx(ctx, s.pool, func(q db.Querier) error {
		profileID, err := q.CreateUserProfile(ctx, db.CreateUserProfileParams{
			UserID:     input.UserID,
			DepthLevel: input.DepthLevel,
		})
		if err != nil {
			return err
		}
		for _, id := range input.FocusIDs {
			err = q.AssignFocusTypeToUser(ctx, db.AssignFocusTypeToUserParams{
				ProfileID:   profileID,
				FocusTypeID: id,
			})
			if err != nil {
				return err
			}
		}
		for _, id := range input.InterestsIDs {
			err = q.AssignInterestToUser(ctx, db.AssignInterestToUserParams{
				ProfileID:  profileID,
				CategoryID: id,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return CreateProfileOutput{}, err
	}
	return s.GetProfile(ctx, input.UserID)
}

func (s *UserProfileService) GetProfile(ctx context.Context, userID int32) (CreateProfileOutput, error) {
	profile, err := s.q.GetUserProfileByUserId(ctx, userID)
	if err != nil {
		return CreateProfileOutput{}, err
	}
	focusTypes, err := s.q.GetUserFocusTypes(ctx, profile.ProfileID)
	if err != nil {
		return CreateProfileOutput{}, err
	}
	focusTypesOutput := make([]FocusTypesOutput, len(focusTypes))
	for i, ft := range focusTypes {
		focusTypesOutput[i] = FocusTypesOutput{
			ID:   ft.ID,
			Name: ft.Name,
		}
	}

	userInterests, err := s.q.GetUserInterests(ctx, profile.ProfileID)
	if err != nil {
		return CreateProfileOutput{}, err
	}
	interestsOutput := make([]UserInterestsOutput, len(userInterests))
	for i, interest := range userInterests {
		interestsOutput[i] = UserInterestsOutput{
			ID:   interest.ID,
			Name: interest.Name,
		}
	}
	return CreateProfileOutput{
		UserID:     profile.UserID,
		Email:      profile.Email,
		ProfileID:  profile.ProfileID,
		DepthLevel: profile.DepthLevel,
		FocusTypes: focusTypesOutput,
		Interests:  interestsOutput,
	}, nil
}
