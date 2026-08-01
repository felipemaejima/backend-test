package memory_test

import (
	"context"
	"sync"
	"testing"

	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/repository/memory"
	"github.com/felipemaejima/backend-test/internal/repository/repositorytest"
)

func TestRepositoryContract(t *testing.T) {
	repositorytest.RunContract(t, func(t *testing.T) domain.PartRepository {
		return memory.NewPartRepository()
	})
}

func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewPartRepository()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(i int) {
			defer wg.Done()

			part, err := domain.NewPart(domain.PartInput{
				Name:              string(rune('A' + i%26)),
				Category:          "engine",
				CurrentStock:      i,
				MinimumStock:      20,
				AverageDailySales: 4,
				LeadTimeDays:      5,
				UnitCost:          18.50,
				CriticalityLevel:  3,
			})
			if err != nil {
				t.Errorf("NewPart: %v", err)
				return
			}

			if err := repo.Create(ctx, part); err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			if _, err := repo.FindByID(ctx, part.ID); err != nil {
				t.Errorf("FindByID: %v", err)
			}
			if _, err := repo.List(ctx, domain.PartFilter{}.Normalize()); err != nil {
				t.Errorf("List: %v", err)
			}
		}(i)
	}

	wg.Wait()

	parts, err := repo.List(ctx, domain.PartFilter{Limit: domain.MaxListLimit}.Normalize())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(parts) != goroutines {
		t.Errorf("expected %d parts, got %d", goroutines, len(parts))
	}
}
