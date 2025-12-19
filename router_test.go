package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
			if !strings.Contains(msg, "duplicate parameter name 'id'") {
				t.Errorf("Expected panic message about 'id', got: %s", msg)
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
			if !strings.Contains(msg, "duplicate parameter name 'path'") {
				t.Errorf("Expected panic message about 'path', got: %s", msg)
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
