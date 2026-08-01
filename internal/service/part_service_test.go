package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/repository/memory"
	"github.com/felipemaejima/backend-test/internal/service"
)

func newService() *service.PartService {
	return service.NewPartService(memory.NewPartRepository())
}

func validInput() domain.PartInput {
	return domain.PartInput{
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

func TestCreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	created, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.ID != created.ID || found.Name != created.Name {
		t.Errorf("retrieved part = %+v, expected %+v", found, created)
	}
}

func TestCreateInvalidDoesNotPersist(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	in := validInput()
	in.CriticalityLevel = 99

	if _, err := svc.Create(ctx, in); err == nil {
		t.Fatal("expected validation error")
	}

	parts, err := svc.List(ctx, domain.PartFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(parts) != 0 {
		t.Errorf("expected empty repository, got %d part(s)", len(parts))
	}
}

func TestGetByIDNotFound(t *testing.T) {
	_, err := newService().GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	created, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	in := validInput()
	in.Name = "Filtro de Óleo Premium"
	in.CurrentStock = 3

	updated, err := svc.Update(ctx, created.ID, in)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.ID != created.ID {
		t.Error("Update should not change the ID")
	}
	if updated.Name != "Filtro de Óleo Premium" || updated.CurrentStock != 3 {
		t.Errorf("part = %+v, expected the updated fields", updated)
	}

	found, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.Name != "Filtro de Óleo Premium" {
		t.Errorf("persisted part = %q, expected the update", found.Name)
	}
}

func TestUpdateNotFound(t *testing.T) {
	_, err := newService().Update(context.Background(), uuid.New(), validInput())
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func TestUpdateInvalidDoesNotMutate(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	created, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	in := validInput()
	in.Name = ""

	if _, err := svc.Update(ctx, created.ID, in); err == nil {
		t.Fatal("expected validation error")
	}

	found, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.Name != created.Name {
		t.Errorf("part = %q, expected it untouched (%q)", found.Name, created.Name)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	created, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.GetByID(ctx, created.ID); !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound after removal", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	err := newService().Delete(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPartNotFound) {
		t.Fatalf("error = %v, expected ErrPartNotFound", err)
	}
}

func TestListFilterByCategory(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	seed := []struct{ name, category string }{
		{"Filtro de Óleo X", "engine"},
		{"Correia Dentada Z", "engine"},
		{"Pastilha de Freio Y", "brakes"},
	}
	for _, item := range seed {
		in := validInput()
		in.Name = item.name
		in.Category = item.category
		if _, err := svc.Create(ctx, in); err != nil {
			t.Fatalf("Create(%s): %v", item.name, err)
		}
	}

	parts, err := svc.List(ctx, domain.PartFilter{Category: "ENGINE"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 engine parts, got %d", len(parts))
	}
	if parts[0].Name != "Correia Dentada Z" {
		t.Errorf("first part = %q, expected name ordering", parts[0].Name)
	}

	all, err := svc.List(ctx, domain.PartFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 parts without filter, got %d", len(all))
	}
}

func TestListPagination(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	for _, name := range []string{"A", "B", "C", "D", "E"} {
		in := validInput()
		in.Name = name
		if _, err := svc.Create(ctx, in); err != nil {
			t.Fatalf("Create(%s): %v", name, err)
		}
	}

	page, err := svc.List(ctx, domain.PartFilter{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("expected 2 parts in the page, got %d", len(page))
	}
	if page[0].Name != "C" || page[1].Name != "D" {
		t.Errorf("page = [%s %s], expected [C D]", page[0].Name, page[1].Name)
	}

	empty, err := svc.List(ctx, domain.PartFilter{Limit: 10, Offset: 99})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty page, got %d", len(empty))
	}
}
