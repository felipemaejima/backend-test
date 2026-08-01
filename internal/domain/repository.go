package domain

import (
	"context"

	"github.com/google/uuid"
)

const (
	DefaultListLimit = 50
	MaxListLimit     = 500
)

type PartFilter struct {
	Category string
	Limit    int
	Offset   int
}

func (f PartFilter) Normalize() PartFilter {
	f.Category = NormalizeCategory(f.Category)
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
}

type PartRepository interface {
	Create(ctx context.Context, part *Part) error
	Update(ctx context.Context, part *Part) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Part, error)
	List(ctx context.Context, filter PartFilter) ([]Part, error)
}
