package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/repository/memory"
	"github.com/felipemaejima/backend-test/internal/service"
)

func seedPart(t *testing.T, svc *service.PartService, name string, currentStock, minimumStock int, averageDailySales float64, leadTimeDays, criticalityLevel int) {
	t.Helper()

	in := domain.PartInput{
		Name:              name,
		Category:          "engine",
		CurrentStock:      currentStock,
		MinimumStock:      minimumStock,
		AverageDailySales: averageDailySales,
		LeadTimeDays:      leadTimeDays,
		UnitCost:          18.50,
		CriticalityLevel:  criticalityLevel,
	}
	if _, err := svc.Create(context.Background(), in); err != nil {
		t.Fatalf("Create(%s): %v", name, err)
	}
}

func TestRestockPriorities(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()
	parts := service.NewPartService(repo)
	restock := service.NewRestockService(repo)

	seedPart(t, parts, "Healthy", 500, 20, 1, 2, 3)
	seedPart(t, parts, "Critical", -42, 20, 4, 5, 5)
	seedPart(t, parts, "Moderate", 8, 20, 4, 5, 4)

	priorities, err := restock.Priorities(ctx)
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}

	if len(priorities) != 2 {
		t.Fatalf("expected 2 priorities, got %d", len(priorities))
	}
	if priorities[0].Part.Name != "Critical" {
		t.Errorf("first = %q, expected Critical", priorities[0].Part.Name)
	}
	if priorities[1].Part.Name != "Moderate" {
		t.Errorf("second = %q, expected Moderate", priorities[1].Part.Name)
	}
	if priorities[0].UrgencyScore != 410 {
		t.Errorf("UrgencyScore = %v, expected 410", priorities[0].UrgencyScore)
	}
	if priorities[0].ProjectedStock != -62 {
		t.Errorf("ProjectedStock = %v, expected -62", priorities[0].ProjectedStock)
	}
}

func TestRestockPrioritiesIgnoresPaginationLimits(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()
	parts := service.NewPartService(repo)
	restock := service.NewRestockService(repo)

	total := domain.MaxListLimit + 10
	for i := range total {
		seedPart(t, parts, fmt.Sprintf("Part %04d", i), 0, 10, 1, 1, 3)
	}

	priorities, err := restock.Priorities(ctx)
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}
	if len(priorities) != total {
		t.Errorf("expected %d priorities, got %d", total, len(priorities))
	}
}

func TestRestockPrioritiesEmptyRepository(t *testing.T) {
	restock := service.NewRestockService(memory.NewPartRepository())

	priorities, err := restock.Priorities(context.Background())
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}
	if len(priorities) != 0 {
		t.Errorf("expected 0 priorities, got %d", len(priorities))
	}
}
