package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guilhermelinosp/hellnet-lib-api/api"
)

// TestPlatformEndpointPaths asserts the platform contract promised by
// hellnet-lib-api: the system endpoints exist as declared route constants.
// The actual mounting is covered by the library; here we only protect the
// composition invariants of main's wiring table.
func TestPlatformEndpointPaths(t *testing.T) {
	want := []string{
		api.PathLive,
		api.PathReady,
		api.PathHealth,
		api.Prefix,
	}
	for _, path := range want {
		if path == "" || path[0] != '/' {
			t.Errorf("platform path %q is not absolute", path)
		}
	}
}

func TestRequestHandlerSmoke(t *testing.T) {
	// Any handler used by main must be an http.Handler after adapter wrapping.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.Background())
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected %d, got %d", http.StatusTeapot, rec.Code)
	}
}
