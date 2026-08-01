package repositorytest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type Factory func(t *testing.T) domain.PartRepository

func RunContract(t *testing.T, newRepo Factory) {
	t.Helper()

	t.Run("CreateAndFind", func(t *testing.T) { testCreateAndFind(t, newRepo(t)) })
	t.Run("FindNotFound", func(t *testing.T) { testFindNotFound(t, newRepo(t)) })
	t.Run("Update", func(t *testing.T) { testUpdate(t, newRepo(t)) })
	t.Run("UpdateNotFound", func(t *testing.T) { testUpdateNotFound(t, newRepo(t)) })
	t.Run("Delete", func(t *testing.T) { testDelete(t, newRepo(t)) })
	t.Run("DeleteNotFound", func(t *testing.T) { testDeleteNotFound(t, newRepo(t)) })
	t.Run("ListOrdersByName", func(t *testing.T) { testListOrder(t, newRepo(t)) })
	t.Run("ListFiltersByCategory", func(t *testing.T) { testListCategory(t, newRepo(t)) })
	t.Run("ListPaginates", func(t *testing.T) { testListPagination(t, newRepo(t)) })
	t.Run("ListOffsetBeyondEnd", func(t *testing.T) { testListOffsetBeyondEnd(t, newRepo(t)) })
}

func testCreateAndFind(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	part := newPart(t, "Filtro de Óleo X", "engine")

	if err := repo.Create(ctx, part); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, part.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	assertPartEqual(t, *found, *part)
}

func testFindNotFound(t *testing.T, repo domain.PartRepository) {
	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func testUpdate(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	part := newPart(t, "Filtro de Óleo X", "engine")
	if err := repo.Create(ctx, part); err != nil {
		t.Fatalf("Create: %v", err)
	}

	in := validInput("Filtro de Óleo Premium", "engine")
	in.CurrentStock = -7
	if err := part.Apply(in); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if err := repo.Update(ctx, part); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := repo.FindByID(ctx, part.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Filtro de Óleo Premium" {
		t.Errorf("Name = %q, expected the update", found.Name)
	}
	if found.CurrentStock != -7 {
		t.Errorf("CurrentStock = %d, expected -7", found.CurrentStock)
	}
}

func testUpdateNotFound(t *testing.T, repo domain.PartRepository) {
	part := newPart(t, "Peça Fantasma", "engine")
	err := repo.Update(context.Background(), part)
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func testDelete(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	part := newPart(t, "Filtro de Óleo X", "engine")
	if err := repo.Create(ctx, part); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, part.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, part.ID); !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound after removal", err)
	}
}

func testDeleteNotFound(t *testing.T, repo domain.PartRepository) {
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func testListOrder(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	seed(t, repo,
		testPart{"Pastilha de Freio Y", "brakes"},
		testPart{"Correia Dentada Z", "engine"},
		testPart{"Filtro de Óleo X", "engine"},
	)

	parts, err := repo.List(ctx, domain.PartFilter{}.Normalize())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	want := []string{"Correia Dentada Z", "Filtro de Óleo X", "Pastilha de Freio Y"}
	assertNames(t, parts, want)
}

func testListCategory(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	seed(t, repo,
		testPart{"Correia Dentada Z", "engine"},
		testPart{"Filtro de Óleo X", "engine"},
		testPart{"Pastilha de Freio Y", "brakes"},
	)

	parts, err := repo.List(ctx, domain.PartFilter{Category: "engine"}.Normalize())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	assertNames(t, parts, []string{"Correia Dentada Z", "Filtro de Óleo X"})
}

func testListPagination(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	seed(t, repo,
		testPart{"A", "engine"}, testPart{"B", "engine"}, testPart{"C", "engine"},
		testPart{"D", "engine"}, testPart{"E", "engine"},
	)

	parts, err := repo.List(ctx, domain.PartFilter{Limit: 2, Offset: 2}.Normalize())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	assertNames(t, parts, []string{"C", "D"})
}

func testListOffsetBeyondEnd(t *testing.T, repo domain.PartRepository) {
	ctx := context.Background()
	seed(t, repo, testPart{"A", "engine"})

	parts, err := repo.List(ctx, domain.PartFilter{Limit: 10, Offset: 99}.Normalize())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(parts) != 0 {
		t.Errorf("expected empty collection, got %d", len(parts))
	}
}

type testPart struct{ name, category string }

func seed(t *testing.T, repo domain.PartRepository, items ...testPart) {
	t.Helper()
	ctx := context.Background()
	for _, item := range items {
		if err := repo.Create(ctx, newPart(t, item.name, item.category)); err != nil {
			t.Fatalf("seeding %q: %v", item.name, err)
		}
	}
}

func validInput(name, category string) domain.PartInput {
	return domain.PartInput{
		Name:              name,
		Category:          category,
		CurrentStock:      15,
		MinimumStock:      20,
		AverageDailySales: 4,
		LeadTimeDays:      5,
		UnitCost:          18.50,
		CriticalityLevel:  3,
	}
}

func newPart(t *testing.T, name, category string) *domain.Part {
	t.Helper()
	created, err := domain.NewPart(validInput(name, category))
	if err != nil {
		t.Fatalf("NewPart(%q): %v", name, err)
	}
	return created
}

func assertNames(t *testing.T, parts []domain.Part, want []string) {
	t.Helper()
	if len(parts) != len(want) {
		t.Fatalf("expected %d part(s), got %d (%v)", len(want), len(parts), names(parts))
	}
	for i, expected := range want {
		if parts[i].Name != expected {
			t.Errorf("position %d = %q, expected %q (list: %v)", i, parts[i].Name, expected, names(parts))
		}
	}
}

func names(parts []domain.Part) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, part.Name)
	}
	return out
}

func assertPartEqual(t *testing.T, got, want domain.Part) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("ID = %v, expected %v", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, expected %q", got.Name, want.Name)
	}
	if got.Category != want.Category {
		t.Errorf("Category = %q, expected %q", got.Category, want.Category)
	}
	if got.CurrentStock != want.CurrentStock {
		t.Errorf("CurrentStock = %d, expected %d", got.CurrentStock, want.CurrentStock)
	}
	if got.MinimumStock != want.MinimumStock {
		t.Errorf("MinimumStock = %d, expected %d", got.MinimumStock, want.MinimumStock)
	}
	if got.AverageDailySales != want.AverageDailySales {
		t.Errorf("AverageDailySales = %v, expected %v", got.AverageDailySales, want.AverageDailySales)
	}
	if got.LeadTimeDays != want.LeadTimeDays {
		t.Errorf("LeadTimeDays = %d, expected %d", got.LeadTimeDays, want.LeadTimeDays)
	}
	if got.UnitCost != want.UnitCost {
		t.Errorf("UnitCost = %v, expected %v", got.UnitCost, want.UnitCost)
	}
	if got.CriticalityLevel != want.CriticalityLevel {
		t.Errorf("CriticalityLevel = %d, expected %d", got.CriticalityLevel, want.CriticalityLevel)
	}
	assertTimeClose(t, "CreatedAt", got.CreatedAt, want.CreatedAt)
	assertTimeClose(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt)
}

func assertTimeClose(t *testing.T, field string, got, want time.Time) {
	t.Helper()
	if diff := got.Sub(want).Abs(); diff > time.Second {
		t.Errorf("%s = %v, expected ~%v (diff of %v)", field, got, want, diff)
	}
}
