# Router

A Rails-inspired, fast, flexible HTTP router for Go with automatic route helper generation.

## Features

- **Fast routing** - Radix tree-based routing with parameter and wildcard support
- **Type-safe route helpers** - Automatically generated functions for building URLs
- **RESTful resources** - Rails-inspired resource scaffolding with conventional routing
- **Flexible middleware** - Apply at global, group, or route level with predictable execution order
- **Rich Context API** - Convenient methods for handling requests and responses
- **Zero dependencies** - Built on Go's standard library

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/douglasgreyling/router"
)

func main() {
    r := router.New()

    r.Get("/", func(c *router.Context) error {
        return c.String(http.StatusOK, "Hello, World!")
    })

    r.Get("/users/:id", func(c *router.Context) error {
        id := c.Param("id")
        return c.JSON(http.StatusOK, map[string]string{
            "id": id,
        })
    })

    r.Serve() // Starts server on localhost:3000
}
```

## RESTful Resources

Create full CRUD routes with a single line:

```go
type PostController struct{}

func (pc *PostController) Index(c *router.Context) error {
    return c.JSON(http.StatusOK, getAllPosts())
}

func (pc *PostController) Show(c *router.Context) error {
    id := c.Param("id")
    return c.JSON(http.StatusOK, getPost(id))
}

// ... other CRUD methods

func main() {
    r := router.New()
    r.Resources("/posts", &PostController{})
    r.Serve()
}
```

## Route Helpers

The router automatically generates type-safe functions for building URLs:

```go
import "yourapp/routes"

// Generated functions like:
url := routes.PostsShowPath("123")      // Returns: "/posts/123"
url := routes.UsersIndexPath()          // Returns: "/users"
```

## Installation

```bash
go get github.com/douglasgreyling/router
```

## Documentation

For complete documentation, guides, and examples, visit:

**📚 [https://douglasgreyling.github.io/router/](https://douglasgreyling.github.io/router/)**

Topics covered:
- [Getting Started](https://douglasgreyling.github.io/router/getting-started/)
- [Handlers](https://douglasgreyling.github.io/router/handlers/)
- [Middleware](https://douglasgreyling.github.io/router/middleware/)
- [Configuration](https://douglasgreyling.github.io/router/configuration/)

## License

MIT