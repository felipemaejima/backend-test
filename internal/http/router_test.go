package http_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	httpapi "github.com/felipemaejima/backend-test/internal/http"
	"github.com/felipemaejima/backend-test/internal/repository/memory"
	"github.com/felipemaejima/backend-test/internal/service"
)

func newTestApp() *fiber.App {
	repo := memory.NewPartRepository()
	return httpapi.NewRouter(
		httpapi.NewPartHandler(service.NewPartService(repo)),
		httpapi.NewRestockHandler(service.NewRestockService(repo)),
	)
}

const validBody = `{
	"name": "Filtro de Óleo X",
	"category": "engine",
	"currentStock": 15,
	"minimumStock": 20,
	"averageDailySales": 4,
	"leadTimeDays": 5,
	"unitCost": 18.50,
	"criticalityLevel": 3
}`

func do(t *testing.T, app *fiber.App, method, target, body string) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if len(raw) == 0 {
		return resp.StatusCode, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("body is not valid JSON (%s): %v", raw, err)
	}
	return resp.StatusCode, decoded
}

func TestHealth(t *testing.T) {
	status, body := do(t, newTestApp(), http.MethodGet, "/health", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}
	if body["status"] != "ok" {
		t.Errorf("body = %v", body)
	}
}

func TestCreatePart(t *testing.T) {
	status, body := do(t, newTestApp(), http.MethodPost, "/parts", validBody)

	if status != http.StatusCreated {
		t.Fatalf("status = %d, expected 201 (body: %v)", status, body)
	}
	if body["id"] == "" || body["id"] == nil {
		t.Error("expected an id in the response")
	}
	if body["name"] != "Filtro de Óleo X" {
		t.Errorf("name = %v", body["name"])
	}
	if body["criticalityLevel"] != float64(3) {
		t.Errorf("criticalityLevel = %v, expected 3", body["criticalityLevel"])
	}
}

func TestCreatePartMalformedBody(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodPost, "/parts", `{"name":`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, expected 400", status)
	}
}

func TestCreatePartInvalidReturns422WithFields(t *testing.T) {
	body := `{"name": "", "category": "engine", "criticalityLevel": 9}`
	status, decoded := do(t, newTestApp(), http.MethodPost, "/parts", body)

	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, expected 422", status)
	}

	fields, ok := decoded["fields"].([]any)
	if !ok {
		t.Fatalf("expected a list of fields in the response, got %v", decoded)
	}
	if len(fields) != 2 {
		t.Errorf("expected 2 rejected fields, got %d: %v", len(fields), fields)
	}
}

func TestGetPartNotFoundReturns404(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodGet,
		"/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff", "")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, expected 404", status)
	}
}

func TestMalformedIDReturns400(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodGet, "/parts/not-a-uuid", "")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, expected 400", status)
	}
}

func TestFullCRUD(t *testing.T) {
	app := newTestApp()

	status, created := do(t, app, http.MethodPost, "/parts", validBody)
	if status != http.StatusCreated {
		t.Fatalf("create: status = %d", status)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("create: no id in the response")
	}

	status, found := do(t, app, http.MethodGet, "/parts/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("get: status = %d", status)
	}
	if found["id"] != id {
		t.Errorf("get: id = %v, expected %s", found["id"], id)
	}

	updateBody := strings.Replace(validBody, `"currentStock": 15`, `"currentStock": 2`, 1)
	status, updated := do(t, app, http.MethodPut, "/parts/"+id, updateBody)
	if status != http.StatusOK {
		t.Fatalf("update: status = %d", status)
	}
	if updated["currentStock"] != float64(2) {
		t.Errorf("update: currentStock = %v, expected 2", updated["currentStock"])
	}
	if updated["id"] != id {
		t.Error("update: id should not change")
	}

	status, list := do(t, app, http.MethodGet, "/parts?category=engine", "")
	if status != http.StatusOK {
		t.Fatalf("list: status = %d", status)
	}
	parts, ok := list["parts"].([]any)
	if !ok || len(parts) != 1 {
		t.Fatalf("list: expected 1 part, got %v", list["parts"])
	}

	status, _ = do(t, app, http.MethodDelete, "/parts/"+id, "")
	if status != http.StatusNoContent {
		t.Fatalf("delete: status = %d, expected 204", status)
	}

	status, _ = do(t, app, http.MethodGet, "/parts/"+id, "")
	if status != http.StatusNotFound {
		t.Fatalf("get after delete: status = %d, expected 404", status)
	}

	status, _ = do(t, app, http.MethodDelete, "/parts/"+id, "")
	if status != http.StatusNotFound {
		t.Fatalf("repeated delete: status = %d, expected 404", status)
	}
}

func TestEmptyListReturnsEmptyArray(t *testing.T) {
	status, body := do(t, newTestApp(), http.MethodGet, "/parts", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200", status)
	}
	parts, ok := body["parts"].([]any)
	if !ok {
		t.Fatalf("parts = %v, expected an array", body["parts"])
	}
	if len(parts) != 0 {
		t.Errorf("expected empty array, got %d items", len(parts))
	}
}
