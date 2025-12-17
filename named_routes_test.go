package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNamedRoutes(t *testing.T) {
	r := New()

	// Test with name as string
	r.Get("/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, "user_show")

	// Test with WithName option
	r.Get("/posts/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("post_show"))

	// Test without name (unnamed route)
	r.Get("/about", func(c *Context) error {
		return c.String(http.StatusOK, "About")
	})

	// Verify routes are registered
	if r.namedRoutes["user_show"] == nil {
		t.Error("user_show route not registered")
	}

	if r.namedRoutes["post_show"] == nil {
		t.Error("post_show route not registered")
	}

	// Verify route patterns
	if r.namedRoutes["user_show"].pattern != "/users/:id" {
		t.Errorf("expected /users/:id, got %s", r.namedRoutes["user_show"].pattern)
	}
}

func TestNamedRoutesWithMiddleware(t *testing.T) {
	r := New()

	called := false
	middleware := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			called = true
			return next(c)
		}
	}

	// Use HandleNamed directly for explicit middleware + name
	r.HandleNamed("GET", "/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, "user_show", middleware)

	// Verify route works
	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !called {
		t.Error("middleware was not called")
	}

	// Verify route is registered
	if r.namedRoutes["user_show"] == nil {
		t.Error("user_show route not registered")
	}
}

func TestNamedRoutesInGroups(t *testing.T) {
	r := New()
	api := r.Group("/api/v1")

	api.Get("/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, "api_user_show")

	api.Get("/posts/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("api_post_show"))

	// Verify routes are registered with full path
	if r.namedRoutes["api_user_show"] == nil {
		t.Error("api_user_show route not registered")
	}

	if r.namedRoutes["api_user_show"].pattern != "/api/v1/users/:id" {
		t.Errorf("expected /api/v1/users/:id, got %s", r.namedRoutes["api_user_show"].pattern)
	}

	if r.namedRoutes["api_post_show"].pattern != "/api/v1/posts/:id" {
		t.Errorf("expected /api/v1/posts/:id, got %s", r.namedRoutes["api_post_show"].pattern)
	}
}

func TestNamedRoutesMultipleParams(t *testing.T) {
	r := New()

	r.Get("/users/:user_id/posts/:post_id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"user_id": c.Param("user_id"),
			"post_id": c.Param("post_id"),
		})
	}, "user_post")

	// Verify route is registered
	if r.namedRoutes["user_post"] == nil {
		t.Error("user_post route not registered")
	}

	if r.namedRoutes["user_post"].pattern != "/users/:user_id/posts/:post_id" {
		t.Errorf("expected /users/:user_id/posts/:post_id, got %s", r.namedRoutes["user_post"].pattern)
	}
}

func TestNamedRoutesAllMethods(t *testing.T) {
	r := New()

	r.Get("/items/:id", func(c *Context) error { return nil }, "item_show")
	r.Post("/items", func(c *Context) error { return nil }, "item_create")
	r.Put("/items/:id", func(c *Context) error { return nil }, "item_update")
	r.Patch("/items/:id", func(c *Context) error { return nil }, "item_patch")
	r.Delete("/items/:id", func(c *Context) error { return nil }, "item_delete")
	r.Head("/items/:id", func(c *Context) error { return nil }, "item_head")
	r.Options("/items/:id", func(c *Context) error { return nil }, "item_options")

	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"item_show", "GET", "/items/:id"},
		{"item_create", "POST", "/items"},
		{"item_update", "PUT", "/items/:id"},
		{"item_patch", "PATCH", "/items/:id"},
		{"item_delete", "DELETE", "/items/:id"},
		{"item_head", "HEAD", "/items/:id"},
		{"item_options", "OPTIONS", "/items/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := r.namedRoutes[tt.name]
			if route == nil {
				t.Fatalf("route %s not registered", tt.name)
			}
			if route.pattern != tt.expected {
				t.Errorf("expected pattern %s, got %s", tt.expected, route.pattern)
			}
			if route.method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, route.method)
			}
		})
	}
}
