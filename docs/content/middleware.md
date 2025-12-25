---
title: Middleware
---

## Introduction

Middleware lets you wrap your route handlers with additional functionality, think logging, authentication, request validation, or anything else you need to run before or after your handlers execute.

You can apply middleware at three different levels:
- **Globally** - runs on every single route
- **Group-level** - runs on all routes within a group
- **Route-specific** - runs only on individual routes (this includes resources)

The beauty of middleware is that you can chain multiple together, and they'll execute in a predictable order.

## Creating Middleware

Middleware in this router is just a function that wraps a handler. The signature is simple:

```go
func(router.HandlerFunc) router.HandlerFunc
```

Here's what a basic middleware looks like:

```go
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        // Code here runs BEFORE the handler
        log.Printf("Request: %s %s", c.Method(), c.Path())

        // Call the next handler in the chain
        err := next(c)

        // Code here runs AFTER the handler
        log.Printf("Response status: %d", c.GetStatus())

        return err
    }
}
```

The pattern is always the same:
1. Do something before (optional)
2. Call `next(c)` to execute the next handler
3. Do something after (optional)
4. Return the error from `next(c)`

### Example: Timing Middleware

Let's build something practical like middleware that measures how long each request takes:

```go
import "time"

func timingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        start := time.Now()

        // Execute the handler
        err := next(c)

        // Calculate and log the duration
        duration := time.Since(start)
        log.Printf("%s %s completed in %v", c.Method(), c.Path(), duration)

        return err
    }
}
```

### Example: Authentication Middleware

Here's a more practical example that checks for authentication:

```go
func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        token := c.Header("Authorization")

        if token == "" {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Missing authorization token",
            })
        }

        // Validate the token
        user, err := validateToken(token)
        if err != nil {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Invalid token",
            })
        }

        // Store the user in context for handlers to use
        c.Set("user", user)

        // Continue to the next handler
        return next(c)
    }
}
```

Notice how this middleware can short-circuit the chain by returning early if authentication fails. When you return without calling `next(c)`, the handler never runs.

### Example: CORS Middleware

Need to handle Cross-Origin Resource Sharing? Here's how:

```go
func corsMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        // Set CORS headers
        c.SetHeader("Access-Control-Allow-Origin", "*")
        c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight requests
        if c.Method() == "OPTIONS" {
            return c.NoContent(http.StatusOK)
        }

        // Continue to the next handler
        return next(c)
    }
}
```

## Global Middleware

Global middleware runs on every single route in your application. It's perfect for things like logging, request IDs, or panic recovery that you want everywhere.

To add global middleware, use the `Use()` method on your router:

```go
func main() {
    r := router.New()

    // Add global middleware
    r.Use(loggingMiddleware)
    r.Use(timingMiddleware)
    r.Use(corsMiddleware)

    // All routes will have these middleware
    r.Get("/users", listUsers)
    r.Post("/users", createUser)

    r.Serve()
}
```

You can add multiple middleware at once:

```go
r.Use(loggingMiddleware, timingMiddleware, corsMiddleware)
```

## Group Middleware

Group middleware runs on all routes within a specific group. This is perfect when you have a section of your API that needs special treatment like requiring authentication for all admin routes, but leaving public routes open.

### Creating Groups with Middleware

You can add middleware when creating a group or add it later:

```go
func main() {
    r := router.New()

    // Public routes (no middleware)
    r.Get("/", homePage)
    r.Get("/about", aboutPage)

    // API routes with authentication
    api := r.Group("/api", authMiddleware)
    api.Get("/users", listUsers)           // Requires auth
    api.Post("/posts", createPost)         // Requires auth

    // Admin routes with extra middleware
    admin := r.Group("/admin", authMiddleware, adminCheckMiddleware)
    admin.Get("/users", listAllUsers)      // Requires auth + admin
    admin.Delete("/users/:id", deleteUser) // Requires auth + admin

    r.Serve()
}
```

### Adding Middleware to Existing Groups

You can also add middleware to a group after creating it:

```go
api := r.Group("/api")

// Add middleware later
api.Use(authMiddleware)
api.Use(rateLimitMiddleware)

api.Get("/users", listUsers)
```

### Nested Groups

Groups can be nested, and middleware stacks up as you go deeper:

```go
func main() {
    r := router.New()

    // API group with logging
    api := r.Group("/api", loggingMiddleware)

    // V1 group inherits logging, adds rate limiting
    v1 := api.Group("/v1", rateLimitMiddleware)
    v1.Get("/users", listUsers)  // Has: logging + rate limiting

    // Admin group inherits logging, adds auth
    admin := api.Group("/admin", authMiddleware)
    admin.Get("/users", adminListUsers)  // Has: logging + auth

    r.Serve()
}
```

## Route-Specific Middleware

Sometimes you need middleware on just one or two specific routes. Maybe you want to cache one endpoint, or add extra validation to a particular route. That's where route-specific middleware shines.

Use the `WithMiddleware()` option when defining routes:

```go
r.Get("/users", listUsers, router.WithMiddleware(cacheMiddleware))

r.Post("/upload", handleUpload,
    router.WithMiddleware(validateFileMiddleware, virusScanMiddleware))
```

### Combining with Other Options

Route-specific middleware works alongside other route options:

```go
r.Get("/users/:id", showUser,
    router.WithName("user_show"), // Router helper name option
    router.WithMiddleware(cacheMiddleware, metricsMiddleware)) // Route-specific middleware option
```

## Resource Middleware

When using `Resources()` to create RESTful routes, you can apply middleware to all resource routes at once using the `WithResourceMiddleware()` option. This is perfect when you want all CRUD operations for a resource to share the same middleware.

### Adding Middleware to Resources

```go
type PostController struct{}

func (pc *PostController) Index(c *router.Context) error {
    posts := getAllPosts()
    return c.JSON(http.StatusOK, posts)
}

func (pc *PostController) Create(c *router.Context) error {
    var post Post
    if err := c.BindJSON(&post); err != nil {
        return err
    }
    created := createPost(post)
    return c.JSON(http.StatusCreated, created)
}

// ... other controller methods

func main() {
    r := router.New()

    // Apply middleware to all post routes
    r.Resources("/posts", &PostController{},
        router.WithResourceMiddleware(authMiddleware, loggingMiddleware))

    r.Serve()
}
```

This applies the middleware to all seven RESTful routes: Index, New, Create, Show, Edit, Update, and Delete.

### Combining with Only/Except

Resource middleware works alongside `Only()` and `Except()` options:

```go
// Auth middleware on all API endpoints (no form routes)
r.Resources("/posts", &PostController{},
    router.Except(router.NewAction, router.EditAction),
    router.WithResourceMiddleware(authMiddleware))

// Rate limiting only on read operations
r.Resources("/posts", &PostController{},
    router.Only(router.IndexAction, router.ShowAction),
    router.WithResourceMiddleware(rateLimitMiddleware))
```

### Resource Middleware vs Group Middleware

You might wonder: when should I use resource middleware versus putting resources in a group? Here's a quick guide:

**Use Resource Middleware when:**
- The middleware is specific to that resource's operations
- You want middleware only on CRUD operations, not other routes in the path
- You're using `Only()` or `Except()` to limit actions

```go
// ✓ Good: Middleware specific to posts
r.Resources("/posts", &PostController{},
    router.WithResourceMiddleware(postValidationMiddleware))
```

**Use Group Middleware when:**
- You have multiple resources under the same path prefix
- The middleware applies to more than just resource routes
- You want to share middleware across different resource types

```go
// ✓ Good: Shared middleware for multiple resources
api := r.Group("/api", authMiddleware, loggingMiddleware)
api.Resources("/posts", &PostController{})
api.Resources("/comments", &CommentController{})
api.Get("/stats", statsHandler)  // Also gets the group middleware
```

**Combine both when:**
- You need shared middleware for all API routes
- Plus specific middleware for certain resources

```go
// ✓ Good: Both shared and specific middleware
api := r.Group("/api", authMiddleware)  // All API routes need auth
api.Resources("/posts", &PostController{})  // Just gets auth
api.Resources("/admin/users", &UserController{},
    router.WithResourceMiddleware(adminOnlyMiddleware))  // Gets auth + admin
```

## Middleware Execution Order

Understanding the order in which middleware executes is crucial. The router has a clear, predictable execution order:

**Order: Global (before) → Group (before) → Route-specific (before) → Handler → Route-specific (after) → Group (after) → Global (after)**

Middleware wraps around your handlers like layers of an onion. Each layer runs its "before" code going in, then runs its "after" code coming back out.

### Multiple Middleware at Each Level

When you have multiple middleware at the same level, they execute in the order you register them:

```go
r.Use(logging, timing, cors)  // Executes: logging → timing → cors

api := r.Group("/api", auth, rateLimit)  // Executes: auth → rateLimit

r.Get("/users", handler,
    router.WithMiddleware(validate, sanitize))  // Executes: validate → sanitize
```
