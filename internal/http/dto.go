package http

import (
	"math"
	"time"

	"github.com/felipemaejima/backend-test/internal/domain"
)

const responseDecimalFactor = 1e4

func round(value float64) float64 {
	return math.Round(value*responseDecimalFactor) / responseDecimalFactor
}

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

type paginationResponse struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"perPage"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

func newPaginationResponse[T any](page domain.Page[T]) paginationResponse {
	return paginationResponse{
		Page:        page.Number,
		PerPage:     page.Size,
		Total:       page.Total,
		TotalPages:  page.TotalPages(),
		HasNext:     page.HasNext(),
		HasPrevious: page.HasPrevious(),
	}
}

type partListResponse struct {
	Parts      []partResponse     `json:"parts"`
	Pagination paginationResponse `json:"pagination"`
}

func newPartListResponse(page domain.Page[domain.Part]) partListResponse {
	items := make([]partResponse, 0, len(page.Items))
	for _, part := range page.Items {
		items = append(items, newPartResponse(part))
	}
	return partListResponse{Parts: items, Pagination: newPaginationResponse(page)}
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
	Pagination paginationResponse        `json:"pagination"`
}

func newRestockPrioritiesResponse(page domain.Page[domain.RestockPriority]) restockPrioritiesResponse {
	items := make([]restockPriorityResponse, 0, len(page.Items))
	for _, priority := range page.Items {
		items = append(items, restockPriorityResponse{
			PartID:         priority.Part.ID.String(),
			Name:           priority.Part.Name,
			CurrentStock:   priority.Part.CurrentStock,
			ProjectedStock: round(priority.ProjectedStock),
			MinimumStock:   priority.Part.MinimumStock,
			UrgencyScore:   round(priority.UrgencyScore),
		})
	}
	return restockPrioritiesResponse{Priorities: items, Pagination: newPaginationResponse(page)}
}
