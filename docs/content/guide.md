---
title: User Guide
---

# User Guide

A comprehensive guide to using the Router package.

## Table of Contents

- [Middleware](#middleware)
- [Route Groups](#route-groups)
- [RESTful Resources](#restful-resources)
- [Route Helpers](#route-helpers)
- [Named Routes](#named-routes)
- [Context API](#context-api)

## Middleware

Middleware wraps handlers to add functionality before and/or after request handling. Middleware is executed in order: global → group → route-specific.

### Creating Middleware

```go
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        start := time.Now()
        log.Printf("Started %s %s", c.Method(), c.Path())

        err := next(c)  // Call the next handler

        log.Printf("Completed in %v", time.Since(start))
        return err
    }
}
```

### Applying Middleware

#### Global Middleware

Applied to all routes:

```go
r := router.New()
r.Use(loggingMiddleware)
r.Use(authMiddleware)

r.Get("/users", listUsers)  // Uses both middleware
```

#### Group Middleware

Applied to all routes in a group:

```go
api := r.Group("/api", authMiddleware, rateLimitMiddleware)
api.Get("/users", listUsers)     // Uses auth + rate limit
api.Get("/posts", listPosts)     // Uses auth + rate limit
```

#### Route-Specific Middleware

Applied to individual routes:

```go
r.Get("/admin", adminHandler,
    router.WithMiddleware(requireAdmin))
```

### Middleware Examples

#### Authentication Middleware

```go
func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        token := c.Request().Header.Get("Authorization")

        if !isValidToken(token) {
            return c.JSON(401, map[string]string{
                "error": "Unauthorized",
            })
        }

        return next(c)
    }
}
```

#### CORS Middleware

```go
func corsMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        c.ResponseWriter().Header().Set("Access-Control-Allow-Origin", "*")
        c.ResponseWriter().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

        if c.Method() == "OPTIONS" {
            return c.String(204, "")
        }

        return next(c)
    }
}
```

## Route Groups

Groups allow you to organize routes with common prefixes and middleware:

```go
r := router.New()

// API v1 group
v1 := r.Group("/api/v1", authMiddleware)
v1.Get("/users", listUsers)
v1.Post("/users", createUser)

// API v2 group
v2 := r.Group("/api/v2", authMiddleware, newFeatureMiddleware)
v2.Get("/users", listUsersV2)
v2.Post("/users", createUserV2)

// Admin group
admin := r.Group("/admin", authMiddleware, adminMiddleware)
admin.Get("/dashboard", showDashboard)
admin.Get("/stats", showStats)
```

### Nested Groups

Groups can be nested for better organization:

```go
api := r.Group("/api")
v1 := api.Group("/v1", authMiddleware)
users := v1.Group("/users")

users.Get("/", listUsers)        // GET /api/v1/users
users.Get("/:id", showUser)      // GET /api/v1/users/:id
users.Post("/", createUser)      // POST /api/v1/users
```

## RESTful Resources

The router provides Rails-inspired resource scaffolding for quickly creating RESTful APIs.

### Controller Interface

Define a controller implementing the desired actions:

```go
type PostController struct{}

func (pc *PostController) Index(c *router.Context) error {
    posts := getAllPosts()
    return c.JSON(200, posts)
}

func (pc *PostController) Show(c *router.Context) error {
    id := c.Param("id")
    post := findPost(id)
    return c.JSON(200, post)
}

func (pc *PostController) Create(c *router.Context) error {
    // Parse request body and create post
    return c.JSON(201, newPost)
}

func (pc *PostController) Edit(c *router.Context) error {
    id := c.Param("id")
    post := findPost(id)
    return c.JSON(200, post)
}

func (pc *PostController) Update(c *router.Context) error {
    id := c.Param("id")
    // Update post logic
    return c.JSON(200, updatedPost)
}

func (pc *PostController) Delete(c *router.Context) error {
    id := c.Param("id")
    // Delete post logic
    return c.String(204, "")
}
```

### Registering Resources

```go
r.Resources("/posts", &PostController{})
```

This automatically creates these routes:

| Method | Path | Action | Description |
|--------|------|--------|-------------|
| GET | /posts | Index | List all posts |
| GET | /posts/new | New | Show create form |
| POST | /posts | Create | Create a post |
| GET | /posts/:id | Show | Show a specific post |
| GET | /posts/:id/edit | Edit | Show edit form |
| PATCH/PUT | /posts/:id | Update | Update a post |
| DELETE | /posts/:id | Delete | Delete a post |

### Limiting Actions

Use `Only` or `Except` to control which actions are created:

```go
// Only these actions
r.Resources("/posts", &PostController{},
    router.Only(router.IndexAction, router.ShowAction))

// All actions except these
r.Resources("/comments", &CommentController{},
    router.Except(router.NewAction, router.EditAction))
```

### Resource Middleware

Apply middleware to all resource routes:

```go
r.Resources("/posts", &PostController{},
    router.WithResourceMiddleware(authMiddleware, loggingMiddleware))
```

## Route Helpers

The router automatically generates type-safe route helpers in development mode.

### Route Naming

Routes can be named in two ways:

#### Explicit Naming

Use `WithName` to explicitly name a route:

```go
r.Get("/users/:id", showUser, router.WithName("users_show"))
r.Post("/login", loginHandler, router.WithName("auth_login"))
```

#### Automatic Naming

Routes without explicit names are automatically named based on their path and HTTP method:

```go
r.Get("/users/:id", showUser)
// Auto-generated name: "users_id" or similar

r.Get("/posts/:post_id/comments", listComments)
// Auto-generated name: "posts_post_id_comments" or similar
```

**Note:** Explicit naming with `WithName` is recommended for important routes to ensure stable helper function names.

#### Resource Route Names

Resource routes are automatically named using a standard convention:

```go
r.Resources("/posts", &PostController{})
```

This creates the following named routes:

| Action | Method | Path | Route Name |
|--------|--------|------|------------|
| Index | GET | /posts | `posts_index` |
| New | GET | /posts/new | `posts_new` |
| Create | POST | /posts | `posts_create` |
| Show | GET | /posts/:id | `posts_show` |
| Edit | GET | /posts/:id/edit | `posts_edit` |
| Update | PATCH/PUT | /posts/:id | `posts_update` |
| Delete | DELETE | /posts/:id | `posts_delete` |

The naming pattern is: `{resource_name}_{action}` (e.g., `posts_show`, `users_create`)

### Generated Helpers

For named routes, helpers are generated in `routes/generated.go`:

```go
r.Get("/users/:id", showUser, router.WithName("users_show"))
r.Get("/posts/:id/comments/:comment_id", showComment,
    router.WithName("post_comments_show"))
```

Generated helpers:

```go
import "yourapp/routes"

// Simple route
url := routes.UsersShowPath("123")
// Returns: /users/123

// Nested route
url := routes.PostCommentsShowPath("456", "789")
// Returns: /posts/456/comments/789

// Resource routes
url := routes.PostsShowPath("42")
// Returns: /posts/42

url := routes.PostsIndexPath()
// Returns: /posts
```

### Customizing Generation

```go
r.Serve(
    router.WithRoutesPackage("api"),
    router.WithRoutesOutputFile("api/routes.go"),
    router.WithGenerateHelpers(true),
)
```

### Disabling Generation

```go
// Disable in production
r.Serve(router.WithGenerateHelpers(false))
```

## Context API

The Context provides methods for request handling and response generation.

### Request Information

```go
func handler(c *router.Context) error {
    method := c.Method()            // HTTP method
    path := c.Path()                // Request path
    req := c.Request()              // *http.Request

    // Route parameters
    id := c.Param("id")
    filepath := c.Param("filepath")

    return nil
}
```

### Response Methods

#### JSON Response

```go
func handler(c *router.Context) error {
    data := map[string]interface{}{
        "users": []string{"Alice", "Bob"},
        "count": 2,
    }
    return c.JSON(200, data)
}
```

#### String Response

```go
func handler(c *router.Context) error {
    return c.String(200, "Hello, World!")
}
```

#### HTML Response

```go
func handler(c *router.Context) error {
    html := "<html><body><h1>Welcome</h1></body></html>"
    return c.HTML(200, html)
}
```

### Direct Access

For advanced use cases, access the underlying objects:

```go
func handler(c *router.Context) error {
    // Direct access to http.ResponseWriter
    w := c.ResponseWriter()
    w.Header().Set("X-Custom-Header", "value")

    // Direct access to http.Request
    r := c.Request()
    userAgent := r.Header.Get("User-Agent")

    // Check if headers were already written
    if c.IsHeaderWritten() {
        // Cannot write headers again
    }

    return c.String(200, "Response")
}
```

## Best Practices

### 1. Use Middleware Wisely

Apply middleware at the appropriate level:

```go
// Global: logging, CORS, recovery
r.Use(loggingMiddleware, corsMiddleware)

// Group: authentication for API routes
api := r.Group("/api", authMiddleware)

// Route: specific rate limiting
r.Get("/expensive", handler, router.WithMiddleware(rateLimitMiddleware))
```

### 2. Organize with Groups

```go
r := router.New()

// Public routes
r.Get("/", homeHandler)
r.Get("/about", aboutHandler)

// Authenticated API
api := r.Group("/api", authMiddleware)
api.Get("/profile", profileHandler)

// Admin panel
admin := r.Group("/admin", authMiddleware, adminMiddleware)
admin.Get("/dashboard", dashboardHandler)
```

### 3. Use Resources for CRUD

When building RESTful APIs, prefer resources over individual routes:

```go
// Instead of:
r.Get("/posts", listPosts)
r.Post("/posts", createPost)
r.Get("/posts/:id", showPost)
// ...

// Use:
r.Resources("/posts", &PostController{})
```

### 4. Handle Errors Consistently

```go
r.ErrorHandler = func(c *router.Context, err error) {
    if c.IsHeaderWritten() {
        log.Printf("Error after headers sent: %v", err)
        return
    }

    code := 500
    message := "Internal Server Error"

    // Type assert for custom error types
    if apiErr, ok := err.(*APIError); ok {
        code = apiErr.Code
        message = apiErr.Message
    }

    c.JSON(code, map[string]string{"error": message})
}
```

### 5. Name Important Routes

```go
// Generate type-safe URL helpers
r.Get("/users/:id", showUser, router.WithName("users_show"))
r.Get("/posts/:id/edit", editPost, router.WithName("posts_edit"))
```
