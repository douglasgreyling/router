package router

// RouteOption is a functional option for configuring routes
type RouteOption interface {
	applyToRoute(*routeConfig)
}

// routeConfig holds the configuration for a route
type routeConfig struct {
	name       string
	middleware []MiddlewareFunc
}

// routeName is an option that sets the route name
type routeName string

func (n routeName) applyToRoute(cfg *routeConfig) {
	cfg.name = string(n)
}

// WithName sets the name for a route (for reverse routing and code generation)
func WithName(name string) RouteOption {
	return routeName(name)
}

// routeMiddleware is an option that adds middleware to a route
type routeMiddleware []MiddlewareFunc

func (m routeMiddleware) applyToRoute(cfg *routeConfig) {
	cfg.middleware = append(cfg.middleware, m...)
}

// WithMiddleware adds middleware to a specific route
func WithMiddleware(middleware ...MiddlewareFunc) RouteOption {
	return routeMiddleware(middleware)
}

// parseRouteOptions extracts configuration from route options
func parseRouteOptions(opts []RouteOption) (string, []MiddlewareFunc) {
	cfg := &routeConfig{}
	for _, opt := range opts {
		opt.applyToRoute(cfg)
	}
	return cfg.name, cfg.middleware
}
