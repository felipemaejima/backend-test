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

func (s *RestockService) Priorities(ctx context.Context, request domain.PageRequest) (domain.Page[domain.RestockPriority], error) {
	parts, err := s.repo.ListAll(ctx)
	if err != nil {
		return domain.Page[domain.RestockPriority]{}, err
	}

	return domain.Paginate(domain.CalculateRestockPriorities(parts), request), nil
}
