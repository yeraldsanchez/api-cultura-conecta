package service

import (
	"api-cultura-conecta/internal/apperrors"
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
	Name         string  `json:"name" validate:"required"`
}

type ProfileOutput struct {
	UserID     int32             `json:"user_id"`
	Name       string            `json:"name"`
	Email      string            `json:"email"`
	ProfileID  int32             `json:"profile_id"`
	DepthLevel string            `json:"depth_level"`
	FocusTypes []FocusTypeOutput `json:"focus_types"`
	Interests  []InterestOutput  `json:"interests"`
}

func (s *UserProfileService) Create(ctx context.Context, input CreateProfileInput) (ProfileOutput, error) {
	err := withTx(ctx, s.pool, func(q db.Querier) error {
		profileID, err := q.CreateUserProfile(ctx, db.CreateUserProfileParams{
			Name:       input.Name,
			UserID:     input.UserID,
			DepthLevel: input.DepthLevel,
		})
		if err != nil {
			return apperrors.FromPgx(err, apperrors.ProfilesConstraints)
		}
		for _, id := range input.FocusIDs {
			err = q.AssignFocusTypeToUser(ctx, db.AssignFocusTypeToUserParams{
				ProfileID:   profileID,
				FocusTypeID: id,
			})
			if err != nil {
				return apperrors.FromPgx(err, apperrors.UsersFocusTypesConstraints)
			}
		}
		for _, id := range input.InterestsIDs {
			err = q.AssignInterestToUser(ctx, db.AssignInterestToUserParams{
				ProfileID:  profileID,
				CategoryID: id,
			})
			if err != nil {
				return apperrors.FromPgx(err, apperrors.UserInterestsConstraints)
			}
		}
		return nil
	})
	if err != nil {
		return ProfileOutput{}, err
	}
	return s.GetProfile(ctx, input.UserID)
}

func (s *UserProfileService) GetProfile(ctx context.Context, userID int32) (ProfileOutput, error) {
	profile, err := s.q.GetUserProfileByUserId(ctx, userID)
	if err != nil {
		return ProfileOutput{}, apperrors.FromPgx(err, nil)
	}
	focusTypes, err := s.q.GetUserFocusTypes(ctx, profile.ProfileID)
	if err != nil {
		return ProfileOutput{}, err
	}
	focusTypesOutput := make([]FocusTypeOutput, len(focusTypes))
	for i, ft := range focusTypes {
		focusTypesOutput[i] = FocusTypeOutput{
			ID:   ft.ID,
			Name: ft.Name,
		}
	}

	userInterests, err := s.q.GetUserInterests(ctx, profile.ProfileID)
	if err != nil {
		return ProfileOutput{}, err
	}
	interestsOutput := make([]InterestOutput, len(userInterests))
	for i, interest := range userInterests {
		interestsOutput[i] = InterestOutput{
			ID:   interest.ID,
			Name: interest.Name,
		}
	}
	return ProfileOutput{
		UserID:     profile.UserID,
		Name:       profile.Name,
		Email:      profile.Email,
		ProfileID:  profile.ProfileID,
		DepthLevel: profile.DepthLevel,
		FocusTypes: focusTypesOutput,
		Interests:  interestsOutput,
	}, nil
}
