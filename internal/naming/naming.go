package naming

import "strings"

// Route stores information about a named route
type Route struct {
	Name    string
	Pattern string
	Method  string
}

// Registry manages named routes for reverse routing and code generation
type Registry struct {
	routes map[string]*Route
}

// NewRegistry creates a new naming registry
func NewRegistry() *Registry {
	return &Registry{
		routes: make(map[string]*Route),
	}
}

// Add registers a named route
func (r *Registry) Add(name, pattern, method string) {
	r.routes[name] = &Route{
		Name:    name,
		Pattern: pattern,
		Method:  method,
	}
}

// Get retrieves a named route by name
func (r *Registry) Get(name string) (*Route, bool) {
	route, ok := r.routes[name]
	return route, ok
}

// All returns all named routes
func (r *Registry) All() map[string]*Route {
	return r.routes
}

// GenerateName creates a route name from path and HTTP method
// Examples:
//
//	GET /users -> users_index
//	GET /users/:id -> users_show
//	POST /users -> users_create
//	PUT /users/:id -> users_update
//	PATCH /users/:id -> users_update
//	DELETE /users/:id -> users_destroy
//	GET /api/v1/products/:id -> api_v1_products_show
func GenerateName(path, method string) string {
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
