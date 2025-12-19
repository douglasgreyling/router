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

// handle registers a route with the group's prefix and middleware.
// This is an internal method. Use HTTP method helpers (Get, Post, etc.) instead.
func (g *Group) handle(method, path string, handler HandlerFunc, name string, middleware ...MiddlewareFunc) {
	fullPath := g.prefix + path

	// Combine group middleware with route-specific middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	g.router.handle(method, fullPath, handler, name, allMiddleware...)
}

// HTTP method helpers for groups with type-safe options
func (g *Group) Get(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("GET", path, handler, name, middleware...)
}

func (g *Group) Post(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("POST", path, handler, name, middleware...)
}

func (g *Group) Put(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("PUT", path, handler, name, middleware...)
}

func (g *Group) Patch(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("PATCH", path, handler, name, middleware...)
}

func (g *Group) Delete(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("DELETE", path, handler, name, middleware...)
}

func (g *Group) Head(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("HEAD", path, handler, name, middleware...)
}

func (g *Group) Options(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	g.handle("OPTIONS", path, handler, name, middleware...)
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
	config := parseResourceOptions(opts)

	// Combine group middleware with resource middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(config.middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, config.middleware...)
	config.middleware = allMiddleware

	// Add the group prefix to the path
	fullPath := g.prefix + path

	// Extract resource name from path (e.g., "/users" -> "users")
	resourceName := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		resourceName = path[idx+1:]
	}

	routes := getResourceRoutes(fullPath)

	for _, route := range routes {
		if !config.shouldIncludeAction(route.action) {
			continue
		}

		handler := getControllerHandler(controller, route.action)
		if handler != nil {
			// Generate route name like "users_index", "users_show", etc.
			routeName := resourceName + "_" + string(route.action)
			g.router.handle(route.method, route.path, handler, routeName, config.middleware...)
		}
	}
}
