package domain

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func validInput() PartInput {
	return PartInput{
		Name:              "Filtro de Óleo X",
		Category:          "engine",
		CurrentStock:      15,
		MinimumStock:      20,
		AverageDailySales: 4,
		LeadTimeDays:      5,
		UnitCost:          18.50,
		CriticalityLevel:  3,
	}
}

func invalidFields(t *testing.T, err error) []string {
	t.Helper()

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}

	names := make([]string, 0, len(validationErr.Fields))
	for _, field := range validationErr.Fields {
		names = append(names, field.Field)
	}
	slices.Sort(names)
	return names
}

func TestNewPartValid(t *testing.T) {
	part, err := NewPart(validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if part.ID == uuid.Nil {
		t.Error("expected a generated ID")
	}
	if part.CreatedAt.IsZero() || part.UpdatedAt.IsZero() {
		t.Error("expected timestamps to be set")
	}
	if part.Name != "Filtro de Óleo X" {
		t.Errorf("Name = %q", part.Name)
	}
}

func TestNewPartNormalizesInput(t *testing.T) {
	in := validInput()
	in.Name = "  Pastilha de Freio Y  "
	in.Category = "  BRAKES  "

	part, err := NewPart(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if part.Name != "Pastilha de Freio Y" {
		t.Errorf("Name = %q, expected trimmed", part.Name)
	}
	if part.Category != "brakes" {
		t.Errorf("Category = %q, expected %q", part.Category, "brakes")
	}
}

func TestNewPartAcceptsNegativeStock(t *testing.T) {
	in := validInput()
	in.CurrentStock = -30

	part, err := NewPart(in)
	if err != nil {
		t.Fatalf("negative stock should be accepted, got error: %v", err)
	}
	if part.CurrentStock != -30 {
		t.Errorf("CurrentStock = %d, expected -30", part.CurrentStock)
	}
}

func TestNewPartRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*PartInput)
		wantFields []string
	}{
		{
			name:       "empty name",
			mutate:     func(in *PartInput) { in.Name = "   " },
			wantFields: []string{"name"},
		},
		{
			name:       "empty category",
			mutate:     func(in *PartInput) { in.Category = "" },
			wantFields: []string{"category"},
		},
		{
			name:       "negative minimum stock",
			mutate:     func(in *PartInput) { in.MinimumStock = -1 },
			wantFields: []string{"minimumStock"},
		},
		{
			name:       "negative average daily sales",
			mutate:     func(in *PartInput) { in.AverageDailySales = -0.5 },
			wantFields: []string{"averageDailySales"},
		},
		{
			name:       "negative lead time",
			mutate:     func(in *PartInput) { in.LeadTimeDays = -3 },
			wantFields: []string{"leadTimeDays"},
		},
		{
			name:       "negative unit cost",
			mutate:     func(in *PartInput) { in.UnitCost = -1 },
			wantFields: []string{"unitCost"},
		},
		{
			name:       "criticality below range",
			mutate:     func(in *PartInput) { in.CriticalityLevel = 0 },
			wantFields: []string{"criticalityLevel"},
		},
		{
			name:       "criticality above range",
			mutate:     func(in *PartInput) { in.CriticalityLevel = 6 },
			wantFields: []string{"criticalityLevel"},
		},
		{
			name: "multiple violations accumulate",
			mutate: func(in *PartInput) {
				in.Name = ""
				in.CriticalityLevel = 9
				in.LeadTimeDays = -1
			},
			wantFields: []string{"criticalityLevel", "leadTimeDays", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.mutate(&in)

			part, err := NewPart(in)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if part != nil {
				t.Error("expected nil part when validation fails")
			}

			got := invalidFields(t, err)
			if !slices.Equal(got, tt.wantFields) {
				t.Errorf("fields = %v, expected %v", got, tt.wantFields)
			}
		})
	}
}

func TestApplyDoesNotMutateOnInvalidInput(t *testing.T) {
	part, err := NewPart(validInput())
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	original := *part

	in := validInput()
	in.Name = ""

	if err := part.Apply(in); err == nil {
		t.Fatal("expected validation error")
	}

	if *part != original {
		t.Errorf("part was modified despite the error:\n  before: %+v\n   after: %+v", original, *part)
	}
}

func TestApplyPreservesIdentity(t *testing.T) {
	part, err := NewPart(validInput())
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	id, createdAt := part.ID, part.CreatedAt

	in := validInput()
	in.Name = "Correia Dentada Z"
	if err := part.Apply(in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if part.ID != id {
		t.Error("Apply should not change the ID")
	}
	if !part.CreatedAt.Equal(createdAt) {
		t.Error("Apply should not change CreatedAt")
	}
	if part.Name != "Correia Dentada Z" {
		t.Errorf("Name = %q, expected the update", part.Name)
	}
}

func TestPartFilterNormalizePagination(t *testing.T) {
	tests := []struct {
		name       string
		filter     PartFilter
		wantLimit  int
		wantOffset int
	}{
		{"missing limit falls back to default", PartFilter{}, DefaultListLimit, 0},
		{"negative limit falls back to default", PartFilter{Limit: -5}, DefaultListLimit, 0},
		{"limit above cap is clamped", PartFilter{Limit: 10_000}, MaxListLimit, 0},
		{"valid limit is preserved", PartFilter{Limit: 20, Offset: 40}, 20, 40},
		{"negative offset becomes zero", PartFilter{Limit: 10, Offset: -1}, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Normalize()
			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, expected %d", got.Limit, tt.wantLimit)
			}
			if got.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, expected %d", got.Offset, tt.wantOffset)
			}
		})
	}
}

func TestPartFilterNormalizeCategory(t *testing.T) {
	got := PartFilter{Category: "  ENGINE  "}.Normalize()
	if got.Category != "engine" {
		t.Errorf("Category = %q, expected %q", got.Category, "engine")
	}
}
