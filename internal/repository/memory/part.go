package memory

import (
	"bytes"
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type PartRepository struct {
	mu    sync.RWMutex
	parts map[uuid.UUID]domain.Part
}

func NewPartRepository() *PartRepository {
	return &PartRepository{parts: make(map[uuid.UUID]domain.Part)}
}

var _ domain.PartRepository = (*PartRepository)(nil)

func (r *PartRepository) Create(ctx context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parts[part.ID] = *part
	return nil
}

func (r *PartRepository) Update(ctx context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.parts[part.ID]; !ok {
		return domain.ErrPartNotFound
	}
	r.parts[part.ID] = *part
	return nil
}

func (r *PartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.parts[id]; !ok {
		return domain.ErrPartNotFound
	}
	delete(r.parts, id)
	return nil
}

func (r *PartRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	part, ok := r.parts[id]
	if !ok {
		return nil, domain.ErrPartNotFound
	}
	return &part, nil
}

func (r *PartRepository) List(ctx context.Context, filter domain.PartFilter) ([]domain.Part, error) {
	filter = filter.Normalize()

	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]domain.Part, 0, len(r.parts))
	for _, part := range r.parts {
		if filter.Category != "" && part.Category != filter.Category {
			continue
		}
		filtered = append(filtered, part)
	}
	sortByName(filtered)

	if filter.Offset >= len(filtered) {
		return []domain.Part{}, nil
	}
	end := min(filter.Offset+filter.Limit, len(filtered))
	return filtered[filter.Offset:end], nil
}

func (r *PartRepository) ListAll(ctx context.Context) ([]domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parts := make([]domain.Part, 0, len(r.parts))
	for _, part := range r.parts {
		parts = append(parts, part)
	}
	sortByName(parts)
	return parts, nil
}

func sortByName(parts []domain.Part) {
	slices.SortFunc(parts, func(a, b domain.Part) int {
		if c := cmp.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return bytes.Compare(a.ID[:], b.ID[:])
	})
}
