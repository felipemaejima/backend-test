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

func allPages() domain.PageRequest {
	return domain.PageRequest{Number: 1, Size: domain.MaxPageSize}
}

func TestRestockPriorities(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()
	parts := service.NewPartService(repo)
	restock := service.NewRestockService(repo)

	seedPart(t, parts, "Healthy", 500, 20, 1, 2, 3)
	seedPart(t, parts, "Critical", -42, 20, 4, 5, 5)
	seedPart(t, parts, "Moderate", 8, 20, 4, 5, 4)

	page, err := restock.Priorities(ctx, allPages())
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("expected 2 priorities, got %d", len(page.Items))
	}
	// A peça saudável não entra na fila, então não conta no total.
	if page.Total != 2 {
		t.Errorf("Total = %d, expected 2", page.Total)
	}
	if page.Items[0].Part.Name != "Critical" {
		t.Errorf("first = %q, expected Critical", page.Items[0].Part.Name)
	}
	if page.Items[1].Part.Name != "Moderate" {
		t.Errorf("second = %q, expected Moderate", page.Items[1].Part.Name)
	}
	if page.Items[0].UrgencyScore != 410 {
		t.Errorf("UrgencyScore = %v, expected 410", page.Items[0].UrgencyScore)
	}
	if page.Items[0].ProjectedStock != -62 {
		t.Errorf("ProjectedStock = %v, expected -62", page.Items[0].ProjectedStock)
	}
}

func TestRestockPrioritiesPaginates(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()
	parts := service.NewPartService(repo)
	restock := service.NewRestockService(repo)

	// Criticidade crescente com o índice mantém a ordem por urgência previsível.
	for i := range 7 {
		seedPart(t, parts, fmt.Sprintf("Part %d", i), 0, 10, 1, 1, i%5+1)
	}

	page, err := restock.Priorities(ctx, domain.PageRequest{Number: 2, Size: 3})
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}

	if len(page.Items) != 3 {
		t.Fatalf("expected 3 items on page 2, got %d", len(page.Items))
	}
	if page.Total != 7 {
		t.Errorf("Total = %d, expected 7", page.Total)
	}
	if page.TotalPages() != 3 {
		t.Errorf("TotalPages = %d, expected 3", page.TotalPages())
	}
	if !page.HasNext() || !page.HasPrevious() {
		t.Error("expected both next and previous on page 2 of 3")
	}

	// A paginação recorta a lista já ordenada: nenhum score da página 2 pode
	// ser maior que os da página 1.
	first, err := restock.Priorities(ctx, domain.PageRequest{Number: 1, Size: 3})
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}
	lastOfFirst := first.Items[len(first.Items)-1].UrgencyScore
	if page.Items[0].UrgencyScore > lastOfFirst {
		t.Errorf("page 2 starts at %v, above the end of page 1 (%v)",
			page.Items[0].UrgencyScore, lastOfFirst)
	}
}

func TestRestockPrioritiesIgnoresRepositoryPageSize(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()
	parts := service.NewPartService(repo)
	restock := service.NewRestockService(repo)

	total := domain.MaxPageSize + 10
	for i := range total {
		seedPart(t, parts, fmt.Sprintf("Part %04d", i), 0, 10, 1, 1, 3)
	}

	page, err := restock.Priorities(ctx, allPages())
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}
	// O cálculo enxerga a base inteira, mesmo com a página limitada a 500.
	if page.Total != total {
		t.Errorf("Total = %d, expected %d", page.Total, total)
	}
	if len(page.Items) != domain.MaxPageSize {
		t.Errorf("expected %d items, got %d", domain.MaxPageSize, len(page.Items))
	}
}

func TestRestockPrioritiesEmptyRepository(t *testing.T) {
	restock := service.NewRestockService(memory.NewPartRepository())

	page, err := restock.Priorities(context.Background(), allPages())
	if err != nil {
		t.Fatalf("Priorities: %v", err)
	}
	if len(page.Items) != 0 {
		t.Errorf("expected 0 priorities, got %d", len(page.Items))
	}
	if page.Total != 0 || page.TotalPages() != 0 {
		t.Errorf("expected an empty page, got total=%d pages=%d", page.Total, page.TotalPages())
	}
}
