package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock controller for testing
type TestController struct {
	indexCalled  bool
	showCalled   bool
	newCalled    bool
	createCalled bool
	editCalled   bool
	updateCalled bool
	deleteCalled bool
}

func (tc *TestController) Index(c *Context) error {
	tc.indexCalled = true
	return c.String(http.StatusOK, "index")
}

func (tc *TestController) Show(c *Context) error {
	tc.showCalled = true
	return c.String(http.StatusOK, "show")
}

func (tc *TestController) New(c *Context) error {
	tc.newCalled = true
	return c.String(http.StatusOK, "new")
}

func (tc *TestController) Create(c *Context) error {
	tc.createCalled = true
	return c.String(http.StatusCreated, "create")
}

func (tc *TestController) Edit(c *Context) error {
	tc.editCalled = true
	return c.String(http.StatusOK, "edit")
}

func (tc *TestController) Update(c *Context) error {
	tc.updateCalled = true
	return c.String(http.StatusOK, "update")
}

func (tc *TestController) Delete(c *Context) error {
	tc.deleteCalled = true
	return c.String(http.StatusOK, "delete")
}

func TestResourcesFullController(t *testing.T) {
	r := New()
	controller := &TestController{}

	r.Resources("/users", controller)

	tests := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
		checkFlag    *bool
	}{
		{"GET", "/users", http.StatusOK, "index", &controller.indexCalled},
		{"GET", "/users/new", http.StatusOK, "new", &controller.newCalled},
		{"POST", "/users", http.StatusCreated, "create", &controller.createCalled},
		{"GET", "/users/123", http.StatusOK, "show", &controller.showCalled},
		{"GET", "/users/123/edit", http.StatusOK, "edit", &controller.editCalled},
		{"PATCH", "/users/123", http.StatusOK, "update", &controller.updateCalled},
		{"DELETE", "/users/123", http.StatusOK, "delete", &controller.deleteCalled},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			if tt.checkFlag != nil && !*tt.checkFlag {
				t.Errorf("handler method was not called")
			}
		})
	}
}

func TestResourcesWithOnly(t *testing.T) {
	r := New()
	controller := &TestController{}

	r.Resources("/posts", controller, Only(IndexAction, ShowAction))

	tests := []struct {
		method       string
		path         string
		expectedCode int
		shouldWork   bool
	}{
		{"GET", "/posts", http.StatusOK, true},
		{"GET", "/posts/123", http.StatusOK, true},
		// /posts/new will match /posts/:id with id="new" since NewAction is not registered
		{"GET", "/posts/new", http.StatusOK, true},
		{"POST", "/posts", http.StatusMethodNotAllowed, false},
		{"GET", "/posts/123/edit", http.StatusNotFound, false},
		{"PATCH", "/posts/123", http.StatusMethodNotAllowed, false},
		{"DELETE", "/posts/123", http.StatusMethodNotAllowed, false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestResourcesWithExcept(t *testing.T) {
	r := New()
	controller := &TestController{}

	r.Resources("/comments", controller, Except(NewAction, EditAction))

	tests := []struct {
		method       string
		path         string
		expectedCode int
		shouldWork   bool
	}{
		{"GET", "/comments", http.StatusOK, true},
		{"POST", "/comments", http.StatusCreated, true},
		{"GET", "/comments/123", http.StatusOK, true},
		{"PATCH", "/comments/123", http.StatusOK, true},
		{"DELETE", "/comments/123", http.StatusOK, true},
		// /comments/new will match /comments/:id with id="new" since NewAction is excluded
		{"GET", "/comments/new", http.StatusOK, true},
		{"GET", "/comments/123/edit", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestResourcesInGroup(t *testing.T) {
	r := New()
	controller := &TestController{}

	api := r.Group("/api/v1")
	api.Resources("/users", controller, Only(IndexAction, ShowAction))

	tests := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"GET", "/api/v1/users", http.StatusOK},
		{"GET", "/api/v1/users/123", http.StatusOK},
		{"POST", "/api/v1/users", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestResourcesWithMiddleware(t *testing.T) {
	r := New()
	controller := &TestController{}

	middlewareCalled := false
	middleware := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			middlewareCalled = true
			c.SetHeader("X-Middleware", "true")
			return next(c)
		}
	}

	r.Resources("/users", controller,
		Only(IndexAction),
		WithResourceMiddleware(middleware),
	)

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("middleware was not called")
	}

	if w.Header().Get("X-Middleware") != "true" {
		t.Error("middleware did not set header")
	}
}

func TestResourcesPutAlsoWorksForUpdate(t *testing.T) {
	r := New()
	controller := &TestController{}

	r.Resources("/users", controller)

	// Both PATCH and PUT should work for Update action
	tests := []struct {
		method string
	}{
		{"PATCH"},
		{"PUT"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			controller.updateCalled = false
			req := httptest.NewRequest(tt.method, "/users/123", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			if !controller.updateCalled {
				t.Error("Update was not called")
			}
		})
	}
}
