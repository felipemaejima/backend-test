package http_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func createPart(t *testing.T, app *fiber.App, name, category string) string {
	t.Helper()

	body := fmt.Sprintf(`{
		"name": %q,
		"category": %q,
		"currentStock": 15,
		"minimumStock": 20,
		"averageDailySales": 4,
		"leadTimeDays": 5,
		"unitCost": 18.50,
		"criticalityLevel": 3
	}`, name, category)

	status, decoded := do(t, app, http.MethodPost, "/parts", body)
	if status != http.StatusCreated {
		t.Fatalf("creating %q: status = %d (%v)", name, status, decoded)
	}

	id, _ := decoded["id"].(string)
	if id == "" {
		t.Fatalf("creating %q: response without id", name)
	}
	return id
}

func partNames(t *testing.T, body map[string]any) []string {
	t.Helper()

	raw, ok := body["parts"].([]any)
	if !ok {
		t.Fatalf("parts = %v, expected an array", body["parts"])
	}

	names := make([]string, 0, len(raw))
	for _, item := range raw {
		part, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("list item = %v, expected an object", item)
		}
		name, _ := part["name"].(string)
		names = append(names, name)
	}
	return names
}

func TestListRespectsLimitAndOffset(t *testing.T) {
	app := newTestApp()
	for _, name := range []string{"A", "B", "C", "D", "E"} {
		createPart(t, app, name, "engine")
	}

	status, body := do(t, app, http.MethodGet, "/parts?limit=2&offset=2", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}

	names := partNames(t, body)
	if len(names) != 2 || names[0] != "C" || names[1] != "D" {
		t.Errorf("list = %v, expected [C D]", names)
	}
}

func TestListIgnoresInvalidPagination(t *testing.T) {
	app := newTestApp()
	createPart(t, app, "A", "engine")

	status, body := do(t, app, http.MethodGet, "/parts?limit=abc&offset=xyz", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}
	if names := partNames(t, body); len(names) != 1 {
		t.Errorf("list = %v, expected 1 part", names)
	}
}

func TestCategoryFilterIsCaseInsensitive(t *testing.T) {
	app := newTestApp()
	createPart(t, app, "Filtro de Óleo X", "engine")
	createPart(t, app, "Pastilha de Freio Y", "brakes")

	status, body := do(t, app, http.MethodGet, "/parts?category=ENGINE", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}

	names := partNames(t, body)
	if len(names) != 1 || names[0] != "Filtro de Óleo X" {
		t.Errorf("list = %v, expected only the engine part", names)
	}
}

func TestCategoryStoredLowercase(t *testing.T) {
	app := newTestApp()

	body := `{
		"name": "Correia Dentada Z",
		"category": "  ENGINE  ",
		"currentStock": 15,
		"minimumStock": 20,
		"averageDailySales": 4,
		"leadTimeDays": 5,
		"unitCost": 18.50,
		"criticalityLevel": 3
	}`
	status, decoded := do(t, app, http.MethodPost, "/parts", body)
	if status != http.StatusCreated {
		t.Fatalf("status = %d, expected 201", status)
	}
	if decoded["category"] != "engine" {
		t.Errorf("category = %v, expected %q", decoded["category"], "engine")
	}
}

func TestUpdateNotFoundReturns404(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodPut,
		"/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff", validBody)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, expected 404", status)
	}
}

func TestUpdateInvalidBodyReturns422(t *testing.T) {
	app := newTestApp()
	id := createPart(t, app, "Filtro de Óleo X", "engine")

	body := `{"name": "Filtro", "category": "engine", "criticalityLevel": 42}`
	status, decoded := do(t, app, http.MethodPut, "/parts/"+id, body)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, expected 422", status)
	}
	if _, ok := decoded["fields"].([]any); !ok {
		t.Errorf("expected the list of rejected fields, got %v", decoded)
	}

	status, found := do(t, app, http.MethodGet, "/parts/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("get: status = %d", status)
	}
	if found["name"] != "Filtro de Óleo X" {
		t.Errorf("name = %v, expected the part untouched", found["name"])
	}
}

func TestUpdateMalformedIDReturns400(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodPut, "/parts/not-a-uuid", validBody)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, expected 400", status)
	}
}

func TestDeleteNotFoundReturns404(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodDelete,
		"/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff", "")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, expected 404", status)
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodGet, "/does-not-exist", "")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, expected 404", status)
	}
}

func TestNegativeStockThroughAPI(t *testing.T) {
	app := newTestApp()

	body := `{
		"name": "Pastilha de Freio Y",
		"category": "brakes",
		"currentStock": -42,
		"minimumStock": 10,
		"averageDailySales": 2.5,
		"leadTimeDays": 7,
		"unitCost": 99.99,
		"criticalityLevel": 5
	}`
	status, decoded := do(t, app, http.MethodPost, "/parts", body)
	if status != http.StatusCreated {
		t.Fatalf("status = %d, expected 201 (body: %v)", status, decoded)
	}
	if decoded["currentStock"] != float64(-42) {
		t.Errorf("currentStock = %v, expected -42", decoded["currentStock"])
	}
}
