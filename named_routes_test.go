package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNamedRoutes(t *testing.T) {
	r := New()

	// Test with WithName option
	r.Get("/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("user_show"))

	r.Get("/posts/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("post_show"))

	// Test without name (unnamed route)
	r.Get("/about", func(c *Context) error {
		return c.String(http.StatusOK, "About")
	})

	// Verify routes are registered
	namedRoutes := r.NamedRoutes()
	if namedRoutes["user_show"] == nil {
		t.Error("user_show route not registered")
	}

	if namedRoutes["post_show"] == nil {
		t.Error("post_show route not registered")
	}

	// Verify route patterns
	if namedRoutes["user_show"].Pattern != "/users/:id" {
		t.Errorf("expected /users/:id, got %s", namedRoutes["user_show"].Pattern)
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

	// Use type-safe options for middleware + name
	r.Get("/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("user_show"), WithMiddleware(middleware))

	// Verify route works
	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !called {
		t.Error("middleware was not called")
	}

	// Verify route is registered
	namedRoutes := r.NamedRoutes()
	if namedRoutes["user_show"] == nil {
		t.Error("user_show route not registered")
	}
}

func TestNamedRoutesInGroups(t *testing.T) {
	r := New()
	api := r.Group("/api/v1")

	api.Get("/users/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("api_user_show"))

	api.Get("/posts/:id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
	}, WithName("api_post_show"))

	// Verify routes are registered with full path
	namedRoutes := r.NamedRoutes()
	if namedRoutes["api_user_show"] == nil {
		t.Error("api_user_show route not registered")
	}

	if namedRoutes["api_user_show"].Pattern != "/api/v1/users/:id" {
		t.Errorf("expected /api/v1/users/:id, got %s", namedRoutes["api_user_show"].Pattern)
	}

	if namedRoutes["api_post_show"].Pattern != "/api/v1/posts/:id" {
		t.Errorf("expected /api/v1/posts/:id, got %s", namedRoutes["api_post_show"].Pattern)
	}
}

func TestNamedRoutesMultipleParams(t *testing.T) {
	r := New()

	r.Get("/users/:user_id/posts/:post_id", func(c *Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"user_id": c.Param("user_id"),
			"post_id": c.Param("post_id"),
		})
	}, WithName("user_post"))

	// Verify route is registered
	namedRoutes := r.NamedRoutes()
	if namedRoutes["user_post"] == nil {
		t.Error("user_post route not registered")
	}

	if namedRoutes["user_post"].Pattern != "/users/:user_id/posts/:post_id" {
		t.Errorf("expected /users/:user_id/posts/:post_id, got %s", namedRoutes["user_post"].Pattern)
	}
}

func TestNamedRoutesAllMethods(t *testing.T) {
	r := New()

	r.Get("/items/:id", func(c *Context) error { return nil }, WithName("item_show"))
	r.Post("/items", func(c *Context) error { return nil }, WithName("item_create"))
	r.Put("/items/:id", func(c *Context) error { return nil }, WithName("item_update"))
	r.Patch("/items/:id", func(c *Context) error { return nil }, WithName("item_patch"))
	r.Delete("/items/:id", func(c *Context) error { return nil }, WithName("item_delete"))
	r.Head("/items/:id", func(c *Context) error { return nil }, WithName("item_head"))
	r.Options("/items/:id", func(c *Context) error { return nil }, WithName("item_options"))

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
			namedRoutes := r.NamedRoutes()
			route := namedRoutes[tt.name]
			if route == nil {
				t.Fatalf("route %s not registered", tt.name)
			}
			if route.Pattern != tt.expected {
				t.Errorf("expected pattern %s, got %s", tt.expected, route.Pattern)
			}
			if route.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, route.Method)
			}
		})
	}
}
