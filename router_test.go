package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestStaticRoutes(t *testing.T) {
	r := New()

	handler := func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	}

	r.Get("/", handler)
	r.Get("/users", handler)
	r.Get("/about", handler)

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/users", http.StatusOK},
		{"GET", "/about", http.StatusOK},
		{"GET", "/notfound", http.StatusNotFound},
		{"POST", "/", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != tt.want {
			t.Errorf("%s %s: got %d, want %d", tt.method, tt.path, w.Code, tt.want)
		}
	}
}

func TestRouteParameters(t *testing.T) {
	r := New()

	var capturedID string
	r.Get("/users/:id", func(c *Context) error {
		capturedID = c.Param("id")
		return c.String(http.StatusOK, capturedID)
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedID != "123" {
		t.Errorf("Expected param 'id' to be '123', got '%s'", capturedID)
	}

	if w.Body.String() != "123" {
		t.Errorf("Expected body '123', got '%s'", w.Body.String())
	}
}

func TestMultipleParameters(t *testing.T) {
	r := New()

	var capturedUser, capturedPost string
	r.Get("/users/:user/posts/:post", func(c *Context) error {
		capturedUser = c.Param("user")
		capturedPost = c.Param("post")
		return c.String(http.StatusOK, capturedUser+":"+capturedPost)
	})

	req := httptest.NewRequest("GET", "/users/john/posts/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedUser != "john" {
		t.Errorf("Expected param 'user' to be 'john', got '%s'", capturedUser)
	}

	if capturedPost != "hello" {
		t.Errorf("Expected param 'post' to be 'hello', got '%s'", capturedPost)
	}
}

func TestWildcards(t *testing.T) {
	r := New()

	var capturedPath string
	r.Get("/files/*filepath", func(c *Context) error {
		capturedPath = c.Param("filepath")
		return c.String(http.StatusOK, capturedPath)
	})

	tests := []struct {
		path string
		want string
	}{
		{"/files/docs/readme.md", "docs/readme.md"},
		{"/files/src/main.go", "src/main.go"},
		{"/files/a/b/c/d/e.txt", "a/b/c/d/e.txt"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", tt.path, w.Code)
		}

		if capturedPath != tt.want {
			t.Errorf("Expected filepath '%s', got '%s'", tt.want, capturedPath)
		}
	}
}

func TestMiddleware(t *testing.T) {
	r := New()

	var calls []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			calls = append(calls, "middleware1-before")
			err := next(c)
			calls = append(calls, "middleware1-after")
			return err
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			calls = append(calls, "middleware2-before")
			err := next(c)
			calls = append(calls, "middleware2-after")
			return err
		}
	}

	handler := func(c *Context) error {
		calls = append(calls, "handler")
		return c.String(http.StatusOK, "OK")
	}

	r.Use(middleware1, middleware2)
	r.Get("/test", handler)

	calls = []string{}
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	if len(calls) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(calls))
	}

	for i, call := range expected {
		if i >= len(calls) || calls[i] != call {
			t.Errorf("Call %d: expected '%s', got '%s'", i, call, calls[i])
		}
	}
}

func TestRouteGroups(t *testing.T) {
	r := New()

	api := r.Group("/api")
	api.Get("/users", func(c *Context) error {
		return c.String(http.StatusOK, "users")
	})
	api.Get("/posts", func(c *Context) error {
		return c.String(http.StatusOK, "posts")
	})

	tests := []struct {
		path string
		want string
	}{
		{"/api/users", "users"},
		{"/api/posts", "posts"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", tt.path, w.Code)
		}

		if w.Body.String() != tt.want {
			t.Errorf("Expected body '%s', got '%s'", tt.want, w.Body.String())
		}
	}
}

func TestNestedGroups(t *testing.T) {
	r := New()

	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/users", func(c *Context) error {
		return c.String(http.StatusOK, "v1-users")
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "v1-users" {
		t.Errorf("Expected body 'v1-users', got '%s'", w.Body.String())
	}
}

func TestGroupMiddleware(t *testing.T) {
	r := New()

	var calls []string

	groupMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			calls = append(calls, "group-middleware")
			return next(c)
		}
	}

	api := r.Group("/api")
	api.Use(groupMiddleware)
	api.Get("/test", func(c *Context) error {
		calls = append(calls, "handler")
		return c.String(http.StatusOK, "OK")
	})

	calls = []string{}
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expected := []string{"group-middleware", "handler"}

	if len(calls) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(calls))
	}

	for i, call := range expected {
		if i >= len(calls) || calls[i] != call {
			t.Errorf("Call %d: expected '%s', got '%s'", i, call, calls[i])
		}
	}
}

func TestAllHTTPMethods(t *testing.T) {
	r := New()

	handler := func(method string) HandlerFunc {
		return func(c *Context) error {
			return c.String(http.StatusOK, method)
		}
	}

	r.Get("/test", handler("GET"))
	r.Post("/test", handler("POST"))
	r.Put("/test", handler("PUT"))
	r.Delete("/test", handler("DELETE"))
	r.Patch("/test", handler("PATCH"))
	r.Options("/test", handler("OPTIONS"))
	r.Head("/test", handler("HEAD"))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("%s /test: expected status 200, got %d", method, w.Code)
		}

		if method != "HEAD" && w.Body.String() != method {
			t.Errorf("%s /test: expected body '%s', got '%s'", method, method, w.Body.String())
		}
	}
}

func TestRoutePriority(t *testing.T) {
	r := New()

	// Static routes should have priority over parameterized routes
	r.Get("/users/new", func(c *Context) error {
		return c.String(http.StatusOK, "new")
	})
	r.Get("/users/:id", func(c *Context) error {
		return c.String(http.StatusOK, "param")
	})

	req := httptest.NewRequest("GET", "/users/new", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "new" {
		t.Errorf("Expected 'new', got '%s'", w.Body.String())
	}

	req = httptest.NewRequest("GET", "/users/123", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "param" {
		t.Errorf("Expected 'param', got '%s'", w.Body.String())
	}
}

func TestDuplicateParameterNames(t *testing.T) {
	r := New()

	// This should panic because :id is used twice
	defer func() {
		if rec := recover(); rec == nil {
			t.Error("Expected panic for duplicate parameter names")
		} else {
			msg := fmt.Sprint(rec)
			if !strings.Contains(msg, "duplicate parameter") || !strings.Contains(msg, "id") {
				t.Errorf("Expected panic message about duplicate parameter 'id', got: %s", msg)
			}
		}
	}()

	r.Get("/users/:id/posts/:id", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})
}

func TestDuplicateParameterNamesWithWildcard(t *testing.T) {
	r := New()

	// This should panic because :path and *path are both named "path"
	defer func() {
		if rec := recover(); rec == nil {
			t.Error("Expected panic for duplicate parameter names")
		} else {
			msg := fmt.Sprint(rec)
			if !strings.Contains(msg, "duplicate parameter") || !strings.Contains(msg, "path") {
				t.Errorf("Expected panic message about duplicate parameter 'path', got: %s", msg)
			}
		}
	}()

	r.Get("/files/:path/*path", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})
}

func TestUniqueParameterNames(t *testing.T) {
	r := New()

	// This should NOT panic - different parameter names
	r.Get("/users/:user_id/posts/:post_id", func(c *Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("%s:%s", c.Param("user_id"), c.Param("post_id")))
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "123:456" {
		t.Errorf("Expected '123:456', got '%s'", w.Body.String())
	}
}

func TestErrorHandlerAfterHeadersWritten(t *testing.T) {
	r := New()

	// Suppress stderr for this test since we expect the error log
	r.ErrorHandler = func(c *Context, err error) {
		if !c.IsHeaderWritten() {
			c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		// Silently ignore - we're testing this behavior
	}

	r.Get("/test", func(c *Context) error {
		// Write headers and part of response
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte("Success"))

		// Now return an error
		return fmt.Errorf("error after headers sent")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should have the original response, not error JSON
	if w.Body.String() != "Success" {
		t.Errorf("Expected 'Success', got '%s'", w.Body.String())
	}

	// Status should be OK, not 500
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestErrorHandlerBeforeHeadersWritten(t *testing.T) {
	r := New()

	r.Get("/test", func(c *Context) error {
		// Return error before writing anything
		return fmt.Errorf("early error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should have error JSON response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "early error") {
		t.Errorf("Expected error message in body, got '%s'", w.Body.String())
	}
}

func TestIsHeaderWritten(t *testing.T) {
	r := New()

	var headerWrittenBefore, headerWrittenAfter bool

	r.Get("/test", func(c *Context) error {
		headerWrittenBefore = c.IsHeaderWritten()
		c.JSON(http.StatusOK, map[string]string{"message": "hello"})
		headerWrittenAfter = c.IsHeaderWritten()
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if headerWrittenBefore {
		t.Error("Headers should not be written before calling JSON")
	}

	if !headerWrittenAfter {
		t.Error("Headers should be written after calling JSON")
	}
}

func TestGetStatus(t *testing.T) {
	r := New()

	var statusBefore, statusAfter int

	r.Get("/test", func(c *Context) error {
		statusBefore = c.GetStatus()
		c.Writer.WriteHeader(http.StatusCreated)
		statusAfter = c.GetStatus()
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if statusBefore != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", statusBefore)
	}

	if statusAfter != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", statusAfter)
	}
}

func TestConcurrentRequests(t *testing.T) {
	r := New()

	callCount := 0
	var mu sync.Mutex

	r.Get("/test", func(c *Context) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return c.String(http.StatusOK, "OK")
	})

	r.Get("/params/:id", func(c *Context) error {
		id := c.Param("id")
		return c.String(http.StatusOK, id)
	})

	const numGoroutines = 100
	var wg sync.WaitGroup

	// Test static routes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		}()
	}

	// Test parameterized routes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := fmt.Sprintf("/params/%d", id)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		}(i)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if callCount != numGoroutines {
		t.Errorf("Expected %d calls, got %d", numGoroutines, callCount)
	}
}

// Benchmarks

func BenchmarkStaticRoutes(b *testing.B) {
	r := New()
	r.Get("/users", func(c *Context) error {
		return c.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest("GET", "/users", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkParameterRoutes(b *testing.B) {
	r := New()
	r.Get("/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.String(http.StatusOK, id)
	})

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkWildcardRoutes(b *testing.B) {
	r := New()
	r.Get("/files/*filepath", func(c *Context) error {
		path := c.Param("filepath")
		return c.String(http.StatusOK, path)
	})

	req := httptest.NewRequest("GET", "/files/some/deep/path/file.txt", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkComplexRouting(b *testing.B) {
	r := New()

	// Multiple routes to simulate realistic routing
	r.Get("/", func(c *Context) error { return c.String(http.StatusOK, "home") })
	r.Get("/about", func(c *Context) error { return c.String(http.StatusOK, "about") })
	r.Get("/contact", func(c *Context) error { return c.String(http.StatusOK, "contact") })
	r.Get("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })
	r.Get("/users/:id", func(c *Context) error { return c.String(http.StatusOK, "user") })
	r.Get("/users/:id/posts", func(c *Context) error { return c.String(http.StatusOK, "posts") })
	r.Get("/users/:id/posts/:post_id", func(c *Context) error { return c.String(http.StatusOK, "post") })
	r.Get("/files/*filepath", func(c *Context) error { return c.String(http.StatusOK, "file") })

	req := httptest.NewRequest("GET", "/users/42/posts/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	r := New()

	// Add multiple middleware
	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return next(c)
		}
	}
	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return next(c)
		}
	}
	middleware3 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return next(c)
		}
	}

	r.Use(middleware1, middleware2, middleware3)

	r.Get("/test", func(c *Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

