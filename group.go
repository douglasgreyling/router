package router

import "strings"

// Group represents a group of routes with a common prefix and middleware
type Group struct {
	router     *Router
	prefix     string
	middleware []MiddlewareFunc
}

// Group creates a new route group with the given prefix
func (r *Router) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	return &Group{
		router:     r,
		prefix:     prefix,
		middleware: middleware,
	}
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// Handle registers a route with the group's prefix and middleware
func (g *Group) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	g.HandleNamed(method, path, handler, "", middleware...)
}

// HandleNamed registers a named route with the group's prefix and middleware
func (g *Group) HandleNamed(method, path string, handler HandlerFunc, name string, middleware ...MiddlewareFunc) {
	fullPath := g.prefix + path

	// Combine group middleware with route-specific middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	g.router.HandleNamed(method, fullPath, handler, name, allMiddleware...)
}

// HTTP method helpers for groups with optional naming
func (g *Group) Get(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("GET", path, handler, name, middleware...)
}

func (g *Group) Post(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("POST", path, handler, name, middleware...)
}

func (g *Group) Put(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("PUT", path, handler, name, middleware...)
}

func (g *Group) Patch(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("PATCH", path, handler, name, middleware...)
}

func (g *Group) Delete(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("DELETE", path, handler, name, middleware...)
}

func (g *Group) Head(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("HEAD", path, handler, name, middleware...)
}

func (g *Group) Options(path string, handler HandlerFunc, opts ...interface{}) {
	name, middleware := parseRouteOptions(opts)
	g.HandleNamed("OPTIONS", path, handler, name, middleware...)
}

// Group creates a nested group with combined prefix and middleware
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	// Combine parent and child middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	return &Group{
		router:     g.router,
		prefix:     g.prefix + prefix,
		middleware: allMiddleware,
	}
}

// Resources registers RESTful routes for a controller within the group
// Example:
//
//	api := r.Group("/api/v1")
//	api.Resources("/users", &UserController{})
//	api.Resources("/posts", &PostController{}, Only(IndexAction, ShowAction))
func (g *Group) Resources(path string, controller Controller, opts ...ResourceOption) {
	options := &ResourceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Combine group middleware with resource middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(options.Middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, options.Middleware...)
	options.Middleware = allMiddleware

	// Add the group prefix to the path
	fullPath := g.prefix + path

	// Extract resource name from path (e.g., "/users" -> "users")
	resourceName := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		resourceName = path[idx+1:]
	}

	routes := getResourceRoutes(fullPath)

	for _, route := range routes {
		if !options.shouldIncludeAction(route.action) {
			continue
		}

		handler := getControllerHandler(controller, route.action)
		if handler != nil {
			// Generate route name like "users_index", "users_show", etc.
			routeName := resourceName + "_" + string(route.action)
			g.router.HandleNamed(route.method, route.path, handler, routeName, options.Middleware...)
		}
	}
}
