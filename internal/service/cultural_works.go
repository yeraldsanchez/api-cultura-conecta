package service

import (
	db "api-cultura-conecta/internal/db/generated"
	"context"
)

type CulturalWorksService struct {
	q db.Querier
}

func NewCulturalWorksService(q db.Querier) *CulturalWorksService {
	return &CulturalWorksService{q: q}
}

type CreateCulturalWorkInput struct {
	Title      string `json:"title" binding:"required"`
	CategoryID int32  `json:"category_id" binding:"required"`
}

type CreateCulturalWorkOutput struct {
	Id           int32  `json:"id"`
	Title        string `json:"title"`
	CategoryID   int32  `json:"category_id"`
	CategoryName string `json:"category_name"`
	CreatedAt    string `json:"created_at"`
}

func (s *CulturalWorksService) CreateCulturalWork(ctx context.Context, input CreateCulturalWorkInput) (CreateCulturalWorkOutput, error) {
	culturalWork, err := s.q.CreateCulturalWork(ctx, db.CreateCulturalWorkParams{
		Title: input.Title,
		ID:    input.CategoryID,
	})
	if err != nil {
		return CreateCulturalWorkOutput{}, err
	}
	return CreateCulturalWorkOutput{
		Id:           culturalWork.ID,
		Title:        culturalWork.Title,
		CategoryID:   culturalWork.CategoryID,
		CategoryName: culturalWork.CategoryName,
		CreatedAt:    culturalWork.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *CulturalWorksService) GetCulturalWorks(ctx context.Context) ([]CreateCulturalWorkOutput, error) {
	culturalWorks, err := s.q.GetCulturalWorks(ctx)
	if err != nil {
		return []CreateCulturalWorkOutput{}, err
	}

	output := make([]CreateCulturalWorkOutput, 0, len(culturalWorks))
	for _, culturalWork := range culturalWorks {
		output = append(output, CreateCulturalWorkOutput{
			Id:           culturalWork.ID,
			Title:        culturalWork.Title,
			CategoryID:   culturalWork.CategoryID,
			CategoryName: culturalWork.CategoryName,
			CreatedAt:    culturalWork.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return output, nil
}
