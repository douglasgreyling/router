---
title: API Reference
---

# API Reference

Complete API documentation for the Router package.

## Router

### Creating a Router

```go
func New() *Router
```

Creates a new router instance with default configuration.

### HTTP Methods

```go
func (r *Router) Get(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Post(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Put(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Patch(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Delete(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Head(path string, handler HandlerFunc, opts ...Option) *route
func (r *Router) Options(path string, handler HandlerFunc, opts ...Option) *route
```

Register a handler for the specified HTTP method and path.

**Parameters:**
- `path`: The route pattern (supports `:param` and `*wildcard`)
- `handler`: The handler function to execute
- `opts`: Optional route configuration (e.g., `WithName`, `WithMiddleware`)

**Example:**
```go
r.Get("/users/:id", showUser, router.WithName("users_show"))
```

### Middleware

```go
func (r *Router) Use(middleware ...MiddlewareFunc)
```

Register global middleware that applies to all routes.

**Example:**
```go
r.Use(loggingMiddleware, authMiddleware)
```

### Groups

```go
func (r *Router) Group(prefix string, middleware ...MiddlewareFunc) *Router
```

Create a route group with a common path prefix and optional middleware.

**Parameters:**
- `prefix`: Path prefix for all routes in the group
- `middleware`: Optional middleware applied to all group routes

**Example:**
```go
api := r.Group("/api/v1", authMiddleware)
api.Get("/users", listUsers)
```

### Resources

```go
func (r *Router) Resources(path string, controller Controller, opts ...ResourceOption)
```

Register RESTful resource routes for a controller.

**Parameters:**
- `path`: Base path for the resource
- `controller`: Controller implementing resource actions
- `opts`: Optional resource configuration (e.g., `Only`, `Except`, `WithResourceMiddleware`)

**Example:**
```go
r.Resources("/posts", &PostController{},
    router.Only(router.IndexAction, router.ShowAction))
```

### Starting the Server

```go
func (r *Router) Serve(opts ...ServerOption) error
```

Start the HTTP server with optional configuration.

**Parameters:**
- `opts`: Optional server configuration (e.g., `WithPort`, `WithGenerateHelpers`)

**Example:**
```go
r.Serve(router.WithPort(8080))
```

### Error Handler

```go
type ErrorHandlerFunc func(*Context, error)

r.ErrorHandler = func(c *Context, err error) {
    // Custom error handling
}
```

Customize how errors are handled globally.

## Context

The Context provides methods for handling requests and responses.

### Request Methods

```go
func (c *Context) Request() *http.Request
func (c *Context) Method() string
func (c *Context) Path() string
func (c *Context) Param(name string) string
```

**Examples:**
```go
req := c.Request()          // Access *http.Request
method := c.Method()        // "GET", "POST", etc.
path := c.Path()            // "/users/123"
id := c.Param("id")         // "123"
```

### Response Methods

```go
func (c *Context) ResponseWriter() http.ResponseWriter
func (c *Context) IsHeaderWritten() bool
func (c *Context) JSON(code int, data interface{}) error
func (c *Context) String(code int, s string) error
func (c *Context) HTML(code int, html string) error
```

**JSON Response:**
```go
return c.JSON(200, map[string]string{"status": "ok"})
```

**String Response:**
```go
return c.String(200, "Hello, World!")
```

**HTML Response:**
```go
return c.HTML(200, "<h1>Welcome</h1>")
```

**Check Header Status:**
```go
if c.IsHeaderWritten() {
    // Headers already sent, cannot modify
}
```

## HandlerFunc

```go
type HandlerFunc func(*Context) error
```

Handler functions receive a Context and return an error. Errors are processed by the ErrorHandler.

**Example:**
```go
func showUser(c *router.Context) error {
    id := c.Param("id")
    user, err := findUser(id)
    if err != nil {
        return err  // Handled by ErrorHandler
    }
    return c.JSON(200, user)
}
```

## MiddlewareFunc

```go
type MiddlewareFunc func(HandlerFunc) HandlerFunc
```

Middleware wraps a handler to add functionality.

**Example:**
```go
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        log.Printf("%s %s", c.Method(), c.Path())
        return next(c)
    }
}
```

## Route Options

### WithName

```go
func WithName(name string) Option
```

Assign a name to a route for generating type-safe URL helpers.

**Example:**
```go
r.Get("/users/:id", showUser, router.WithName("users_show"))
```

### WithMiddleware

```go
func WithMiddleware(middleware ...MiddlewareFunc) Option
```

Apply middleware to a specific route.

**Example:**
```go
r.Get("/admin", handler, router.WithMiddleware(requireAdmin))
```

## Resource Options

### Only

```go
func Only(actions ...ResourceAction) ResourceOption
```

Limit resource routes to only the specified actions.

**Example:**
```go
r.Resources("/posts", &PostController{},
    router.Only(router.IndexAction, router.ShowAction))
```

### Except

```go
func Except(actions ...ResourceAction) ResourceOption
```

Exclude specific actions from resource routes.

**Example:**
```go
r.Resources("/posts", &PostController{},
    router.Except(router.NewAction, router.EditAction))
```

### WithResourceMiddleware

```go
func WithResourceMiddleware(middleware ...MiddlewareFunc) ResourceOption
```

Apply middleware to all routes in a resource.

**Example:**
```go
r.Resources("/posts", &PostController{},
    router.WithResourceMiddleware(authMiddleware))
```

## Resource Actions

```go
const (
    IndexAction  ResourceAction = "index"   // GET /resources
    NewAction    ResourceAction = "new"     // GET /resources/new
    CreateAction ResourceAction = "create"  // POST /resources
    ShowAction   ResourceAction = "show"    // GET /resources/:id
    EditAction   ResourceAction = "edit"    // GET /resources/:id/edit
    UpdateAction ResourceAction = "update"  // PATCH/PUT /resources/:id
    DeleteAction ResourceAction = "delete"  // DELETE /resources/:id
)
```

## Controller Interface

```go
type ResourceController interface {
    Index(c *Context) error   // GET    /resources
    New(c *Context) error     // GET    /resources/new
    Create(c *Context) error  // POST   /resources
    Show(c *Context) error    // GET    /resources/:id
    Edit(c *Context) error    // GET    /resources/:id/edit
    Update(c *Context) error  // PATCH  /resources/:id
    Delete(c *Context) error  // DELETE /resources/:id
}
```

Controllers can implement any subset of these methods.

**Example:**
```go
type PostController struct{}

func (pc *PostController) Index(c *router.Context) error {
    return c.JSON(200, getAllPosts())
}

func (pc *PostController) Show(c *router.Context) error {
    id := c.Param("id")
    return c.JSON(200, findPost(id))
}
```

## Server Options

### WithPort

```go
func WithPort(port int) ServerOption
```

Specify the port for the HTTP server.

**Example:**
```go
r.Serve(router.WithPort(8080))
```

### WithGenerateHelpers

```go
func WithGenerateHelpers(generate bool) ServerOption
```

Enable or disable automatic route helper generation.

**Example:**
```go
r.Serve(router.WithGenerateHelpers(false))
```

### WithRoutesPackage

```go
func WithRoutesPackage(pkg string) ServerOption
```

Specify the package name for generated route helpers.

**Example:**
```go
r.Serve(router.WithRoutesPackage("api"))
```

### WithRoutesOutputFile

```go
func WithRoutesOutputFile(file string) ServerOption
```

Specify the output file path for generated route helpers.

**Example:**
```go
r.Serve(router.WithRoutesOutputFile("api/routes.go"))
```

## Complete pkg.go.dev Reference

For the most up-to-date and detailed API documentation, visit:

[pkg.go.dev/github.com/douglasgreyling/router](https://pkg.go.dev/github.com/douglasgreyling/router)
