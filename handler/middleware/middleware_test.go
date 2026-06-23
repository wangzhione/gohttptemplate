// go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wangzhione/gohttptemplate/configs"
)

func TestResponseWriterPanicError(t *testing.T) {
	rec := httptest.NewRecorder()
	ResponseWriterPanicError(rec)

	if rec.Code != 589 {
		t.Errorf("expected status code 589, got %d", rec.Code)
	}

	expectedBody := `{"code":"589", "message":"Internal Server Panic Error"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestMainMiddlewarePprofAuthDisabled(t *testing.T) {
	oldConfig := configs.G
	configs.G = &configs.Config{}
	defer func() {
		configs.G = oldConfig
	}()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	MainMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestMainMiddlewarePprofBearer(t *testing.T) {
	oldConfig := configs.G
	configs.G = &configs.Config{}
	configs.G.Serve.PprofBearer = "secret"
	defer func() {
		configs.G = oldConfig
	}()

	t.Run("invalid token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unauthorized pprof request should not reach next handler")
		})

		MainMiddleware(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
		req.Header.Set("pprofbearer", "Bearer secret")
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		MainMiddleware(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status code %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
