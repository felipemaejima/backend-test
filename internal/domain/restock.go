package domain

import (
	"sort"
	"strings"
)

type RestockPriority struct {
	Part                Part
	ExpectedConsumption float64
	ProjectedStock      float64
	UrgencyScore        float64
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

	sort.SliceStable(priorities, func(i, j int) bool {
		return higherPriority(priorities[i], priorities[j])
	})

	return priorities
}

func higherPriority(a, b RestockPriority) bool {
	if a.UrgencyScore != b.UrgencyScore {
		return a.UrgencyScore > b.UrgencyScore
	}
	if a.Part.CriticalityLevel != b.Part.CriticalityLevel {
		return a.Part.CriticalityLevel > b.Part.CriticalityLevel
	}
	if a.Part.AverageDailySales != b.Part.AverageDailySales {
		return a.Part.AverageDailySales > b.Part.AverageDailySales
	}
	return strings.ToLower(a.Part.Name) < strings.ToLower(b.Part.Name)
}
