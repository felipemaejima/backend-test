package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func partBodyWith(field, value string) string {
	fields := map[string]string{
		"name":              `"Filtro de Óleo X"`,
		"category":          `"engine"`,
		"currentStock":      "15",
		"minimumStock":      "20",
		"averageDailySales": "4",
		"leadTimeDays":      "5",
		"unitCost":          "18.50",
		"criticalityLevel":  "3",
	}
	fields[field] = value

	pairs := make([]string, 0, len(fields))
	for key, raw := range fields {
		pairs = append(pairs, fmt.Sprintf("%q: %s", key, raw))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

func TestPutReplacesTheWholeResource(t *testing.T) {
	app := newTestApp()
	id := createPart(t, app, "Filtro de Óleo X", "engine")

	body := `{"name": "Filtro de Óleo X", "category": "engine", "criticalityLevel": 3}`
	status, updated := do(t, app, http.MethodPut, "/parts/"+id, body)

	if status != http.StatusOK {
		t.Fatalf("status = %d, expected 200 (body: %v)", status, updated)
	}
	for _, field := range []string{"currentStock", "minimumStock", "averageDailySales", "leadTimeDays", "unitCost"} {
		if updated[field] != float64(0) {
			t.Errorf("%s = %v, expected 0: PUT replaces the whole resource, so omitted fields reset",
				field, updated[field])
		}
	}
}

func TestInvalidFieldTypesAreRejected(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"fraction in an integer field", partBodyWith("currentStock", "1.5")},
		{"string in a numeric field", partBodyWith("currentStock", `"abc"`)},
		{"number in a text field", partBodyWith("name", "123")},
		{"boolean in a numeric field", partBodyWith("criticalityLevel", "true")},
		{"integer far beyond int64", partBodyWith("currentStock", "99999999999999999999")},
		{"array where an object is expected", `[{"name": "X"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := do(t, newTestApp(), http.MethodPost, "/parts", tt.body)

			if status != http.StatusBadRequest {
				t.Errorf("status = %d, expected 400 (body: %v)", status, body)
			}
		})
	}
}

func TestUnknownFieldsAreIgnored(t *testing.T) {
	body := partBodyWith("unexpectedField", `"whatever"`)

	status, decoded := do(t, newTestApp(), http.MethodPost, "/parts", body)
	if status != http.StatusCreated {
		t.Fatalf("status = %d, expected 201 (body: %v)", status, decoded)
	}
	if _, present := decoded["unexpectedField"]; present {
		t.Error("the unknown field leaked into the response")
	}
}

func TestEmptyBodyIsRejected(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodPost, "/parts", " ")
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, expected 400", status)
	}
}

func TestJSONNullBodyIsRejected(t *testing.T) {
	status, _ := do(t, newTestApp(), http.MethodPost, "/parts", "null")

	if status != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, expected 422", status)
	}
}

func TestMissingContentTypeIsRejected(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/parts", strings.NewReader(validBody))

	resp, err := newTestApp().Test(req, 5000)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, expected 400 without Content-Type", resp.StatusCode)
	}
}
