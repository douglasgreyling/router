package router

// MiddlewareFunc is the function signature for middleware.
// Middleware wraps a HandlerFunc and can perform actions before and/or after
// the handler executes. Multiple middleware are chained together, with each
// middleware calling the next handler in the chain.
//
// Middleware execution order: global → group → route-specific
//
// Example middleware:
//
//	func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
//	    return func(c *router.Context) error {
//	        start := time.Now()
//	        log.Printf("Started %s %s", c.Method(), c.Path())
//	
//	        err := next(c)  // Call the next handler
//	
//	        log.Printf("Completed in %v", time.Since(start))
//	        return err
//	    }
//	}
//
// Apply middleware:
//
//	r.Use(loggingMiddleware)                      // Global
//	api := r.Group("/api", authMiddleware)        // Group
//	r.Get("/users", handler, WithMiddleware(mw))  // Route-specific
type MiddlewareFunc func(HandlerFunc) HandlerFunc
