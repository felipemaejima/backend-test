package domain

import (
	"cmp"
	"math"
	"slices"
	"strings"
)

const urgencyScoreEpsilon = 1e-9

type RestockPriority struct {
	Part                Part
	ExpectedConsumption float64
	ProjectedStock      float64
	UrgencyScore        float64

	sortName string
}

func (p Part) ExpectedConsumption() float64 {
	return p.AverageDailySales * float64(p.LeadTimeDays)
}

func (p Part) ProjectedStock() float64 {
	return float64(p.CurrentStock) - p.ExpectedConsumption()
}

func (p Part) NeedsRestock() bool {
	return p.ProjectedStock() < float64(p.MinimumStock)
}

func (p Part) UrgencyScore() float64 {
	return (float64(p.MinimumStock) - p.ProjectedStock()) * float64(p.CriticalityLevel)
}

func NewRestockPriority(part Part) RestockPriority {
	return RestockPriority{
		Part:                part,
		ExpectedConsumption: part.ExpectedConsumption(),
		ProjectedStock:      part.ProjectedStock(),
		UrgencyScore:        part.UrgencyScore(),
		sortName:            strings.ToLower(part.Name),
	}
}

func CalculateRestockPriorities(parts []Part) []RestockPriority {
	priorities := make([]RestockPriority, 0, len(parts))
	for _, part := range parts {
		if !part.NeedsRestock() {
			continue
		}
		priorities = append(priorities, NewRestockPriority(part))
	}

	slices.SortStableFunc(priorities, comparePriorities)

	return priorities
}

func comparePriorities(a, b RestockPriority) int {
	if math.Abs(a.UrgencyScore-b.UrgencyScore) > urgencyScoreEpsilon {
		return cmp.Compare(b.UrgencyScore, a.UrgencyScore)
	}
	if c := cmp.Compare(b.Part.CriticalityLevel, a.Part.CriticalityLevel); c != 0 {
		return c
	}
	if c := cmp.Compare(b.Part.AverageDailySales, a.Part.AverageDailySales); c != 0 {
		return c
	}
	return cmp.Compare(a.sortName, b.sortName)
}
