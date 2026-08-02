package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/repository/memory"
)

func newLoggingApp(t *testing.T, repo domain.PartRepository) (*fiber.App, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}
	log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return newTestAppWith(repo, func(context.Context) error { return nil }, log), buf
}

func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	lines := make([]map[string]any, 0)
	for decoder.More() {
		var line map[string]any
		if err := decoder.Decode(&line); err != nil {
			t.Fatalf("linha de log não é JSON válido: %v (%s)", err, buf.String())
		}
		lines = append(lines, line)
	}
	return lines
}

func lastLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	lines := logLines(t, buf)
	if len(lines) == 0 {
		t.Fatal("esperava ao menos uma linha de log")
	}
	return lines[len(lines)-1]
}

func TestRequestIsLogged(t *testing.T) {
	app, buf := newLoggingApp(t, memory.NewPartRepository())

	if status, _ := do(t, app, http.MethodPost, "/parts", validBody); status != http.StatusCreated {
		t.Fatalf("status = %d, expected 201", status)
	}

	line := lastLine(t, buf)

	if line["msg"] != "requisição" {
		t.Errorf("msg = %v", line["msg"])
	}
	if line["method"] != http.MethodPost {
		t.Errorf("method = %v, expected POST", line["method"])
	}
	if line["path"] != "/parts" {
		t.Errorf("path = %v, expected /parts", line["path"])
	}
	if line["status"] != float64(http.StatusCreated) {
		t.Errorf("status = %v, expected 201", line["status"])
	}
	if line["level"] != "INFO" {
		t.Errorf("level = %v, expected INFO", line["level"])
	}
	if id, _ := line["request_id"].(string); id == "" {
		t.Error("expected a request_id in the log line")
	}
	if _, ok := line["duration_ms"].(float64); !ok {
		t.Errorf("duration_ms = %v, expected a number", line["duration_ms"])
	}
}

func TestLogLevelFollowsStatus(t *testing.T) {
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   string
	}{
		{"success is info", http.MethodGet, "/parts", "", "INFO"},
		{"client error is warn", http.MethodGet, "/parts/not-a-uuid", "", "WARN"},
		{"not found is warn", http.MethodGet, "/does-not-exist", "", "WARN"},
		{"health is debug", http.MethodGet, "/health", "", "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, buf := newLoggingApp(t, memory.NewPartRepository())

			do(t, app, tt.method, tt.target, tt.body)

			if level := lastLine(t, buf)["level"]; level != tt.want {
				t.Errorf("level = %v, expected %v", level, tt.want)
			}
		})
	}
}

func TestInternalErrorIsLoggedWithDetail(t *testing.T) {
	app, buf := newLoggingApp(t, failingRepository{err: errTestRepository})

	if status, _ := do(t, app, http.MethodGet, "/parts", ""); status != http.StatusInternalServerError {
		t.Fatalf("status = %d, expected 500", status)
	}

	line := lastLine(t, buf)

	if line["level"] != "ERROR" {
		t.Errorf("level = %v, expected ERROR", line["level"])
	}
	// O detalhe fica no log, nunca na resposta — o inverso é coberto em errors_test.go.
	if detail, _ := line["error"].(string); detail != internalErrorDetail {
		t.Errorf("error = %v, expected the underlying detail", line["error"])
	}
}

func TestRequestIDIsSharedWithTheResponseHeader(t *testing.T) {
	app, buf := newLoggingApp(t, memory.NewPartRepository())

	req := httptest.NewRequest(http.MethodGet, "/parts", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	header := resp.Header.Get(fiber.HeaderXRequestID)
	if header == "" {
		t.Fatal("expected the X-Request-ID header")
	}
	if logged := lastLine(t, buf)["request_id"]; logged != header {
		t.Errorf("request_id = %v, expected it to match the header %q", logged, header)
	}
}
