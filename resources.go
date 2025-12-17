package router

import "strings"

// Controller defines the interface for RESTful resource controllers
// Controllers can implement any subset of these methods
type Controller interface{}

// ResourceController defines all possible RESTful actions
type ResourceController interface {
	Index(c *Context) error  // GET    /resources
	New(c *Context) error    // GET    /resources/new
	Create(c *Context) error // POST   /resources
	Show(c *Context) error   // GET    /resources/:id
	Edit(c *Context) error   // GET    /resources/:id/edit
	Update(c *Context) error // PATCH  /resources/:id
	Delete(c *Context) error // DELETE /resources/:id
}

// ResourceAction represents the available RESTful actions
type ResourceAction string

const (
	IndexAction  ResourceAction = "index"
	NewAction    ResourceAction = "new"
	CreateAction ResourceAction = "create"
	ShowAction   ResourceAction = "show"
	EditAction   ResourceAction = "edit"
	UpdateAction ResourceAction = "update"
	DeleteAction ResourceAction = "delete"
)

// AllResourceActions contains all available resource actions
var AllResourceActions = []ResourceAction{
	IndexAction,
	NewAction,
	CreateAction,
	ShowAction,
	EditAction,
	UpdateAction,
	DeleteAction,
}

// ResourceOptions configures which actions to generate for a resource
type ResourceOptions struct {
	// Only specifies which actions to include (whitelist)
	Only []ResourceAction

	// Except specifies which actions to exclude (blacklist)
	Except []ResourceAction

	// Middleware to apply to all resource routes
	Middleware []MiddlewareFunc
}

// ResourceOption is a functional option for configuring resources
type ResourceOption func(*ResourceOptions)

// Only limits the resource to only the specified actions
func Only(actions ...ResourceAction) ResourceOption {
	return func(opts *ResourceOptions) {
		opts.Only = actions
	}
}

// Except excludes the specified actions from the resource
func Except(actions ...ResourceAction) ResourceOption {
	return func(opts *ResourceOptions) {
		opts.Except = actions
	}
}

// WithMiddleware adds middleware to all resource routes
func WithMiddleware(middleware ...MiddlewareFunc) ResourceOption {
	return func(opts *ResourceOptions) {
		opts.Middleware = middleware
	}
}

// shouldIncludeAction determines if an action should be included based on Only/Except options
func (opts *ResourceOptions) shouldIncludeAction(action ResourceAction) bool {
	// If Only is specified, action must be in the list
	if len(opts.Only) > 0 {
		for _, a := range opts.Only {
			if a == action {
				return true
			}
		}
		return false
	}

	// If Except is specified, action must NOT be in the list
	if len(opts.Except) > 0 {
		for _, a := range opts.Except {
			if a == action {
				return false
			}
		}
		return true
	}

	// Default: include all actions
	return true
}

// actionRoute defines the HTTP method and path for each action
type actionRoute struct {
	method string
	path   string
	action ResourceAction
}

// getResourceRoutes returns the route definitions for RESTful resources
// Order matters! Static routes (/new, /:id/edit) must come before dynamic routes (/:id)
func getResourceRoutes(basePath string) []actionRoute {
	return []actionRoute{
		{"GET", basePath, IndexAction},
		{"GET", basePath + "/new", NewAction}, // Must be before /:id
		{"POST", basePath, CreateAction},
		{"GET", basePath + "/:id/edit", EditAction}, // Must be before /:id
		{"GET", basePath + "/:id", ShowAction},
		{"PATCH", basePath + "/:id", UpdateAction},
		{"PUT", basePath + "/:id", UpdateAction}, // Also accept PUT for Update
		{"DELETE", basePath + "/:id", DeleteAction},
	}
}

// Resources registers RESTful routes for a controller
// Example:
//
//	r.Resources("/users", &UserController{})
//	r.Resources("/posts", &PostController{}, Only(IndexAction, ShowAction))
//	r.Resources("/comments", &CommentController{}, Except(NewAction, EditAction))
func (r *Router) Resources(path string, controller Controller, opts ...ResourceOption) {
	options := &ResourceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// If no Only/Except options are provided, validate that all methods are implemented
	requireAll := len(options.Only) == 0 && len(options.Except) == 0

	// Extract resource name from path (e.g., "/todos" -> "todos", "/api/v1/users" -> "users")
	resourceName := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		resourceName = path[idx+1:]
	}

	routes := getResourceRoutes(path)

	for _, route := range routes {
		if !options.shouldIncludeAction(route.action) {
			continue
		}

		handler := getControllerHandler(controller, route.action)
		if handler == nil {
			if requireAll {
				panic("controller must implement all ResourceController methods when using Resources() without Only() or Except() options. Missing: " + string(route.action))
			}
			continue
		}

		// Generate route name like "todos_index", "todos_show", etc.
		routeName := resourceName + "_" + string(route.action)
		r.HandleNamed(route.method, route.path, handler, routeName, options.Middleware...)
	}
}

// getControllerHandler extracts the appropriate handler method from a controller
func getControllerHandler(controller Controller, action ResourceAction) HandlerFunc {
	switch action {
	case IndexAction:
		if c, ok := controller.(interface {
			Index(*Context) error
		}); ok {
			return c.Index
		}
	case NewAction:
		if c, ok := controller.(interface {
			New(*Context) error
		}); ok {
			return c.New
		}
	case CreateAction:
		if c, ok := controller.(interface {
			Create(*Context) error
		}); ok {
			return c.Create
		}
	case ShowAction:
		if c, ok := controller.(interface {
			Show(*Context) error
		}); ok {
			return c.Show
		}
	case EditAction:
		if c, ok := controller.(interface {
			Edit(*Context) error
		}); ok {
			return c.Edit
		}
	case UpdateAction:
		if c, ok := controller.(interface {
			Update(*Context) error
		}); ok {
			return c.Update
		}
	case DeleteAction:
		if c, ok := controller.(interface {
			Delete(*Context) error
		}); ok {
			return c.Delete
		}
	}
	return nil
}
