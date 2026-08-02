package http_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

const internalErrorDetail = "pq: connection refused on 10.0.0.7:5432"

var errTestRepository = errors.New(internalErrorDetail)

type failingRepository struct {
	err error
}

var _ domain.PartRepository = failingRepository{}

func (r failingRepository) Create(context.Context, *domain.Part) error { return r.err }
func (r failingRepository) Update(context.Context, *domain.Part) error { return r.err }
func (r failingRepository) Delete(context.Context, uuid.UUID) error    { return r.err }

func (r failingRepository) FindByID(context.Context, uuid.UUID) (*domain.Part, error) {
	return nil, r.err
}

func (r failingRepository) List(context.Context, domain.PartFilter) (domain.Page[domain.Part], error) {
	return domain.Page[domain.Part]{}, r.err
}

func (r failingRepository) ListAll(context.Context) ([]domain.Part, error) {
	return nil, r.err
}

func newFailingApp() *fiber.App {
	return newTestAppWith(
		failingRepository{err: errTestRepository},
		func(context.Context) error { return nil },
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestRepositoryFailureReturns500(t *testing.T) {
	const someID = "/parts/1e2d3c4b-5a69-4788-9900-aabbccddeeff"

	tests := []struct {
		name   string
		method string
		target string
		body   string
	}{
		{"create", http.MethodPost, "/parts", validBody},
		{"list", http.MethodGet, "/parts", ""},
		{"get by id", http.MethodGet, someID, ""},
		{"update", http.MethodPut, someID, validBody},
		{"delete", http.MethodDelete, someID, ""},
		{"restock priorities", http.MethodGet, "/restock/priorities", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := do(t, newFailingApp(), tt.method, tt.target, tt.body)

			if status != http.StatusInternalServerError {
				t.Fatalf("status = %d, expected 500 (body: %v)", status, body)
			}
			if body["error"] != "erro interno" {
				t.Errorf("error = %v, expected a generic message", body["error"])
			}
			if body["fields"] != nil {
				t.Errorf("fields = %v, expected no field list on an internal error", body["fields"])
			}
		})
	}
}

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	resp, err := newFailingApp().Test(httptest.NewRequest(http.MethodGet, "/parts", nil), 5000)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	for _, leak := range []string{internalErrorDetail, "connection refused", "10.0.0.7"} {
		if strings.Contains(string(raw), leak) {
			t.Errorf("response leaks %q: %s", leak, raw)
		}
	}
}
