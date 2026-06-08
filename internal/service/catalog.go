package service

import (
	"api-cultura-conecta/internal/apperrors"
	db "api-cultura-conecta/internal/db/generated"
	"context"
)

type CatalogService struct {
	q db.Querier
}

func NewCatalogService(q db.Querier) *CatalogService {
	return &CatalogService{q: q}
}

type FocusTypeInput struct {
	Name string `json:"name" binding:"required"`
}
type FocusTypeOutput struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type InterestInput struct {
	Name string `json:"name" binding:"required"`
}

type InterestOutput struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (s *CatalogService) GetFocusTypes(ctx context.Context) ([]FocusTypeOutput, error) {
	focusTypes, err := s.q.GetFocusTypes(ctx)
	if err != nil {
		return nil, err
	}

	output := make([]FocusTypeOutput, 0, len(focusTypes))
	for _, ft := range focusTypes {
		output = append(output, FocusTypeOutput{
			ID:   ft.ID,
			Name: ft.Name,
		})
	}
	return output, nil
}

func (s *CatalogService) CreateFocusType(ctx context.Context, input FocusTypeInput) (FocusTypeOutput, error) {
	focusType, err := s.q.CreateFocusType(ctx, input.Name)
	if err != nil {
		return FocusTypeOutput{}, apperrors.FromPgx(err, apperrors.FocusTypesConstraints)
	}
	return FocusTypeOutput{
		ID:   focusType.ID,
		Name: focusType.Name,
	}, nil
}

func (s *CatalogService) GetInterests(ctx context.Context) ([]InterestOutput, error) {
	interest, err := s.q.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	output := make([]InterestOutput, 0, len(interest))
	for _, inter := range interest {
		output = append(output, InterestOutput{
			ID:   inter.ID,
			Name: inter.Name,
		})
	}
	return output, nil
}

func (s *CatalogService) CreateInterest(ctx context.Context, input InterestInput) (InterestOutput, error) {
	interest, err := s.q.CreateCategory(ctx, input.Name)
	if err != nil {
		return InterestOutput{}, apperrors.FromPgx(err, apperrors.CategoriesConstraints)
	}
	return InterestOutput{
		ID:   interest.ID,
		Name: interest.Name,
	}, nil
}
