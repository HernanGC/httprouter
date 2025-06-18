package httprouter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Helper for creating a string result handler
func stringHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}
}

func TestNewApplication(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	// Check that the application implements the WebApplication interface
	_, ok := app.(WebApplication)
	if !ok {
		t.Errorf("Expected NewApplication to return WebApplication, got %T", app)
	}
}

func TestApplication_HandleMethods(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Get("/test", stringHandler("GET"))
	app.Post("/test", stringHandler("POST"))
	app.Put("/test", stringHandler("PUT"))
	app.Patch("/test", stringHandler("PATCH"))
	app.Delete("/test", stringHandler("DELETE"))

	tests := []struct {
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{"GET", http.StatusOK, "GET"},
		{"POST", http.StatusOK, "POST"},
		{"PUT", http.StatusOK, "PUT"},
		{"PATCH", http.StatusOK, "PATCH"},
		{"DELETE", http.StatusOK, "DELETE"},
		{"OPTIONS", http.StatusMethodNotAllowed, "Method Not Allowed\n"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/test", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("%s /test: got status %d, want %d", tc.method, resp.StatusCode, tc.expectedStatus)
			}
			if string(body) != tc.expectedBody {
				t.Errorf("%s /test: got body %q, want %q", tc.method, string(body), tc.expectedBody)
			}
		})
	}
}

func TestApplication_MiddlewareOrder(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	var called []string
	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "mw1")
			next(w, r)
		}
	}
	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "mw2")
			next(w, r)
		}
	}

	app.Get("/mw", func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "handler")
	}, mw1, mw2)

	req := httptest.NewRequest("GET", "/mw", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	expected := []string{"mw1", "mw2", "handler"}
	if len(called) != len(expected) {
		t.Fatalf("Expected %d middleware calls, got %d", len(expected), len(called))
	}
	for i, s := range expected {
		if called[i] != s {
			t.Errorf("Expected call order %v, got %v", expected, called)
		}
	}
}

func TestApplication_AllowedMethodsHeader(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)
	app.Get("/foo", stringHandler("bar"))
	app.Post("/foo", stringHandler("baz"))

	req := httptest.NewRequest("DELETE", "/foo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	resp := rr.Result()
	allow := resp.Header.Get("Allow")
	if !strings.Contains(allow, "GET") || !strings.Contains(allow, "POST") {
		t.Errorf("Allow header want to contain GET and POST, got: %s", allow)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", resp.StatusCode)
	}
}

func TestApplication_MultipleRoutes(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)
	app.Get("/route1", stringHandler("route1"))
	app.Get("/route2", stringHandler("route2"))

	tests := []struct {
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"/route1", http.StatusOK, "route1"},
		{"/route2", http.StatusOK, "route2"},
		{"/route3", http.StatusNotFound, "404 page not found\n"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("%s: got status %d, want %d", tc.path, resp.StatusCode, tc.expectedStatus)
			}
			if string(body) != tc.expectedBody {
				t.Errorf("%s: got body %q, want %q", tc.path, string(body), tc.expectedBody)
			}
		})
	}
}

func TestApplication_OverwriteHandler(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Get("/test", stringHandler("first"))
	app.Get("/test", stringHandler("second"))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	resp := rr.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "second" {
		t.Errorf("Expected overwritten handler to respond, got %q", string(body))
	}
}

func TestApplication_MiddlewaresApplication(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(r.Header.Get("X-Test")))
		if err != nil {
			panic(err)
		}
	}, func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Test", "middleware-applied")
			next(w, r)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	resp := rr.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "middleware-applied" {
		t.Errorf("Expected middleware to modify request, got %q", string(body))
	}
}

func TestApplication_GlobalMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	var called []string
	globalMw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "global")
			next(w, r)
		}
	}

	app.WithGlobalMiddlewares(globalMw)
	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "handler")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	expected := []string{"global", "handler"}
	if !reflect.DeepEqual(called, expected) {
		t.Errorf("Expected call order %v, got %v", expected, called)
	}
}

func BenchmarkApplication_Get(b *testing.B) {
	mux := http.NewServeMux()
	app := NewApplication(mux)
	app.Get("/bench", stringHandler("foo"))

	req := httptest.NewRequest("GET", "/bench", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

func BenchmarkApplication_WithMiddleware(b *testing.B) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	}

	app.Get("/bench", stringHandler("foo"), mw, mw, mw)

	req := httptest.NewRequest("GET", "/bench", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

func BenchmarkApplication_AllowedMethods(b *testing.B) {
	mux := http.NewServeMux()
	app := NewApplication(mux).(*Application)
	app.Get("/bench", stringHandler("foo"))
	app.Post("/bench", stringHandler("bar"))
	app.Put("/bench", stringHandler("baz"))
	app.Patch("/bench", stringHandler("qux"))
	app.Delete("/bench", stringHandler("quux"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		methods := app.allowedMethods("/bench")
		_ = methods
	}
}

func TestApplication_MiddlewareInterruption(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	var called []string
	interruptMw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "interrupt")
			w.WriteHeader(http.StatusUnauthorized)
			// Don't call next
		}
	}

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "handler")
	}, interruptMw)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if len(called) != 1 || called[0] != "interrupt" {
		t.Errorf("Expected only interrupt middleware to be called, got %v", called)
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestApplication_ResponseHeaders(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "test")
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type header to be set")
	}
	if rr.Header().Get("X-Custom") != "test" {
		t.Error("Expected X-Custom header to be set")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}
}

func TestApplication_ConcurrentRequests(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Get("/test", stringHandler("test"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}
			if rr.Body.String() != "test" {
				t.Errorf("Expected body 'test', got %q", rr.Body.String())
			}
		}()
	}
	wg.Wait()
}

func TestApplication_InvalidHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected a panic due to nil handler, but did not get one")
		}
	}()

	mux := http.NewServeMux()
	app := NewApplication(mux)

	// Test with nil handler
	app.Get("/test", nil)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}
