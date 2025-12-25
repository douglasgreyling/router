---
title: Examples
---

# Examples

Real-world examples demonstrating various router features.

## Basic REST API

A simple REST API with CRUD operations:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "net/http"
)

type Post struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body"`
}

var posts = []Post{
    {ID: "1", Title: "First Post", Body: "Hello World"},
    {ID: "2", Title: "Second Post", Body: "Another post"},
}

func main() {
    r := router.New()

    r.Get("/posts", listPosts)
    r.Get("/posts/:id", showPost)
    r.Post("/posts", createPost)
    r.Put("/posts/:id", updatePost)
    r.Delete("/posts/:id", deletePost)

    r.Serve()
}

func listPosts(c *router.Context) error {
    return c.JSON(200, posts)
}

func showPost(c *router.Context) error {
    id := c.Param("id")
    for _, post := range posts {
        if post.ID == id {
            return c.JSON(200, post)
        }
    }
    return c.JSON(404, map[string]string{"error": "Post not found"})
}

func createPost(c *router.Context) error {
    var post Post
    // Parse JSON body (simplified)
    posts = append(posts, post)
    return c.JSON(201, post)
}

func updatePost(c *router.Context) error {
    id := c.Param("id")
    // Update logic here
    return c.JSON(200, map[string]string{"id": id, "status": "updated"})
}

func deletePost(c *router.Context) error {
    id := c.Param("id")
    // Delete logic here
    return c.String(204, "")
}
```

## RESTful Resources

Using resource controllers for cleaner code:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "net/http"
)

type UserController struct{}

func (uc *UserController) Index(c *router.Context) error {
    users := []map[string]string{
        {"id": "1", "name": "Alice"},
        {"id": "2", "name": "Bob"},
    }
    return c.JSON(200, users)
}

func (uc *UserController) Show(c *router.Context) error {
    id := c.Param("id")
    user := map[string]string{
        "id":   id,
        "name": "User " + id,
    }
    return c.JSON(200, user)
}

func (uc *UserController) Create(c *router.Context) error {
    // Parse request and create user
    return c.JSON(201, map[string]string{
        "status": "created",
    })
}

func (uc *UserController) Update(c *router.Context) error {
    id := c.Param("id")
    return c.JSON(200, map[string]string{
        "id":     id,
        "status": "updated",
    })
}

func (uc *UserController) Delete(c *router.Context) error {
    return c.String(204, "")
}

func main() {
    r := router.New()

    r.Resources("/users", &UserController{})

    r.Serve()
}
```

## Middleware Chain

Authentication and logging middleware:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "log"
    "time"
)

func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        start := time.Now()
        log.Printf("→ %s %s", c.Method(), c.Path())

        err := next(c)

        duration := time.Since(start)
        log.Printf("← %s %s (%v)", c.Method(), c.Path(), duration)

        return err
    }
}

func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        token := c.Request().Header.Get("Authorization")

        if token != "Bearer secret-token" {
            return c.JSON(401, map[string]string{
                "error": "Unauthorized",
            })
        }

        return next(c)
    }
}

func rateLimitMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        // Rate limiting logic
        return next(c)
    }
}

func main() {
    r := router.New()

    // Global logging
    r.Use(loggingMiddleware)

    // Public routes
    r.Get("/", func(c *router.Context) error {
        return c.String(200, "Public page")
    })

    // Protected API routes
    api := r.Group("/api", authMiddleware, rateLimitMiddleware)
    api.Get("/profile", func(c *router.Context) error {
        return c.JSON(200, map[string]string{
            "user": "authenticated-user",
        })
    })

    r.Serve()
}
```

## API with Versioning

Managing multiple API versions:

```go
package main

import (
    "github.com/douglasgreyling/router"
)

func main() {
    r := router.New()

    // API v1
    v1 := r.Group("/api/v1")
    v1.Get("/users", listUsersV1)
    v1.Get("/posts", listPostsV1)

    // API v2 with breaking changes
    v2 := r.Group("/api/v2")
    v2.Get("/users", listUsersV2)
    v2.Get("/posts", listPostsV2)

    r.Serve()
}

func listUsersV1(c *router.Context) error {
    return c.JSON(200, map[string]interface{}{
        "version": "v1",
        "users":   []string{"Alice", "Bob"},
    })
}

func listUsersV2(c *router.Context) error {
    return c.JSON(200, map[string]interface{}{
        "version": "v2",
        "data": []map[string]interface{}{
            {"id": 1, "name": "Alice", "email": "alice@example.com"},
            {"id": 2, "name": "Bob", "email": "bob@example.com"},
        },
        "meta": map[string]int{"total": 2},
    })
}

func listPostsV1(c *router.Context) error {
    return c.JSON(200, []string{"Post 1", "Post 2"})
}

func listPostsV2(c *router.Context) error {
    return c.JSON(200, map[string]interface{}{
        "posts": []string{"Post 1", "Post 2"},
        "count": 2,
    })
}
```

## Custom Error Handling

Structured error responses:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "log"
)

type APIError struct {
    Code    int    `json:"-"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}

func NewNotFoundError(resource string) *APIError {
    return &APIError{
        Code:    404,
        Message: "Resource not found",
        Details: resource,
    }
}

func NewValidationError(msg string) *APIError {
    return &APIError{
        Code:    400,
        Message: "Validation error",
        Details: msg,
    }
}

func main() {
    r := router.New()

    // Custom error handler
    r.ErrorHandler = func(c *router.Context, err error) {
        if c.IsHeaderWritten() {
            log.Printf("Error after headers sent: %v", err)
            return
        }

        code := 500
        message := "Internal Server Error"
        var details string

        // Check for custom error types
        if apiErr, ok := err.(*APIError); ok {
            code = apiErr.Code
            message = apiErr.Message
            details = apiErr.Details
        }

        response := map[string]interface{}{
            "error": message,
        }
        if details != "" {
            response["details"] = details
        }

        c.JSON(code, response)
    }

    r.Get("/users/:id", func(c *router.Context) error {
        id := c.Param("id")

        if id == "" {
            return NewValidationError("User ID is required")
        }

        // Simulate not found
        if id == "999" {
            return NewNotFoundError("User with ID " + id)
        }

        return c.JSON(200, map[string]string{"id": id, "name": "User"})
    })

    r.Serve()
}
```

## File Server

Serving static files:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "net/http"
    "os"
    "path/filepath"
)

func main() {
    r := router.New()

    // Serve files from public directory
    r.Get("/static/*filepath", func(c *router.Context) error {
        filepath := c.Param("filepath")
        fullPath := filepath.Join("public", filepath)

        // Check if file exists
        if _, err := os.Stat(fullPath); os.IsNotExist(err) {
            return c.String(404, "File not found")
        }

        http.ServeFile(c.ResponseWriter(), c.Request(), fullPath)
        return nil
    })

    // API routes
    r.Get("/api/users", func(c *router.Context) error {
        return c.JSON(200, []string{"Alice", "Bob"})
    })

    r.Serve()
}
```

## Complete Application

A more complete example with multiple features:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "log"
    "time"
)

// Middleware
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        start := time.Now()
        log.Printf("→ %s %s", c.Method(), c.Path())
        err := next(c)
        log.Printf("← %s %s (%v)", c.Method(), c.Path(), time.Since(start))
        return err
    }
}

func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        token := c.Request().Header.Get("Authorization")
        if token != "Bearer secret" {
            return c.JSON(401, map[string]string{"error": "Unauthorized"})
        }
        return next(c)
    }
}

// Controllers
type PostController struct{}

func (pc *PostController) Index(c *router.Context) error {
    posts := []map[string]string{
        {"id": "1", "title": "First Post"},
        {"id": "2", "title": "Second Post"},
    }
    return c.JSON(200, posts)
}

func (pc *PostController) Show(c *router.Context) error {
    id := c.Param("id")
    return c.JSON(200, map[string]string{
        "id":    id,
        "title": "Post " + id,
    })
}

func (pc *PostController) Create(c *router.Context) error {
    return c.JSON(201, map[string]string{"status": "created"})
}

func main() {
    r := router.New()

    // Global middleware
    r.Use(loggingMiddleware)

    // Public routes
    r.Get("/", func(c *router.Context) error {
        return c.HTML(200, "<h1>Welcome to the API</h1>")
    })

    r.Get("/health", func(c *router.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    })

    // Protected API
    api := r.Group("/api", authMiddleware)
    api.Resources("/posts", &PostController{},
        router.Only(router.IndexAction, router.ShowAction, router.CreateAction))

    // Start server
    log.Println("Server starting on :3000")
    r.Serve(router.WithPort(":3000"))
}
```

## Testing Routes

Example of how to test your routes:

```go
package main

import (
    "github.com/douglasgreyling/router"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetUser(t *testing.T) {
    r := router.New()

    r.Get("/users/:id", func(c *router.Context) error {
        id := c.Param("id")
        return c.JSON(200, map[string]string{"id": id})
    })

    // Create test request
    req := httptest.NewRequest("GET", "/users/123", nil)
    rec := httptest.NewRecorder()

    // Serve the request
    r.ServeHTTP(rec, req)

    // Check response
    if rec.Code != 200 {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }

    expected := `{"id":"123"}`
    if rec.Body.String() != expected {
        t.Errorf("Expected body %s, got %s", expected, rec.Body.String())
    }
}
```
