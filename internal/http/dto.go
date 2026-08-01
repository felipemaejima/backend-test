package http

import (
	"time"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type partRequest struct {
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	CurrentStock      int     `json:"currentStock"`
	MinimumStock      int     `json:"minimumStock"`
	AverageDailySales float64 `json:"averageDailySales"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	UnitCost          float64 `json:"unitCost"`
	CriticalityLevel  int     `json:"criticalityLevel"`
}

func (r partRequest) toInput() domain.PartInput {
	return domain.PartInput{
		Name:              r.Name,
		Category:          r.Category,
		CurrentStock:      r.CurrentStock,
		MinimumStock:      r.MinimumStock,
		AverageDailySales: r.AverageDailySales,
		LeadTimeDays:      r.LeadTimeDays,
		UnitCost:          r.UnitCost,
		CriticalityLevel:  r.CriticalityLevel,
	}
}

type partResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Category          string    `json:"category"`
	CurrentStock      int       `json:"currentStock"`
	MinimumStock      int       `json:"minimumStock"`
	AverageDailySales float64   `json:"averageDailySales"`
	LeadTimeDays      int       `json:"leadTimeDays"`
	UnitCost          float64   `json:"unitCost"`
	CriticalityLevel  int       `json:"criticalityLevel"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func newPartResponse(part domain.Part) partResponse {
	return partResponse{
		ID:                part.ID.String(),
		Name:              part.Name,
		Category:          part.Category,
		CurrentStock:      part.CurrentStock,
		MinimumStock:      part.MinimumStock,
		AverageDailySales: part.AverageDailySales,
		LeadTimeDays:      part.LeadTimeDays,
		UnitCost:          part.UnitCost,
		CriticalityLevel:  part.CriticalityLevel,
		CreatedAt:         part.CreatedAt,
		UpdatedAt:         part.UpdatedAt,
	}
}

type partListResponse struct {
	Parts []partResponse `json:"parts"`
}

func newPartListResponse(parts []domain.Part) partListResponse {
	items := make([]partResponse, 0, len(parts))
	for _, part := range parts {
		items = append(items, newPartResponse(part))
	}
	return partListResponse{Parts: items}
}

type restockPriorityResponse struct {
	PartID         string  `json:"partId"`
	Name           string  `json:"name"`
	CurrentStock   int     `json:"currentStock"`
	ProjectedStock float64 `json:"projectedStock"`
	MinimumStock   int     `json:"minimumStock"`
	UrgencyScore   float64 `json:"urgencyScore"`
}

type restockPrioritiesResponse struct {
	Priorities []restockPriorityResponse `json:"priorities"`
}

func newRestockPrioritiesResponse(priorities []domain.RestockPriority) restockPrioritiesResponse {
	items := make([]restockPriorityResponse, 0, len(priorities))
	for _, priority := range priorities {
		items = append(items, restockPriorityResponse{
			PartID:         priority.Part.ID.String(),
			Name:           priority.Part.Name,
			CurrentStock:   priority.Part.CurrentStock,
			ProjectedStock: priority.ProjectedStock,
			MinimumStock:   priority.Part.MinimumStock,
			UrgencyScore:   priority.UrgencyScore,
		})
	}
	return restockPrioritiesResponse{Priorities: items}
}
