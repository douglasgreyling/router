---
title: Getting Started
---

# Getting Started

## Installation

Install the router package using Go modules:

```bash
go get github.com/douglasgreyling/router
```

## Basic Usage

Create a simple HTTP server with routing:

```go
package main

import (
    "github.com/douglasgreyling/router"
)

func main() {
    r := router.New()

    r.Get("/", func(c *router.Context) error {
        return c.String(200, "Hello, World!")
    })

    r.Get("/users/:id", func(c *router.Context) error {
        id := c.Param("id")
        return c.JSON(200, map[string]string{
            "id": id,
            "message": "User found",
        })
    })

    r.Serve() // Starts server on :3000
}
```

## Route Parameters

The router supports two types of dynamic segments:

### Named Parameters

Named parameters match a single path segment:

```go
r.Get("/users/:id", func(c *router.Context) error {
    id := c.Param("id")
    return c.String(200, "User ID: " + id)
})

// Matches: /users/123
// Does NOT match: /users/123/posts
```

### Wildcards

Wildcards match everything after the prefix:

```go
r.Get("/files/*filepath", func(c *router.Context) error {
    filepath := c.Param("filepath")
    return c.String(200, "File: " + filepath)
})

// Matches: /files/docs/readme.txt
// Matches: /files/images/photo.jpg
```

## HTTP Methods

All standard HTTP methods are supported:

```go
r.Get("/posts", listPosts)
r.Post("/posts", createPost)
r.Put("/posts/:id", updatePost)
r.Patch("/posts/:id", patchPost)
r.Delete("/posts/:id", deletePost)
```

## Response Helpers

The Context provides convenient methods for sending responses:

```go
// JSON response
r.Get("/json", func(c *router.Context) error {
    return c.JSON(200, map[string]string{"status": "ok"})
})

// String response
r.Get("/text", func(c *router.Context) error {
    return c.String(200, "Plain text response")
})

// HTML response
r.Get("/html", func(c *router.Context) error {
    return c.HTML(200, "<h1>Hello</h1>")
})
```

## Error Handling

Handlers return errors, which are processed by the ErrorHandler:

```go
r.Get("/users/:id", func(c *router.Context) error {
    user, err := findUser(c.Param("id"))
    if err != nil {
        return err  // ErrorHandler will process this
    }
    return c.JSON(200, user)
})
```

Customize error handling:

```go
r.ErrorHandler = func(c *router.Context, err error) {
    if c.IsHeaderWritten() {
        log.Printf("Error after headers sent: %v", err)
        return
    }
    c.JSON(500, map[string]string{"error": err.Error()})
}
```

## Server Configuration

Customize server options:

```go
r.Serve(
    router.WithPort(":8080"),
)
```
