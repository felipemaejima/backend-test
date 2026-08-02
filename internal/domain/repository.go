package domain

import (
	"context"

	"github.com/google/uuid"
)

type PartFilter struct {
	Category string
	Page     PageRequest
}

func (f PartFilter) Normalize() PartFilter {
	f.Category = NormalizeCategory(f.Category)
	f.Page = f.Page.Normalize()
	return f
}

type PartRepository interface {
	Create(ctx context.Context, part *Part) error
	Update(ctx context.Context, part *Part) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Part, error)
	List(ctx context.Context, filter PartFilter) (Page[Part], error)
	ListAll(ctx context.Context) ([]Part, error)
}
