---
title: Configuration
---

## Introduction

The router is designed to work great out of the box with sensible defaults, but it also gives you plenty of options to customize its behavior. Most configuration happens when you call `Serve()` to start your server.

Let's explore the different ways you can configure the router to fit your needs.

## Starting the Server

The simplest way to start your server is just calling `Serve()` with no arguments:

```go
func main() {
    r := router.New()

    // Register your routes...
    r.Get("/", homeHandler)

    // Start the server with defaults
    r.Serve()
}
```

This starts your server on port `:3000` and automatically generates route helpers in development mode. Easy!

## Server Port

Want to use a different port? Use the `WithPort()` option:

```go
// Listen on port 8080
r.Serve(router.WithPort(":8080"))

// Listen on port 80
r.Serve(router.WithPort(":80"))

// Use a different host and port
r.Serve(router.WithPort("0.0.0.0:3000"))
```

**Default:** `:3000`

## Route Helper Generation

The router automatically generates type-safe route helper functions that make building URLs easier. By default, this happens in development but not in production.

### Default Behavior

The router checks the `ROUTER_ENV` environment variable:

- **Development** (default): Generates route helpers automatically
- **Production** (`ROUTER_ENV=production`): Skips generation

```go
// In development, this generates routes/generated.go
r.Serve()

// In production with ROUTER_ENV=production, this skips generation
r.Serve()
```

### Disabling Generation in Development

Want to skip generation even in development? Use `WithGenerateHelpers(false)`:

```go
// Explicitly disable route helper generation
r.Serve(router.WithGenerateHelpers(false))
```

This is useful if:
- You're not using route helpers
- Generation is slow with many routes
- You're in a CI/CD pipeline and want faster builds

### Forcing Generation in Production

Need route helpers generated in production? Enable it explicitly:

```go
// Force generation even in production
r.Serve(router.WithGenerateHelpers(true))
```

**Note:** Most production deployments should commit the generated file to version control instead of generating at runtime. This makes deployments faster and more predictable.

## Route Helpers Output Location

By default, route helpers are generated to `routes/generated.go` with package name `routes`. You can customize both the package name and output location.

### Custom Package Name

Change the package name with `WithRoutesPackage()`:

```go
// Generate to 'api' package instead of 'routes'
r.Serve(router.WithRoutesPackage("api"))
```

This generates:
```go
package api

func UsersIndexPath() string {
    return "/users"
}
// ... more helper functions
```

**Default:** `routes`

### Custom Output Location

Change where the file is generated with `WithRoutesOutputFile()`:

```go
// Generate to a different directory
r.Serve(router.WithRoutesOutputFile("internal/routes/helpers.go"))

// Generate to project root
r.Serve(router.WithRoutesOutputFile("routes.go"))
```

**Default:** `routes/generated.go`

### Combining Package and Output Options

You can use both options together:

```go
r.Serve(
    router.WithRoutesPackage("api"),
    router.WithRoutesOutputFile("internal/api/routes.go"),
)
```

This generates `internal/api/routes.go` with `package api`.

## Combining Configuration Options

All configuration options can be combined. The router uses functional options, so the order doesn't matter:

```go
// Combine multiple options
r.Serve(
    router.WithPort(":8080"),
    router.WithGenerateHelpers(true),
    router.WithRoutesPackage("api"),
    router.WithRoutesOutputFile("api/routes.go"),
)

// Same thing, different order
r.Serve(
    router.WithRoutesPackage("api"),
    router.WithPort(":8080"),
    router.WithRoutesOutputFile("api/routes.go"),
    router.WithGenerateHelpers(true),
)
```

## Custom Error Handlers

You can customize how the router handles different error scenarios:

### 404 Not Found

Override the default 404 handler:

```go
r := router.New()

r.NotFound = func(c *router.Context) error {
    return c.JSON(http.StatusNotFound, map[string]string{
        "error": "The page you're looking for doesn't exist",
        "path":  c.Path(),
    })
}
```

### 405 Method Not Allowed

Override the method not allowed handler:

```go
r.MethodNotAllowed = func(c *router.Context) error {
    return c.JSON(http.StatusMethodNotAllowed, map[string]string{
        "error":  "Method not allowed",
        "method": c.Method(),
        "path":   c.Path(),
    })
}
```

### Global Error Handler

Customize how errors from handlers are processed:

```go
r.ErrorHandler = func(c *router.Context, err error) {
    // Can't modify response if headers already sent
    if c.IsHeaderWritten() {
        log.Printf("Error after headers sent: %v", err)
        return
    }

    // Handle different error types
    if appErr, ok := err.(*AppError); ok {
        c.JSON(appErr.Status, map[string]interface{}{
            "error": appErr.Message,
            "code":  appErr.Code,
        })
        return
    }

    // Log unexpected errors
    log.Printf("Internal error: %v", err)

    // Return generic error to client
    c.JSON(http.StatusInternalServerError, map[string]string{
        "error": "Internal server error",
    })
}
```