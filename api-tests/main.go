package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	green = "\033[32m"
	red   = "\033[31m"
	dim   = "\033[2m"
	bold  = "\033[1m"
	reset = "\033[0m"
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		green, red, dim, bold, reset = "", "", "", "", ""
	}
}

func main() {
	baseURL := strings.TrimSuffix(envOr("BASE_URL", "http://localhost:8080"), "/")

	r := &runner{
		baseURL:  baseURL,
		client:   &http.Client{Timeout: 10 * time.Second},
		category: "smoke-" + time.Now().Format("150405"),
	}

	fmt.Printf("%sAPI tests%s  →  %s  %s(category: %s)%s\n",
		bold, reset, baseURL, dim, r.category, reset)

	if res := r.request(http.MethodGet, "/health", ""); res.err != nil {
		fmt.Printf("\n%sAPI unreachable at %s%s\n", red, baseURL, reset)
		fmt.Printf("%v\n\nStart the environment with `make up` before running.\n", res.err)
		os.Exit(1)
	}

	defer r.cleanup()

	r.testHealth()
	partID := r.testCreate()
	r.testValidation()
	r.testRead(partID)
	r.testList()
	r.testUpdate(partID)
	r.testDelete(partID)
	r.testRoutes()

	os.Exit(r.report())
}

func (r *runner) testHealth() {
	r.section("Health")

	res := r.request(http.MethodGet, "/health", "")
	r.expectStatus(res, http.StatusOK, "GET /health")
	r.expectField(res, "status", "ok", "body reports status ok")
}

func (r *runner) testCreate() string {
	r.section("Create")

	res := r.request(http.MethodPost, "/parts", r.partBody("Filtro Smoke", r.category, 15, 3))
	r.expectStatus(res, http.StatusCreated, "POST /parts creates the part")
	r.expectField(res, "name", "Filtro Smoke", "response echoes the name")
	partID := r.trackCreated(res)
	r.check(partID != "", "response carries an id", "no id in the body")

	res = r.request(http.MethodPost, "/parts",
		r.partBody("  Correia Smoke  ", "  "+strings.ToUpper(r.category)+"  ", 8, 4))
	r.expectStatus(res, http.StatusCreated, "POST /parts with padding and uppercase")
	r.expectField(res, "name", "Correia Smoke", "name is trimmed")
	r.expectField(res, "category", r.category, "category is lowercased")
	r.trackCreated(res)

	res = r.request(http.MethodPost, "/parts", r.partBody("Pastilha Smoke", r.category, -42, 5))
	r.expectStatus(res, http.StatusCreated, "POST /parts accepts negative stock")
	r.expectField(res, "currentStock", -42, "negative stock is preserved")
	r.trackCreated(res)

	return partID
}

func (r *runner) testValidation() {
	r.section("Validation")

	res := r.request(http.MethodPost, "/parts",
		`{"name": "", "category": "engine", "criticalityLevel": 9}`)
	r.expectStatus(res, http.StatusUnprocessableEntity, "POST /parts with invalid data")
	r.check(hasField(res, "name"), "points at the name field", "field missing from the response")
	r.check(hasField(res, "criticalityLevel"), "points at the criticalityLevel field", "field missing from the response")

	res = r.request(http.MethodPost, "/parts", `{"name":`)
	r.expectStatus(res, http.StatusBadRequest, "POST /parts with malformed JSON")

	res = r.request(http.MethodPost, "/parts", r.partBody("Criticidade Alta", r.category, 10, 9))
	r.expectStatus(res, http.StatusUnprocessableEntity, "POST /parts with criticality outside 1-5")
}

func (r *runner) testRead(partID string) {
	r.section("Read")

	res := r.request(http.MethodGet, "/parts/"+partID, "")
	r.expectStatus(res, http.StatusOK, "GET /parts/:id returns the part")
	r.expectField(res, "id", partID, "id matches")

	res = r.request(http.MethodGet, "/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff", "")
	r.expectStatus(res, http.StatusNotFound, "GET /parts/:id not found")

	res = r.request(http.MethodGet, "/parts/not-a-uuid", "")
	r.expectStatus(res, http.StatusBadRequest, "GET /parts/:id with malformed uuid")
}

func (r *runner) testList() {
	r.section("List")

	res := r.request(http.MethodGet, "/parts?category="+r.category+"&limit=100", "")
	r.expectStatus(res, http.StatusOK, "GET /parts?category=...")
	r.expectNames(res, []string{"Correia Smoke", "Filtro Smoke", "Pastilha Smoke"},
		"returns this run's parts ordered by name")

	res = r.request(http.MethodGet, "/parts?category="+strings.ToUpper(r.category)+"&limit=100", "")
	r.expectStatus(res, http.StatusOK, "category filter is case insensitive")
	r.expectCount(res, 3, "finds the same parts with an uppercase category")

	res = r.request(http.MethodGet, "/parts?category=category-that-does-not-exist", "")
	r.expectStatus(res, http.StatusOK, "category with no parts")
	r.check(isEmptyArray(res), "returns an empty array, not null", "parts field is not an empty array")

	res = r.request(http.MethodGet, "/parts?category="+r.category+"&limit=2", "")
	r.expectStatus(res, http.StatusOK, "GET /parts?limit=2")
	r.expectCount(res, 2, "respects the limit")

	res = r.request(http.MethodGet, "/parts?category="+r.category+"&limit=2&offset=1", "")
	r.expectStatus(res, http.StatusOK, "GET /parts?limit=2&offset=1")
	r.expectNames(res, []string{"Filtro Smoke", "Pastilha Smoke"}, "pagination advances in order")

	res = r.request(http.MethodGet, "/parts?category="+r.category+"&limit=abc&offset=xyz", "")
	r.expectStatus(res, http.StatusOK, "invalid pagination falls back to defaults")
	r.expectCount(res, 3, "full list with defaults")
}

func (r *runner) testUpdate(partID string) {
	r.section("Update")

	res := r.request(http.MethodPut, "/parts/"+partID, r.partBody("Filtro Smoke v2", r.category, 2, 3))
	r.expectStatus(res, http.StatusOK, "PUT /parts/:id updates")
	r.expectField(res, "name", "Filtro Smoke v2", "name updated")
	r.expectField(res, "currentStock", 2, "stock updated")
	r.expectField(res, "id", partID, "id preserved")

	res = r.request(http.MethodGet, "/parts/"+partID, "")
	r.expectField(res, "name", "Filtro Smoke v2", "change was persisted")

	res = r.request(http.MethodPut, "/parts/"+partID,
		`{"name": "x", "category": "engine", "criticalityLevel": 99}`)
	r.expectStatus(res, http.StatusUnprocessableEntity, "PUT /parts/:id with invalid body")

	res = r.request(http.MethodGet, "/parts/"+partID, "")
	r.expectField(res, "name", "Filtro Smoke v2", "part untouched after invalid update")

	res = r.request(http.MethodPut, "/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff",
		r.partBody("Fantasma", r.category, 10, 3))
	r.expectStatus(res, http.StatusNotFound, "PUT /parts/:id not found")

	res = r.request(http.MethodPut, "/parts/not-a-uuid", r.partBody("X", r.category, 10, 3))
	r.expectStatus(res, http.StatusBadRequest, "PUT /parts/:id with malformed uuid")
}

func (r *runner) testDelete(partID string) {
	r.section("Delete")

	res := r.request(http.MethodDelete, "/parts/"+partID, "")
	r.expectStatus(res, http.StatusNoContent, "DELETE /parts/:id removes")

	res = r.request(http.MethodGet, "/parts/"+partID, "")
	r.expectStatus(res, http.StatusNotFound, "part no longer exists")

	res = r.request(http.MethodDelete, "/parts/"+partID, "")
	r.expectStatus(res, http.StatusNotFound, "repeated DELETE returns 404, not 204")
}

func (r *runner) testRoutes() {
	r.section("Routes")

	res := r.request(http.MethodGet, "/does-not-exist", "")
	r.expectStatus(res, http.StatusNotFound, "unknown route")
}

type result struct {
	status int
	body   map[string]any
	raw    string
	err    error
}

type runner struct {
	baseURL  string
	client   *http.Client
	category string

	passed  int
	failed  int
	created []string
}

func (r *runner) request(method, path, body string) result {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, r.baseURL+path, reader)
	if err != nil {
		return result{err: err}
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return result{err: err}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return result{status: resp.StatusCode, err: err}
	}

	res := result{status: resp.StatusCode, raw: string(raw)}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &res.body)
	}
	return res
}

func (r *runner) partBody(name, category string, stock, criticality int) string {
	return fmt.Sprintf(`{
		"name": %q,
		"category": %q,
		"currentStock": %d,
		"minimumStock": 20,
		"averageDailySales": 4,
		"leadTimeDays": 5,
		"unitCost": 18.50,
		"criticalityLevel": %d
	}`, name, category, stock, criticality)
}

func (r *runner) trackCreated(res result) string {
	id, _ := res.body["id"].(string)
	if id != "" {
		r.created = append(r.created, id)
	}
	return id
}

func (r *runner) cleanup() {
	for _, id := range r.created {
		r.request(http.MethodDelete, "/parts/"+id, "")
	}
}

func (r *runner) check(ok bool, label, detail string) {
	if ok {
		r.passed++
		fmt.Printf("  %s✓%s %s\n", green, reset, label)
		return
	}
	r.failed++
	fmt.Printf("  %s✗%s %s\n      %s%s%s\n", red, reset, label, dim, detail, reset)
}

func (r *runner) expectStatus(res result, want int, label string) {
	if res.err != nil {
		r.check(false, label, "request error: "+res.err.Error())
		return
	}
	r.check(res.status == want, label,
		fmt.Sprintf("expected status %d, got %d — %s", want, res.status, truncate(res.raw)))
}

func (r *runner) expectField(res result, key string, want any, label string) {
	got := fmt.Sprintf("%v", res.body[key])
	wantStr := fmt.Sprintf("%v", want)
	r.check(got == wantStr, label,
		fmt.Sprintf("field %q = %s, expected %s", key, got, wantStr))
}

func (r *runner) expectCount(res result, want int, label string) {
	got := len(partNames(res))
	r.check(got == want, label, fmt.Sprintf("expected %d part(s), got %d", want, got))
}

func (r *runner) expectNames(res result, want []string, label string) {
	got := partNames(res)
	r.check(equal(got, want), label, fmt.Sprintf("list = %v, expected %v", got, want))
}

func (r *runner) section(title string) {
	fmt.Printf("\n%s%s%s\n", bold, title, reset)
}

func (r *runner) report() int {
	total := r.passed + r.failed
	fmt.Println()
	if r.failed == 0 {
		fmt.Printf("%s%d/%d checks passed.%s\n", green, r.passed, total, reset)
		return 0
	}
	fmt.Printf("%s%d of %d checks failed.%s\n", red, r.failed, total, reset)
	return 1
}

func hasField(res result, name string) bool {
	fields, ok := res.body["fields"].([]any)
	if !ok {
		return false
	}
	for _, item := range fields {
		field, ok := item.(map[string]any)
		if ok && field["field"] == name {
			return true
		}
	}
	return false
}

func partNames(res result) []string {
	parts, ok := res.body["parts"].([]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(parts))
	for _, item := range parts {
		part, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := part["name"].(string)
		names = append(names, name)
	}
	return names
}

func isEmptyArray(res result) bool {
	parts, ok := res.body["parts"].([]any)
	return ok && len(parts) == 0
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func truncate(s string) string {
	const max = 200
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
