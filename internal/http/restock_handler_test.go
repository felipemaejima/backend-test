package http_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func createPartWithStock(t *testing.T, app *fiber.App, name string, currentStock, minimumStock int, averageDailySales float64, leadTimeDays, criticalityLevel int) {
	t.Helper()

	body := fmt.Sprintf(`{
		"name": %q,
		"category": "engine",
		"currentStock": %d,
		"minimumStock": %d,
		"averageDailySales": %v,
		"leadTimeDays": %d,
		"unitCost": 18.50,
		"criticalityLevel": %d
	}`, name, currentStock, minimumStock, averageDailySales, leadTimeDays, criticalityLevel)

	status, decoded := do(t, app, http.MethodPost, "/parts", body)
	if status != http.StatusCreated {
		t.Fatalf("creating %q: status = %d (%v)", name, status, decoded)
	}
}

func priorities(t *testing.T, body map[string]any) []map[string]any {
	t.Helper()

	raw, ok := body["priorities"].([]any)
	if !ok {
		t.Fatalf("priorities = %v, expected an array", body["priorities"])
	}

	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		priority, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("list item = %v, expected an object", item)
		}
		items = append(items, priority)
	}
	return items
}

func TestRestockPriorities(t *testing.T) {
	app := newTestApp()
	createPartWithStock(t, app, "Healthy", 500, 20, 1, 2, 3)
	createPartWithStock(t, app, "Critical", -42, 20, 4, 5, 5)
	createPartWithStock(t, app, "Moderate", 8, 20, 4, 5, 4)

	status, body := do(t, app, http.MethodGet, "/restock/priorities", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}

	items := priorities(t, body)
	if len(items) != 2 {
		t.Fatalf("expected 2 priorities, got %d (%v)", len(items), items)
	}

	first := items[0]
	if first["name"] != "Critical" {
		t.Errorf("first = %v, expected Critical", first["name"])
	}
	if first["projectedStock"] != float64(-62) {
		t.Errorf("projectedStock = %v, expected -62", first["projectedStock"])
	}
	if first["urgencyScore"] != float64(410) {
		t.Errorf("urgencyScore = %v, expected 410", first["urgencyScore"])
	}
	if first["currentStock"] != float64(-42) {
		t.Errorf("currentStock = %v, expected -42", first["currentStock"])
	}
	if first["minimumStock"] != float64(20) {
		t.Errorf("minimumStock = %v, expected 20", first["minimumStock"])
	}
	if id, _ := first["partId"].(string); id == "" {
		t.Error("expected partId in the response")
	}

	if items[1]["name"] != "Moderate" {
		t.Errorf("second = %v, expected Moderate", items[1]["name"])
	}
}

func TestRestockPrioritiesEmptyReturnsEmptyArray(t *testing.T) {
	status, body := do(t, newTestApp(), http.MethodGet, "/restock/priorities", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}
	if items := priorities(t, body); len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

func TestRestockPrioritiesOnlyIncludesPartsBelowMinimum(t *testing.T) {
	app := newTestApp()
	createPartWithStock(t, app, "Exactly At Minimum", 30, 10, 2, 10, 3)
	createPartWithStock(t, app, "Just Below", 29, 10, 2, 10, 3)

	status, body := do(t, app, http.MethodGet, "/restock/priorities", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}

	items := priorities(t, body)
	if len(items) != 1 {
		t.Fatalf("expected 1 priority, got %d (%v)", len(items), items)
	}
	if items[0]["name"] != "Just Below" {
		t.Errorf("name = %v, expected Just Below", items[0]["name"])
	}
}
