package service

import (
	"context"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type RestockService struct {
	repo domain.PartRepository
}

func NewRestockService(repo domain.PartRepository) *RestockService {
	return &RestockService{repo: repo}
}

func (s *RestockService) Priorities(ctx context.Context) ([]domain.RestockPriority, error) {
	parts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return domain.CalculateRestockPriorities(parts), nil
}
