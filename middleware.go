package router

// MiddlewareFunc is the function signature for middleware
type MiddlewareFunc func(HandlerFunc) HandlerFunc
