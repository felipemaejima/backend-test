package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type PartService struct {
	repo domain.PartRepository
}

func NewPartService(repo domain.PartRepository) *PartService {
	return &PartService{repo: repo}
}

func (s *PartService) Create(ctx context.Context, in domain.PartInput) (*domain.Part, error) {
	part, err := domain.NewPart(in)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

func (s *PartService) Update(ctx context.Context, id uuid.UUID, in domain.PartInput) (*domain.Part, error) {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := part.Apply(in); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

func (s *PartService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *PartService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PartService) List(ctx context.Context, filter domain.PartFilter) ([]domain.Part, error) {
	return s.repo.List(ctx, filter.Normalize())
}
