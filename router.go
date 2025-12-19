package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// HandlerFunc is the function signature for route handlers
type HandlerFunc func(*Context) error

// MiddlewareFunc is the function signature for middleware
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Params holds route parameters extracted from the URL
type Params map[string]string

// nodeType represents the type of node in the radix tree
type nodeType uint8

const (
	static   nodeType = iota // static route segment
	param                    // :param - matches a single segment
	wildcard                 // *wildcard - matches everything after
)

// node represents a node in the radix tree
type node struct {
	// The path segment this node represents
	path string

	// Type of node (static, param, wildcard)
	nType nodeType

	// The full pattern if this node ends a route
	pattern string

	// Handlers for different HTTP methods
	handlers map[string]HandlerFunc

	// Child nodes
	children []*node

	// Parameter name if this is a param or wildcard node
	paramName string

	// Middleware chain for this specific route
	middleware []MiddlewareFunc
}

// Router is the main router structure
type Router struct {
	// Root nodes for each HTTP method
	trees map[string]*node

	// Global middleware applied to all routes
	middleware []MiddlewareFunc

	// NotFound handler
	NotFound HandlerFunc

	// MethodNotAllowed handler
	MethodNotAllowed HandlerFunc

	// ErrorHandler handles errors returned from handlers
	ErrorHandler func(*Context, error)

	// Named routes for reverse routing and path helper generation
	namedRoutes map[string]*namedRoute
}

// namedRoute stores information about a named route
type namedRoute struct {
	name    string
	pattern string
	method  string
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		trees:       make(map[string]*node),
		namedRoutes: make(map[string]*namedRoute),
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

// Handle registers a new route with the given method and path.
// A route name is automatically generated for path helper generation.
// Use the HTTP method helpers (Get, Post, etc.) with WithName() for custom names.
//
// Panics if:
//   - path does not begin with '/'
//   - path contains duplicate parameter names (e.g., /users/:id/posts/:id)
//
// This method is intended for use during application startup. Panics are
// deliberate to catch configuration errors early, preventing the application
// from starting with invalid routes.
func (r *Router) Handle(method, path string, handler HandlerFunc, name string, middleware ...MiddlewareFunc) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	if r.trees[method] == nil {
		r.trees[method] = &node{
			path:     "/",
			handlers: make(map[string]HandlerFunc),
			children: make([]*node, 0),
		}
	}

	r.addRoute(method, path, handler, middleware)

	// Auto-generate route name if not provided
	if name == "" {
		name = generateRouteName(path, method)
	}

	// Register named route
	if name != "" {
		r.addNamedRoute(name, path, method)
	}
}

// Get registers a GET route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Get(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("GET", path, handler, name, middleware...)
}

// Post registers a POST route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Post(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("POST", path, handler, name, middleware...)
}

// Put registers a PUT route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Put(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("PUT", path, handler, name, middleware...)
}

// Patch registers a PATCH route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Patch(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("PATCH", path, handler, name, middleware...)
}

// Delete registers a DELETE route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Delete(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("DELETE", path, handler, name, middleware...)
}

// Head registers a HEAD route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Head(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("HEAD", path, handler, name, middleware...)
}

// Options registers an OPTIONS route with optional name and middleware.
// Panics on invalid paths (see Handle for details).
func (r *Router) Options(path string, handler HandlerFunc, opts ...RouteOption) {
	name, middleware := parseRouteOptions(opts)
	r.Handle("OPTIONS", path, handler, name, middleware...)
}

// addNamedRoute registers a named route for code generation
func (r *Router) addNamedRoute(name, pattern, method string) {
	r.namedRoutes[name] = &namedRoute{
		name:    name,
		pattern: pattern,
		method:  method,
	}
}

// generateRouteName creates a route name from path and HTTP method
// Examples:
//
//	GET /users -> users_index
//	GET /users/:id -> users_show
//	POST /users -> users_create
//	PUT /users/:id -> users_update
//	PATCH /users/:id -> users_update
//	DELETE /users/:id -> users_destroy
//	GET /api/v1/products/:id -> api_v1_products_show
func generateRouteName(path, method string) string {
	// Clean the path: remove leading/trailing slashes and parameters
	path = strings.Trim(path, "/")
	if path == "" {
		return "" // Don't auto-name root path
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	// Build base name from non-parameter segments
	var baseParts []string
	hasParams := false

	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			hasParams = true
			// Skip parameter segments in the base name
			continue
		}
		// Replace hyphens with underscores for valid identifiers
		cleanSegment := strings.ReplaceAll(segment, "-", "_")
		baseParts = append(baseParts, cleanSegment)
	}

	if len(baseParts) == 0 {
		return "" // Path only has parameters
	}

	baseName := strings.Join(baseParts, "_")

	// Determine action suffix based on method and whether path has parameters
	var action string
	switch method {
	case "GET":
		if hasParams {
			action = "show"
		} else {
			action = "index"
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "destroy"
	case "HEAD":
		if hasParams {
			action = "show"
		} else {
			action = "index"
		}
	case "OPTIONS":
		action = "options"
	default:
		action = strings.ToLower(method)
	}

	return baseName + "_" + action
}

// addRoute adds a route to the radix tree
func (r *Router) addRoute(method, path string, handler HandlerFunc, middleware []MiddlewareFunc) {
	root := r.trees[method]

	if path == "/" {
		root.handlers[method] = handler
		root.pattern = path
		root.middleware = middleware
		return
	}

	// Remove leading and trailing slashes, split path
	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")

	// Validate no duplicate parameter names
	paramNames := make(map[string]bool)
	for _, segment := range segments {
		if len(segment) > 0 && (segment[0] == ':' || segment[0] == '*') {
			paramName := segment[1:]
			if paramNames[paramName] {
				panic(fmt.Sprintf("duplicate parameter name '%s' in route '%s %s'", paramName, method, path))
			}
			paramNames[paramName] = true
		}
	}

	current := root
	for i, segment := range segments {
		// Determine node type
		nType := static
		paramName := ""

		if len(segment) > 0 {
			if segment[0] == ':' {
				nType = param
				paramName = segment[1:]
			} else if segment[0] == '*' {
				nType = wildcard
				paramName = segment[1:]
			}
		}

		// Look for existing child with matching segment
		var next *node
		for _, child := range current.children {
			if child.path == segment && child.nType == nType {
				next = child
				break
			}
		}

		// Create new node if no match found
		if next == nil {
			next = &node{
				path:      segment,
				nType:     nType,
				paramName: paramName,
				handlers:  make(map[string]HandlerFunc),
				children:  make([]*node, 0),
			}
			current.children = append(current.children, next)
		}

		// If this is the last segment, set the handler
		if i == len(segments)-1 {
			next.handlers[method] = handler
			next.pattern = "/" + strings.Join(segments, "/")
			next.middleware = middleware
		}

		current = next
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	// Create context
	c := newContext(w, req)

	// Get the root node for this method
	root := r.trees[method]

	var handler HandlerFunc
	var middleware []MiddlewareFunc

	if root != nil {
		// Find the matching route
		handler, c.Params, middleware = r.findRoute(root, path, method)
	}

	if handler == nil {
		// Check if route exists for a different method
		for m := range r.trees {
			if m != method {
				h, _, _ := r.findRoute(r.trees[m], path, m)
				if h != nil {
					if err := r.MethodNotAllowed(c); err != nil && r.ErrorHandler != nil {
						r.ErrorHandler(c, err)
					}
					return
				}
			}
		}
		if err := r.NotFound(c); err != nil && r.ErrorHandler != nil {
			r.ErrorHandler(c, err)
		}
		return
	}

	// Build middleware chain (global + route-specific)
	finalHandler := handler

	// Apply route-specific middleware first (innermost)
	for i := len(middleware) - 1; i >= 0; i-- {
		finalHandler = middleware[i](finalHandler)
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

// findRoute finds a matching route in the tree
func (r *Router) findRoute(root *node, path, method string) (HandlerFunc, Params, []MiddlewareFunc) {
	if path == "/" {
		if handler, ok := root.handlers[method]; ok {
			return handler, nil, root.middleware
		}
		return nil, nil, nil
	}

	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")
	params := make(Params)

	handler, middleware := r.search(root, segments, 0, params, method)
	return handler, params, middleware
}

// search recursively searches for a matching route
func (r *Router) search(n *node, segments []string, index int, params Params, method string) (HandlerFunc, []MiddlewareFunc) {
	// If we've matched all segments, check if this node has a handler
	if index == len(segments) {
		if handler, ok := n.handlers[method]; ok {
			return handler, n.middleware
		}
		return nil, nil
	}

	segment := segments[index]

	// Try children in order: static > param > wildcard
	for _, child := range n.children {
		switch child.nType {
		case static:
			if child.path == segment {
				if handler, middleware := r.search(child, segments, index+1, params, method); handler != nil {
					return handler, middleware
				}
			}
		case param:
			params[child.paramName] = segment
			if handler, middleware := r.search(child, segments, index+1, params, method); handler != nil {
				return handler, middleware
			}
			delete(params, child.paramName) // backtrack
		case wildcard:
			// Wildcard matches everything remaining
			params[child.paramName] = strings.Join(segments[index:], "/")
			if handler, ok := child.handlers[method]; ok {
				return handler, child.middleware
			}
		}
	}

	return nil, nil
}

// GenerateRoutes generates type-safe route helpers
func (r *Router) GenerateRoutes(packageName, outputFile string) error {
	cg := NewPathHelperGenerator()

	// print out all named routes
	fmt.Printf("Generating route helpers for %d named routes...\n", len(r.namedRoutes))

	for name, route := range r.namedRoutes {
		cg.AddRoute(name, route.pattern, route.method)
	}
	return cg.Generate(packageName, outputFile)
}

// ServeConfig holds configuration for the Serve method
type ServeConfig struct {
	Addr             string
	GenerateRoutes   bool
	RoutesPackage    string
	RoutesOutputFile string
}

// ServeOption is a functional option for configuring Serve
type ServeOption func(*ServeConfig)

// WithAddr sets the server address
func WithAddr(addr string) ServeOption {
	return func(c *ServeConfig) {
		c.Addr = addr
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

// ListenAndServe starts the HTTP server on the specified address.
// This is a thin wrapper around http.ListenAndServe for convenience.
func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r)
}

// Serve starts the HTTP server with optional configuration and automatic route generation.
// By default, route helpers are generated in development mode (ROUTER_ENV != "production").
//
// Defaults:
//   - addr: ":3000"
//   - generateRoutes: true in development, false in production
//   - routesPackage: "routes"
//   - routesOutputFile: "routes/generated.go"
//
// Usage:
//
//	r.Serve()                                          // Use all defaults
//	r.Serve(WithAddr(":8080"))                         // Custom port
//	r.Serve(WithGenerateHelpers(false))                // Disable helper generation
//	r.Serve(WithAddr(":8080"), WithGenerateHelpers(true)) // Production with helper generation
func (r *Router) Serve(opts ...ServeOption) error {
	env := os.Getenv("ROUTER_ENV")
	isProduction := env == "production"

	// Apply defaults
	config := &ServeConfig{
		Addr:             ":3000",
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

	fmt.Printf("Starting server on http://localhost%s\n", config.Addr)
	return r.ListenAndServe(config.Addr)
}
