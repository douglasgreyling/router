// Package router provides a fast, flexible HTTP router with automatic
// route helper generation, RESTful resource scaffolding, and type-safe
// middleware composition.
//
// Basic Usage:
//
//	r := router.New()
//
//	r.Get("/users/:id", func(c *router.Context) error {
//	    id := c.Param("id")
//	    return c.JSON(200, map[string]string{"id": id})
//	})
//
//	r.Serve() // Starts server on :3000 with route helper generation
//
// Features:
//
//   - Fast radix tree routing with parameter and wildcard support
//   - Automatic code generation of type-safe route helpers
//   - RESTful resource scaffolding (Rails-inspired)
//   - Flexible middleware with group and route-level composition
//   - Rich Context API for request/response handling
//
// Route Parameters:
//
// The router supports two types of dynamic segments:
//
//   - Named parameters (:param) match a single path segment
//   - Wildcards (*wildcard) match everything after the prefix
//
// Example:
//
//	r.Get("/users/:id", handler)           // Matches: /users/123
//	r.Get("/files/*filepath", handler)     // Matches: /files/docs/readme.txt
//
// Middleware:
//
// Middleware can be applied at the router, group, or route level.
// Middleware is executed in order: global → group → route-specific.
//
//	r.Use(loggingMiddleware)  // Applied to all routes (executes first)
//
//	api := r.Group("/api", authMiddleware)  // Applied to group (executes second)
//	api.Get("/users", handler)
//
//	r.Get("/public", handler, WithMiddleware(cacheMiddleware))  // Route-specific (executes last)
//
// Error Handling:
//
// Handlers return errors, which are processed by the ErrorHandler.
// The default ErrorHandler sends a JSON error response:
//
//	r.Get("/users/:id", func(c *Context) error {
//	    user, err := findUser(c.Param("id"))
//	    if err != nil {
//	        return err  // ErrorHandler will process this
//	    }
//	    return c.JSON(200, user)
//	})
//
// Customize error handling:
//
//	r.ErrorHandler = func(c *Context, err error) {
//	    if c.IsHeaderWritten() {
//	        log.Printf("Error after headers sent: %v", err)
//	        return
//	    }
//	    c.JSON(500, map[string]string{"error": err.Error()})
//	}
//
// Generated Route Helpers:
//
// The router automatically generates type-safe route helpers in development mode.
// For a route named "users_show", you can generate URLs like:
//
//	import "yourapp/routes"
//	url := routes.UsersShowPath(id)  // Generates: /users/:id
//
// Control generation:
//
//	r.Serve()                           // Auto-generate in development
//	r.Serve(WithGenerateHelpers(false)) // Disable generation
//	r.Serve(WithRoutesPackage("api"), WithRoutesOutputFile("api/routes.go"))
//
// RESTful Resources:
//
// The router provides Rails-inspired resource scaffolding:
//
//	type PostController struct{}
//
//	func (pc *PostController) Index(c *router.Context) error {
//	    return c.JSON(200, posts)
//	}
//
//	func (pc *PostController) Show(c *router.Context) error {
//	    id := c.Param("id")
//	    return c.JSON(200, findPost(id))
//	}
//
//	r.Resources("/posts", &PostController{},
//	    Only(IndexAction, ShowAction),
//	)
//
// For more examples and documentation, see: https://github.com/douglasgreyling/router
package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/douglasgreyling/router/internal/naming"
	"github.com/douglasgreyling/router/internal/tree"
	"github.com/douglasgreyling/router/routehelper"
)

// HandlerFunc is the function signature for route handlers.
// Handlers receive a Context and return an error. If an error is returned,
// it will be processed by the router's ErrorHandler.
//
// Example:
//
//	func myHandler(c *router.Context) error {
//	    data := map[string]string{"message": "hello"}
//	    return c.JSON(200, data)
//	}
type HandlerFunc func(*Context) error

// Params holds route parameters extracted from the URL.
// Parameters are defined in routes using :name syntax.
//
// Example:
//
//	r.Get("/users/:id/posts/:post_id", func(c *Context) error {
//	    userID := c.Param("id")        // Access via Context.Param()
//	    postID := c.Param("post_id")
//	    return c.String(200, "User: %s, Post: %s", userID, postID)
//	})
type Params map[string]string

// Router is the main router structure
type Router struct {
	// Route tree for fast lookups
	tree *tree.Tree

	// Named routes registry
	names *naming.Registry

	// Global middleware applied to all routes
	middleware []MiddlewareFunc

	// NotFound handler
	NotFound HandlerFunc

	// MethodNotAllowed handler
	MethodNotAllowed HandlerFunc

	// ErrorHandler handles errors returned from handlers
	ErrorHandler func(*Context, error)
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		tree:  tree.New(),
		names: naming.NewRegistry(),
		NotFound: func(c *Context) error {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Not Found",
			})
		},
		MethodNotAllowed: func(c *Context) error {
			return c.JSON(http.StatusMethodNotAllowed, map[string]string{
				"error": "Method Not Allowed",
			})
		},
		ErrorHandler: func(c *Context, err error) {
			// Can't modify response if headers already sent
			if c.IsHeaderWritten() {
				// Log error since we can't send proper error response
				fmt.Fprintf(os.Stderr, "Error after headers sent: %v\n", err)
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		},
	}
}

// Use adds global middleware to the router
func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

// handle registers a new route with the given method and path.
// This is an internal method called by HTTP method helpers (Get, Post, etc.).
// A route name is automatically generated if not provided.
//
// Panics if:
//   - path does not begin with '/'
//   - path contains duplicate parameter names (e.g., /users/:id/posts/:id)
func (r *Router) handle(method, path string, handler HandlerFunc, name string, middleware ...MiddlewareFunc) {
	// Convert middleware to interface{} slice for tree package
	mw := make([]interface{}, len(middleware))
	for i, m := range middleware {
		mw[i] = m
	}

	// Add route to tree
	if err := r.tree.AddRoute(method, path, handler, mw); err != nil {
		panic(err.Error())
	}

	// Auto-generate route name if not provided
	if name == "" {
		name = naming.GenerateName(path, method)
	}

	// Register named route
	if name != "" {
		r.names.Add(name, path, method)
	}
}

// Get registers a GET route with optional configuration.
//
// Options can be provided using WithName() and WithMiddleware():
//
//	r.Get("/users/:id", handler)
//	r.Get("/users/:id", handler, WithName("user_show"))
//	r.Get("/users/:id", handler, WithMiddleware(auth, logging))
//	r.Get("/users/:id", handler, WithName("user_show"), WithMiddleware(auth))
//
// Panics on invalid paths (see handle for details).
func (r *Router) Get(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("GET", path, handler, name, middleware...)
}

// Post registers a POST route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Post(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("POST", path, handler, name, middleware...)
}

// Put registers a PUT route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Put(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("PUT", path, handler, name, middleware...)
}

// Patch registers a PATCH route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Patch(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("PATCH", path, handler, name, middleware...)
}

// Delete registers a DELETE route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Delete(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("DELETE", path, handler, name, middleware...)
}

// Head registers a HEAD route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Head(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("HEAD", path, handler, name, middleware...)
}

// Options registers an OPTIONS route with optional configuration.
// See Get() for usage examples.
// Panics on invalid paths (see handle for details).
func (r *Router) Options(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.handle("OPTIONS", path, handler, name, middleware...)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	// Create context
	c := newContext(w, req)

	// Find the matching route
	handler, params, middlewareList := r.tree.Find(method, path)

	if handler == nil {
		// Check if route exists for a different method
		if r.tree.HasMethod(path) {
			if err := r.MethodNotAllowed(c); err != nil && r.ErrorHandler != nil {
				r.ErrorHandler(c, err)
			}
			return
		}

		if err := r.NotFound(c); err != nil && r.ErrorHandler != nil {
			r.ErrorHandler(c, err)
		}
		return
	}

	// Set params on context
	c.Params = params

	// Convert handler from interface{}
	h := handler.(HandlerFunc)

	// Convert middleware from []interface{}
	routeMiddleware := make([]MiddlewareFunc, len(middlewareList))
	for i, mw := range middlewareList {
		routeMiddleware[i] = mw.(MiddlewareFunc)
	}

	// Build middleware chain (global + route-specific)
	finalHandler := h

	// Apply route-specific middleware first (innermost)
	for i := len(routeMiddleware) - 1; i >= 0; i-- {
		finalHandler = routeMiddleware[i](finalHandler)
	}

	// Apply global middleware (outermost)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}

	// Execute the handler and handle any errors
	if err := finalHandler(c); err != nil && r.ErrorHandler != nil {
		r.ErrorHandler(c, err)
	}
}

// GenerateRoutes generates type-safe route helpers
func (r *Router) GenerateRoutes(packageName, outputFile string) error {
	rh := routehelper.New()

	// Get all named routes
	namedRoutes := r.names.All()

	// print out all named routes
	fmt.Printf("Generating route helpers for %d named routes...\n", len(namedRoutes))

	for name, route := range namedRoutes {
		rh.AddRoute(name, route.Pattern, route.Method)
	}
	return rh.Generate(packageName, outputFile)
}

// NamedRoutes returns all named routes (useful for testing and introspection)
func (r *Router) NamedRoutes() map[string]*naming.Route {
	return r.names.All()
}

// ServeConfig holds configuration for the Serve method
type ServeConfig struct {
	Port             string
	GenerateRoutes   bool
	RoutesPackage    string
	RoutesOutputFile string
}

// ServeOption is a functional option for configuring Serve
type ServeOption func(*ServeConfig)

// WithPort sets the server port
func WithPort(port string) ServeOption {
	return func(c *ServeConfig) {
		c.Port = port
	}
}

// WithGenerateHelpers enables route helper code generation on server start.
// By default, helpers are generated automatically in development mode (when ROUTER_ENV != "production").
// Use this option to explicitly control route helper generation.
func WithGenerateHelpers(enabled bool) ServeOption {
	return func(c *ServeConfig) {
		c.GenerateRoutes = enabled
	}
}

// WithRoutesPackage sets the routes package name for code generation
func WithRoutesPackage(pkg string) ServeOption {
	return func(c *ServeConfig) {
		c.RoutesPackage = pkg
	}
}

// WithRoutesOutputFile sets the output file path for generated routes
func WithRoutesOutputFile(file string) ServeOption {
	return func(c *ServeConfig) {
		c.RoutesOutputFile = file
	}
}

// listenAndServe is an internal helper that starts the HTTP server.
// Users should use Serve() instead, or http.ListenAndServe(addr, router) for direct control.
func (r *Router) listenAndServe(addr string) error {
	return http.ListenAndServe(addr, r)
}

// Serve starts the HTTP server with optional configuration and automatic route generation.
// By default, route helpers are generated in development mode (ROUTER_ENV != "production").
//
// Defaults:
//   - port: ":3000"
//   - generateRoutes: true in development, false in production
//   - routesPackage: "routes"
//   - routesOutputFile: "routes/generated.go"
//
// Usage:
//
//	r.Serve()                                          // Use all defaults
//	r.Serve(WithPort(":8080"))                         // Custom port
//	r.Serve(WithGenerateHelpers(false))                // Disable helper generation
//	r.Serve(WithPort(":8080"), WithGenerateHelpers(true)) // Production with helper generation
func (r *Router) Serve(opts ...ServeOption) error {
	env := os.Getenv("ROUTER_ENV")
	isProduction := env == "production"

	// Apply defaults
	config := &ServeConfig{
		Port:             ":3000",
		GenerateRoutes:   !isProduction, // Auto-generate in development
		RoutesPackage:    "routes",
		RoutesOutputFile: "routes/generated.go",
	}

	// Apply user options (can override defaults)
	for _, opt := range opts {
		opt(config)
	}

	// Generate route helpers if enabled
	if config.GenerateRoutes {
		fmt.Println("Generating route helpers...")
		if err := r.GenerateRoutes(config.RoutesPackage, config.RoutesOutputFile); err != nil {
			return fmt.Errorf("failed to generate routes: %w", err)
		}
		fmt.Println("Route generation complete!")
	}

	fmt.Printf("Starting server on http://localhost%s\n", config.Port)
	return r.listenAndServe(config.Port)
}
