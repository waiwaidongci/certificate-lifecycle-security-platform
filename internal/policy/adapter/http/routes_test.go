package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRejectsNilRouteDependencies(t *testing.T) {
	invalid := []struct {
		name    string
		handler *Handler
	}{
		{name: "nil handler"},
		{name: "missing service", handler: NewHandler(nil)},
	}

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mustNotPanic(t, func() { test.handler.Register(mux) })

			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/policies", nil))
			if response.Code != http.StatusNotFound {
				t.Fatalf("invalid handler registered policy route: status=%d", response.Code)
			}
		})
	}

	t.Run("nil mux", func(t *testing.T) {
		mustNotPanic(t, func() { NewHandler(nil).Register(nil) })
	})
}

func mustNotPanic(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("registration panicked: %v", recovered)
		}
	}()
	call()
}
