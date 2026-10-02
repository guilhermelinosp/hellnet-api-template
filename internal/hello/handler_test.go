package hello

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guilhermelinosp/fast-platform/platform"
)

// ───────────────────── Service unit tests (framework-free) ─────────────────

type stubService struct {
	message string
	err     error
}

func (s *stubService) Greet(_ context.Context, _ string) (string, error) {
	return s.message, s.err
}

func TestServiceGreeting(t *testing.T) {
	svc := NewService(slog.Default())

	msg, err := svc.Greet(context.Background(), " ana ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Hello, ana!" {
		t.Fatalf("got %q", msg)
	}
}

func TestServiceRejectsOverlongNames(t *testing.T) {
	svc := NewService(slog.Default())
	_, err := svc.Greet(context.Background(), strings.Repeat("x", 200))

	var httpErr *platform.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusBadRequest {
		t.Fatalf("expected VALIDATION_ERROR 400, got %v", err)
	}
}

// ─────────────────── Handler tests through a real gin router ───────────────

func newRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(svc).Register(router.Group("/api/v1"))
	return router
}

func do(router http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func assertGreeting(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMsg string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status %d ≠ %d (body %s)", rec.Code, wantStatus, rec.Body.String())
	}
	var out greetResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not the greeting JSON: %v (%s)", err, rec.Body.String())
	}
	if out.Message != wantMsg {
		t.Fatalf("message %q ≠ %q", out.Message, wantMsg)
	}
}

func TestHandlerHappyPaths(t *testing.T) {
	router := newRouter(&stubService{message: "Hello, x!"})

	t.Run("query", func(t *testing.T) {
		assertGreeting(t, do(router, http.MethodGet, "/api/v1/hello?name=world", ""), http.StatusOK, "Hello, x!")
	})
	t.Run("path", func(t *testing.T) {
		assertGreeting(t, do(router, http.MethodGet, "/api/v1/hello/gin", ""), http.StatusOK, "Hello, x!")
	})
	t.Run("body", func(t *testing.T) {
		assertGreeting(t, do(router, http.MethodPost, "/api/v1/hello", `{"name":"ana"}`), http.StatusCreated, "Hello, x!")
	})
}

func TestHandlerInputFailures(t *testing.T) {
	router := newRouter(&stubService{message: "unused"})

	cases := map[string]struct{ method, target, body string }{
		"blank path name":       {http.MethodGet, "/api/v1/hello/%20", ""},
		"empty body name":       {http.MethodPost, "/api/v1/hello", `{}`},
		"malformed body":        {http.MethodPost, "/api/v1/hello", `not-json`},
		"unknown body field":    {http.MethodPost, "/api/v1/hello", `{"name":"a","extra":1}`},
		"wrong body field type": {http.MethodPost, "/api/v1/hello", `{"name":1}`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := do(router, tc.method, tc.target, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status %d, want 400 (body %s)", rec.Code, rec.Body.String())
			}
			var env struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || env.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("want the VALIDATION_ERROR envelope, got %s", rec.Body.String())
			}
		})
	}
}

func TestHandlerMapsServiceErrors(t *testing.T) {
	router := newRouter(&stubService{err: platform.ValidationError("name", "is invalid")})
	if rec := do(router, http.MethodGet, "/api/v1/hello?name=x", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("service validation error must be a 400, got %d", rec.Code)
	}
}

func TestRoutesAreMounted(t *testing.T) {
	routes := map[string]bool{}
	for _, r := range newRouter(&stubService{}).Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{"GET /api/v1/hello", "GET /api/v1/hello/:name", "POST /api/v1/hello"} {
		if !routes[want] {
			t.Errorf("route %q is not mounted (have %v)", want, routes)
		}
	}
}
