---
layout: home
title: Home
---

# Router

A fast, flexible HTTP router for Go with automatic route helper generation, RESTful resource scaffolding, and type-safe middleware composition.

## Features

- 🚀 **Fast radix tree routing** with parameter and wildcard support
- 🔧 **Automatic code generation** of type-safe route helpers
- 📦 **RESTful resource scaffolding** (Rails-inspired)
- 🎯 **Flexible middleware** with group and route-level composition
- 💎 **Rich Context API** for request/response handling

## Quick Start

```go
package main

import "github.com/douglasgreyling/router"

func main() {
    r := router.New()

    r.Get("/users/:id", func(c *router.Context) error {
        id := c.Param("id")
        return c.JSON(200, map[string]string{"id": id})
    })

    r.Serve() // Starts server on :3000
}
```

## Installation

```bash
go get github.com/douglasgreyling/router
```

## Documentation

- [Getting Started](getting-started) - Installation and basic usage
- [User Guide](guide) - Comprehensive feature guide
- [API Reference](api-reference) - Detailed API documentation
- [Examples](examples) - Real-world examples

## Community

- [GitHub Repository](https://github.com/douglasgreyling/router)
- [Issue Tracker](https://github.com/douglasgreyling/router/issues)
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/douglasgreyling/router)

## License

This project is licensed under the MIT License.
