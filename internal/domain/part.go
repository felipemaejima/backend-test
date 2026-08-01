package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MinCriticalityLevel = 1
	MaxCriticalityLevel = 5

	maxNameLength     = 120
	maxCategoryLength = 60
)

type Part struct {
	ID                uuid.UUID
	Name              string
	Category          string
	CurrentStock      int
	MinimumStock      int
	AverageDailySales float64
	LeadTimeDays      int
	UnitCost          float64
	CriticalityLevel  int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PartInput struct {
	Name              string
	Category          string
	CurrentStock      int
	MinimumStock      int
	AverageDailySales float64
	LeadTimeDays      int
	UnitCost          float64
	CriticalityLevel  int
}

func NewPart(in PartInput) (*Part, error) {
	now := time.Now().UTC()
	part := &Part{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}
	if err := part.Apply(in); err != nil {
		return nil, err
	}
	return part, nil
}

func (p *Part) Apply(in PartInput) error {
	candidate := *p
	candidate.Name = strings.TrimSpace(in.Name)
	candidate.Category = NormalizeCategory(in.Category)
	candidate.CurrentStock = in.CurrentStock
	candidate.MinimumStock = in.MinimumStock
	candidate.AverageDailySales = in.AverageDailySales
	candidate.LeadTimeDays = in.LeadTimeDays
	candidate.UnitCost = in.UnitCost
	candidate.CriticalityLevel = in.CriticalityLevel

	if err := candidate.Validate(); err != nil {
		return err
	}

	candidate.UpdatedAt = time.Now().UTC()
	*p = candidate
	return nil
}

func (p *Part) Validate() error {
	v := &ValidationError{}

	switch {
	case p.Name == "":
		v.add("name", "é obrigatório")
	case len([]rune(p.Name)) > maxNameLength:
		v.add("name", "excede o limite de 120 caracteres")
	}

	switch {
	case p.Category == "":
		v.add("category", "é obrigatória")
	case len([]rune(p.Category)) > maxCategoryLength:
		v.add("category", "excede o limite de 60 caracteres")
	}

	if p.MinimumStock < 0 {
		v.add("minimumStock", "não pode ser negativo")
	}
	if p.AverageDailySales < 0 {
		v.add("averageDailySales", "não pode ser negativa")
	}
	if p.LeadTimeDays < 0 {
		v.add("leadTimeDays", "não pode ser negativo")
	}
	if p.UnitCost < 0 {
		v.add("unitCost", "não pode ser negativo")
	}
	if p.CriticalityLevel < MinCriticalityLevel || p.CriticalityLevel > MaxCriticalityLevel {
		v.add("criticalityLevel", "deve estar entre 1 e 5")
	}

	return v.orNil()
}

func NormalizeCategory(category string) string {
	return strings.ToLower(strings.TrimSpace(category))
}
