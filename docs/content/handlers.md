---
title: Handlers
---

## Introduction

Handlers are functions that process incoming HTTP requests and generate responses. If you're familiar with the MVC (Model View Controller) design pattern, think of handlers as "actions."

For larger applications, we recommend organizing your handlers within controllers. It keeps things structured and maintainable. That said, for simple or specialized routes, standalone handlers work perfectly fine. RESTful routes typically map to controller actions, which we'll explore in detail below.

## Creating a Handler with a Specific Function

The simplest way to define a handler is with a function that has the signature `func(*router.Context) error`. Your handler receives a `Context` object, which gives you access to request data and convenient methods for sending responses.

```go
r.Get("/hello", func(c *router.Context) error {
    return c.String(http.StatusOK, "Hello, World!")
})
```

You can also define handlers as separate functions for better organization:

```go
func showWelcome(c *router.Context) error {
    return c.HTML(http.StatusOK, "<h1>Welcome!</h1>")
}

func listProducts(c *router.Context) error {
    products := getProducts()
    return c.JSON(http.StatusOK, products)
}

func main() {
    r := router.New()
    r.Get("/", showWelcome)
    r.Get("/products", listProducts)
    r.Serve()
}
```

## Creating Handlers via Controllers

Controllers give you a structured way to organize related handlers. Think of a controller as a simple struct with methods that match the `HandlerFunc` signature. This approach really shines when you're building RESTful resources.

```go
type UserController struct {}

func (uc *UserController) Index(c *router.Context) error {
    users := []string{"Alice", "Bob", "Charlie"}
    return c.JSON(http.StatusOK, users)
}

func (uc *UserController) Show(c *router.Context) error {
    id := c.Param("id")
    user := map[string]string{"id": id, "name": "Alice"}
    return c.JSON(http.StatusOK, user)
}

func (uc *UserController) Create(c *router.Context) error {
    var data map[string]string
    if err := c.BindJSON(&data); err != nil {
        return err
    }
    return c.JSON(http.StatusCreated, data)
}

func main() {
    r := router.New()
    userController := &UserController{}

    // Register controller methods as handlers
    r.Get("/users", userController.Index)
    r.Get("/users/:id", userController.Show)
    r.Post("/users", userController.Create)

    r.Serve()
}
```

## Creating RESTful Resource Handlers

Here's where things get really convenient. Instead of manually registering each route one by one, you can use `Resources()` to automatically create all your RESTful routes at once. This method maps standard HTTP verbs and paths to controller actions, just like Rails does.

When you call `r.Resources("/posts", &PostController{})`, the router automatically registers up to 7 routes for you:

```go
type PostController struct{}

// Index - GET /posts - List all posts
func (pc *PostController) Index(c *router.Context) error {
    posts := getAllPosts()
    return c.JSON(http.StatusOK, posts)
}

// New - GET /posts/new - Show form to create a new post
func (pc *PostController) New(c *router.Context) error {
    return c.HTML(http.StatusOK, "<form>Create new post</form>")
}

// Create - POST /posts - Create a new post
func (pc *PostController) Create(c *router.Context) error {
    var post Post
    if err := c.BindJSON(&post); err != nil {
        return err
    }
    created := createPost(post)
    return c.JSON(http.StatusCreated, created)
}

// Show - GET /posts/:id - Show a specific post
func (pc *PostController) Show(c *router.Context) error {
    id := c.Param("id")
    post := getPost(id)
    return c.JSON(http.StatusOK, post)
}

// Edit - GET /posts/:id/edit - Show form to edit a post
func (pc *PostController) Edit(c *router.Context) error {
    id := c.Param("id")
    post := getPost(id)
    return c.HTML(http.StatusOK, "<form>Edit post "+id+"</form>")
}

// Update - PATCH/PUT /posts/:id - Update a specific post
func (pc *PostController) Update(c *router.Context) error {
    id := c.Param("id")
    var post Post
    if err := c.BindJSON(&post); err != nil {
        return err
    }
    updated := updatePost(id, post)
    return c.JSON(http.StatusOK, updated)
}

// Delete - DELETE /posts/:id - Delete a specific post
func (pc *PostController) Delete(c *router.Context) error {
    id := c.Param("id")
    deletePost(id)
    return c.NoContent(http.StatusNoContent)
}

func main() {
    r := router.New()

    // This single line creates all 7 routes!
    r.Resources("/posts", &PostController{})

    r.Serve()
}
```

**This creates the following routes automatically:**

| Method      | Path             | Controller Method | Route Name    |
|-------------|------------------|-------------------|---------------|
| GET         | /posts           | Index()           | posts_index   |
| GET         | /posts/new       | New()             | posts_new     |
| POST        | /posts           | Create()          | posts_create  |
| GET         | /posts/:id       | Show()            | posts_show    |
| GET         | /posts/:id/edit  | Edit()            | posts_edit    |
| PATCH/PUT   | /posts/:id       | Update()          | posts_update  |
| DELETE      | /posts/:id       | Delete()          | posts_delete  |

### Best Practices for Resource Handlers

**Always name your controller methods according to the standard RESTful actions:**
- `Index` — lists all resources
- `New` — shows the create form
- `Create` — creates a new resource
- `Show` — displays a single resource
- `Edit` — shows the edit form
- `Update` — updates an existing resource
- `Delete` — deletes a resource

This naming convention makes your code predictable and easier to understand. When another developer sees `Index()`, they immediately know it's listing resources. No guesswork needed!

### Limiting Resource Actions

You don't always need all seven actions. Maybe you're building a read-only API, or you don't need HTML forms. Use `Only()` or `Except()` to specify exactly which routes you want:

```go
// Only create routes for listing and showing posts (read-only API)
r.Resources("/posts", &PostController{},
    router.Only(router.IndexAction, router.ShowAction))

// Create all routes except New and Edit (API without HTML forms)
r.Resources("/posts", &PostController{},
    router.Except(router.NewAction, router.EditAction))
```

If you don't use `Only()` or `Except()`, you must implement all seven methods or the router will panic.

## Working with Request Data

The `Context` object provides several helpful methods for accessing request data. Let's explore each one.

### Route Parameters

Route parameters are defined in your path using the `:name` syntax. They're perfect for capturing dynamic values like user IDs or post slugs:

```go
r.Get("/users/:id", func(c *router.Context) error {
    id := c.Param("id")
    return c.String(http.StatusOK, "User ID: %s", id)
})

r.Get("/users/:userID/posts/:postID", func(c *router.Context) error {
    userID := c.Param("userID")
    postID := c.Param("postID")
    return c.JSON(http.StatusOK, map[string]string{
        "user_id": userID,
        "post_id": postID,
    })
})

// Wildcard parameters capture everything after the prefix
r.Get("/files/*filepath", func(c *router.Context) error {
    filepath := c.Param("filepath")
    // filepath might be "docs/readme.txt" for /files/docs/readme.txt
    return c.String(http.StatusOK, "File: %s", filepath)
})
```

**Important:** Parameter names must be unique within a route. The following will cause a panic:

```go
// ❌ Invalid: duplicate parameter name "id"
r.Get("/users/:id/posts/:id", handler)

// ✓ Valid: unique parameter names
r.Get("/users/:userID/posts/:postID", handler)
```

### Query Parameters

Query parameters (like `?page=2&limit=10`) can be accessed using `Query()` or `QueryDefault()`. Use `Query()` when you need to know if a parameter exists, or `QueryDefault()` to provide a fallback value:

```go
r.Get("/search", func(c *router.Context) error {
    // Get query parameter with existence check
    if q, ok := c.Query("q"); ok {
        // Parameter exists
        return c.String(http.StatusOK, "Searching for: %s", q)
    }
    return c.String(http.StatusBadRequest, "Missing search query")
})

r.Get("/items", func(c *router.Context) error {
    // Get query parameter with default value
    page := c.QueryDefault("page", "1")
    limit := c.QueryDefault("limit", "10")

    return c.JSON(http.StatusOK, map[string]string{
        "page":  page,
        "limit": limit,
    })
})
```

### Request Body

You'll often need to parse JSON from POST or PUT requests. The router makes this easy:

```go
// Bind JSON to a struct
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

r.Post("/users", func(c *router.Context) error {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        return err
    }

    // Use req.Name and req.Email
    return c.JSON(http.StatusCreated, req)
})

// Read raw body
r.Post("/webhook", func(c *router.Context) error {
    body, err := c.Body()
    if err != nil {
        return err
    }

    // Process body bytes
    return c.NoContent(http.StatusOK)
})
```

### Headers

Read incoming headers and set outgoing ones. Here's a common pattern for API authentication:

```go
r.Get("/api/data", func(c *router.Context) error {
    // Read request header
    authToken := c.Header("Authorization")
    if authToken == "" {
        return c.JSON(http.StatusUnauthorized,
            map[string]string{"error": "Missing authorization"})
    }

    // Set response header
    c.SetHeader("X-API-Version", "1.0")
    c.SetHeader("X-Request-ID", "12345")

    return c.JSON(http.StatusOK, map[string]string{"data": "value"})
})
```

### Context Store

The context store lets you stash request-scoped values—super useful for passing data between middleware and handlers:

```go
r.Get("/profile", func(c *router.Context) error {
    // Set a value
    c.Set("user_id", 123)
    c.Set("username", "alice")

    // Get a value
    if userID, ok := c.GetInt("user_id"); ok {
        // Type-safe retrieval
    }

    if username, ok := c.GetString("username"); ok {
        // Type-safe retrieval
    }

    // Generic retrieval
    if val, ok := c.Get("user_id"); ok {
        // val is interface{}
    }

    return c.NoContent(http.StatusOK)
})
```

## Responding to Requests

Your handlers need to send responses back to clients. The router's `Context` provides several convenient methods for different response types. Let's explore each one.

### JSON Responses

JSON is the most common format for APIs. Use `c.JSON()` to send JSON responses:

```go
r.Get("/users/:id", func(c *router.Context) error {
    user := map[string]interface{}{
        "id":    c.Param("id"),
        "name":  "Alice",
        "email": "alice@example.com",
    }
    return c.JSON(http.StatusOK, user)
})

// With structs
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

r.Get("/users/:id", func(c *router.Context) error {
    user := User{
        ID:    123,
        Name:  "Bob",
        Email: "bob@example.com",
    }
    return c.JSON(http.StatusOK, user)
})
```

The `JSON()` method automatically:
- Sets the `Content-Type` header to `application/json`
- Encodes your data as JSON
- Sends the response with the status code you specify

### Plain Text Responses

For simple text responses, use `c.String()`:

```go
r.Get("/hello", func(c *router.Context) error {
    return c.String(http.StatusOK, "Hello, World!")
})

// With formatting
r.Get("/greet/:name", func(c *router.Context) error {
    name := c.Param("name")
    return c.String(http.StatusOK, "Hello, %s!", name)
})
```

The `String()` method sets `Content-Type` to `text/plain` and supports `fmt.Sprintf`-style formatting.

### HTML Responses

Sending HTML is just as easy with `c.HTML()`:

```go
r.Get("/page", func(c *router.Context) error {
    html := `
        <!DOCTYPE html>
        <html>
        <head><title>My Page</title></head>
        <body>
            <h1>Welcome!</h1>
            <p>This is a sample page.</p>
        </body>
        </html>
    `
    return c.HTML(http.StatusOK, html)
})

// With dynamic content
r.Get("/profile/:name", func(c *router.Context) error {
    name := c.Param("name")
    html := fmt.Sprintf(`
        <html>
        <body>
            <h1>Profile: %s</h1>
        </body>
        </html>
    `, name)
    return c.HTML(http.StatusOK, html)
})
```

This sets `Content-Type` to `text/html; charset=utf-8`.

### Raw Data Responses

Need to send binary data or a custom content type? Use `c.Data()`:

```go
// Send an image
r.Get("/image", func(c *router.Context) error {
    imageData, err := os.ReadFile("photo.jpg")
    if err != nil {
        return err
    }
    return c.Data(http.StatusOK, "image/jpeg", imageData)
})

// Send a PDF
r.Get("/document.pdf", func(c *router.Context) error {
    pdfData := generatePDF()
    return c.Data(http.StatusOK, "application/pdf", pdfData)
})

// Send custom data
r.Get("/data", func(c *router.Context) error {
    data := []byte("custom data")
    return c.Data(http.StatusOK, "application/octet-stream", data)
})
```

The `Data()` method gives you complete control over the content type and raw byte data.

### No Content Responses

Sometimes you just need to acknowledge a request without sending a body—like after a successful DELETE operation:

```go
r.Delete("/users/:id", func(c *router.Context) error {
    id := c.Param("id")
    deleteUser(id)
    return c.NoContent(http.StatusNoContent)
})

// 200 OK with no body
r.Post("/acknowledge", func(c *router.Context) error {
    processInBackground()
    return c.NoContent(http.StatusOK)
})
```

This is cleaner than sending an empty response and signals your intent clearly.

### Redirects

Redirect clients to different URLs with `c.Redirect()`:

```go
// Temporary redirect (302)
r.Get("/old-page", func(c *router.Context) error {
    return c.Redirect(http.StatusFound, "/new-page")
})

// Permanent redirect (301)
r.Get("/old-url", func(c *router.Context) error {
    return c.Redirect(http.StatusMovedPermanently, "/new-url")
})

// Redirect after form submission (303)
r.Post("/submit", func(c *router.Context) error {
    // Process form...
    return c.Redirect(http.StatusSeeOther, "/success")
})
```

### Setting Custom Headers

Add custom headers to your responses:

```go
r.Get("/api/data", func(c *router.Context) error {
    // Set custom headers
    c.SetHeader("X-API-Version", "2.0")
    c.SetHeader("X-Rate-Limit", "100")
    c.SetHeader("Cache-Control", "max-age=3600")

    return c.JSON(http.StatusOK, data)
})
```

### Streaming Responses

For large files or streaming data, you can write directly to the response:

```go
r.Get("/stream", func(c *router.Context) error {
    c.SetHeader("Content-Type", "text/event-stream")
    c.SetHeader("Cache-Control", "no-cache")
    c.SetHeader("Connection", "keep-alive")

    // Write directly to the response
    for i := 0; i < 10; i++ {
        fmt.Fprintf(c.Writer, "data: Message %d\n\n", i)
        c.Writer.(http.Flusher).Flush()
        time.Sleep(1 * time.Second)
    }

    return nil
})
```

### File Downloads

Serve files for download with proper headers:

```go
r.Get("/download/:filename", func(c *router.Context) error {
    filename := c.Param("filename")
    filepath := "/path/to/files/" + filename

    // Read the file
    data, err := os.ReadFile(filepath)
    if err != nil {
        return c.JSON(http.StatusNotFound,
            map[string]string{"error": "File not found"})
    }

    // Set headers for download
    c.SetHeader("Content-Disposition",
        fmt.Sprintf("attachment; filename=%s", filename))
    c.SetHeader("Content-Type", "application/octet-stream")

    return c.Data(http.StatusOK, "application/octet-stream", data)
})
```

### Combining Response Methods

You can build complex responses by combining methods:

```go
r.Get("/api/users", func(c *router.Context) error {
    users := getAllUsers()

    // Add pagination headers
    c.SetHeader("X-Total-Count", strconv.Itoa(len(users)))
    c.SetHeader("X-Page", "1")
    c.SetHeader("X-Per-Page", "20")

    // Add CORS headers
    c.SetHeader("Access-Control-Allow-Origin", "*")

    // Send JSON response
    return c.JSON(http.StatusOK, users)
})
```

## Working with Cookies

Cookies are essential for sessions, authentication, and user preferences. The router provides simple methods for reading and setting them.

### Reading Cookies

```go
r.Get("/dashboard", func(c *router.Context) error {
    // Get a cookie by name
    cookie, err := c.Cookie("session_id")
    if err != nil {
        // Cookie doesn't exist or error occurred
        return c.JSON(http.StatusUnauthorized,
            map[string]string{"error": "Not authenticated"})
    }

    sessionID := cookie.Value
    // Use the session ID to look up user session

    return c.String(http.StatusOK, "Welcome back! Session: %s", sessionID)
})
```

### Setting Cookies

```go
import (
    "net/http"
    "time"
)

r.Post("/login", func(c *router.Context) error {
    // Authenticate user...

    // Create and set a cookie
    cookie := &http.Cookie{
        Name:     "session_id",
        Value:    "abc123xyz",
        Path:     "/",
        MaxAge:   3600,           // 1 hour in seconds
        HttpOnly: true,           // Not accessible via JavaScript
        Secure:   true,           // Only sent over HTTPS
        SameSite: http.SameSiteStrictMode,
    }
    c.SetCookie(cookie)

    return c.JSON(http.StatusOK, map[string]string{
        "message": "Logged in successfully",
    })
})

r.Post("/logout", func(c *router.Context) error {
    // Delete a cookie by setting MaxAge to -1
    cookie := &http.Cookie{
        Name:   "session_id",
        Value:  "",
        Path:   "/",
        MaxAge: -1, // Expire immediately
    }
    c.SetCookie(cookie)

    return c.JSON(http.StatusOK, map[string]string{
        "message": "Logged out successfully",
    })
})
```

### Cookie Options

When setting cookies, you can configure various options:

- **Name**: The cookie's identifier
- **Value**: The cookie's value (stored as a string)
- **Path**: URL path that must exist for the cookie to be sent (default: `/`)
- **Domain**: Domain where the cookie is valid
- **Expires**: Absolute expiration date/time
- **MaxAge**: Relative expiration time in seconds (use `-1` to delete)
- **HttpOnly**: If `true`, cookie is inaccessible to JavaScript
- **Secure**: If `true`, cookie only sent over HTTPS
- **SameSite**: Controls cross-site cookie behavior:
  - `http.SameSiteStrictMode`: Never sent in cross-site requests
  - `http.SameSiteLaxMode`: Sent with top-level navigations
  - `http.SameSiteNoneMode`: Sent with all requests (requires `Secure: true`)

```go
// Example: Setting a persistent cookie
r.Get("/remember", func(c *router.Context) error {
    cookie := &http.Cookie{
        Name:     "remember_token",
        Value:    generateToken(),
        Path:     "/",
        Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    c.SetCookie(cookie)

    return c.String(http.StatusOK, "Cookie set!")
})
```

## Route Helpers

Route helpers are one of the router's best features. They're automatically generated, type-safe functions that help you build URLs. No more manual string concatenation like `"/users/" + userID + "/posts/" + postID"` (which is error-prone and fragile). The router generates helpers that ensure your URLs are always correct and match your actual routes.

### How Route Helper Generation Works

When you start your server with `r.Serve()`, the router does some magic behind the scenes:

1. **Scans all your routes** — both explicitly named routes and resource routes
2. **Generates a Go file** — creates `routes/generated.go` with helper functions
3. **Creates type-safe functions** — each route gets a function that accepts exactly the right parameters
4. **Catches broken links at compile time** — if you reference a route that doesn't exist, your code won't compile

This happens automatically in development mode (when `ROUTER_ENV` isn't set to `"production"`).

### Why Route Helpers Are Awesome

**1. Type Safety**

Route helpers are checked at compile time. Try to use a route that doesn't exist? Your code won't even compile. It's like having a safety net that catches mistakes before they reach production:

```go
// ❌ This won't compile if posts_show route doesn't exist
url := routes.PostsShowPath(postID)

// ✓ You'll know immediately if you misspell a route name
url := routes.PosstsShowPath(postID) // Compile error!
```

**2. Automatic Parameter Handling**

The generated functions accept the exact parameters needed for the route:

```go
// For route: /users/:id
url := routes.UsersShowPath("123")  // Returns: "/users/123"

// For route: /users/:userID/posts/:postID
url := routes.UserPostShowPath("42", "99")  // Returns: "/users/42/posts/99"

// No parameters for simple routes: /about
url := routes.AboutPath()  // Returns: "/about"
```

**3. Refactoring Safety**

Here's where route helpers really shine: when you change a route's path, the generated helper automatically updates. No hunting through your codebase to find hardcoded URLs:

```go
// Before: r.Get("/articles/:id", handler, router.WithName("article_show"))
url := routes.ArticleShowPath(id)  // Works!

// After changing route: r.Get("/blog/:id", handler, router.WithName("article_show"))
url := routes.ArticleShowPath(id)  // Still works! Now returns "/blog/:id"
```

**4. Prevents Typos and String Concatenation Errors**

```go
// ❌ Error-prone: manual string building
url := "/users/" + userID + "/posts/" + postID  // Easy to make mistakes

// ✓ Safe: generated helper
url := routes.UserPostShowPath(userID, postID)  // Always correct
```

### Using Route Helpers in Your Handlers

```go
import "yourapp/routes"

func (uc *UserController) Show(c *router.Context) error {
    userID := c.Param("id")

    // Generate links to other routes
    editURL := routes.UsersEditPath(userID)
    indexURL := routes.UsersIndexPath()

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id": userID,
        "edit_url": editURL,
        "list_url": indexURL,
    })
}
```

### Naming Your Routes

The router automatically names your routes based on their path and HTTP method, but you can override this with custom names when it makes sense:

```go
// Automatic naming: "users_show"
r.Get("/users/:id", handler)

// Custom naming: "profile"
r.Get("/users/:id", handler, router.WithName("profile"))
// Generates: routes.ProfilePath(id)

// Resources automatically create named routes
r.Resources("/posts", &PostController{})
// Generates: posts_index, posts_show, posts_create, etc.
```

### Generated File Example

Here's what the router generates in `routes/generated.go`:

```go
package routes

// UsersIndexPath generates the path for users_index route
// GET /users
func UsersIndexPath() string {
    return "/users"
}

// UsersShowPath generates the path for users_show route
// GET /users/:id
func UsersShowPath(id string) string {
    return "/users/" + id
}

// UserPostsShowPath generates the path for user_posts_show route
// GET /users/:userID/posts/:postID
func UserPostsShowPath(userID string, postID string) string {
    return "/users/" + userID + "/posts/" + postID
}
```

### Configuration Options

You can customize route helper generation:

```go
// Disable generation in development
r.Serve(router.WithGenerateHelpers(false))

// Force generation in production
r.Serve(router.WithGenerateHelpers(true))

// Custom package name and output location
r.Serve(
    router.WithRoutesPackage("api"),
    router.WithRoutesOutputFile("api/routes.go"),
)
```

### Best Practices

1. **Always use route helpers** instead of hardcoded strings
2. **Name your routes explicitly** when the auto-generated name isn't clear
3. **Commit the generated file** to version control so it's available in production
4. **Import the routes package** in files where you need to generate URLs

```go
import "yourapp/routes"

// ✓ Good: Type-safe and refactoring-friendly
redirectURL := routes.PostsShowPath(post.ID)
c.Redirect(http.StatusSeeOther, redirectURL)

// ❌ Bad: Error-prone and breaks when routes change
redirectURL := "/posts/" + post.ID
c.Redirect(http.StatusSeeOther, redirectURL)
```

## Error Handling

All handlers return an `error` type. This simple design choice makes error handling consistent across your entire application and lets the router manage how errors are processed and returned to clients centrally.

### How Error Handling Works

When your handler returns an error, the router's `ErrorHandler` function kicks in automatically. The default error handler does three things:

1. Checks if response headers have already been sent to the client
2. If headers are sent, logs the error (can't modify the response at this point)
3. If headers haven't been sent, returns a JSON error response with status 500

```go
// Default ErrorHandler (built into the router)
r.ErrorHandler = func(c *router.Context, err error) {
    if c.IsHeaderWritten() {
        fmt.Fprintf(os.Stderr, "Error after headers sent: %v\n", err)
        return
    }
    c.JSON(http.StatusInternalServerError, map[string]string{
        "error": err.Error(),
    })
}
```

### Returning Errors from Handlers

The simplest approach is to return errors directly:

```go
r.Get("/users/:id", func(c *router.Context) error {
    id := c.Param("id")

    user, err := database.GetUser(id)
    if err != nil {
        return err  // ErrorHandler will process this
    }

    return c.JSON(http.StatusOK, user)
})
```

### Custom Error Types

Custom error types give you fine-grained control over error responses. Here's a practical example that you can adapt for your needs:

```go
// Define custom error types
type AppError struct {
    Status  int
    Message string
    Code    string
}

func (e *AppError) Error() string {
    return e.Message
}

// Helper functions to create specific errors
func NotFoundError(message string) *AppError {
    return &AppError{
        Status:  http.StatusNotFound,
        Message: message,
        Code:    "NOT_FOUND",
    }
}

func ValidationError(message string) *AppError {
    return &AppError{
        Status:  http.StatusBadRequest,
        Message: message,
        Code:    "VALIDATION_ERROR",
    }
}

func UnauthorizedError(message string) *AppError {
    return &AppError{
        Status:  http.StatusUnauthorized,
        Message: message,
        Code:    "UNAUTHORIZED",
    }
}

// Use custom errors in handlers
r.Get("/posts/:id", func(c *router.Context) error {
    id := c.Param("id")

    post, err := database.GetPost(id)
    if err != nil {
        return NotFoundError("Post not found")
    }

    return c.JSON(http.StatusOK, post)
})
```

### Custom Error Handler

Customize the global error handler to handle your custom error types:

```go
func main() {
    r := router.New()

    // Custom error handler
    r.ErrorHandler = func(c *router.Context, err error) {
        // Can't modify response if headers already sent
        if c.IsHeaderWritten() {
            log.Printf("Error after headers sent: %v", err)
            return
        }

        // Handle custom AppError type
        if appErr, ok := err.(*AppError); ok {
            c.JSON(appErr.Status, map[string]interface{}{
                "error": appErr.Message,
                "code":  appErr.Code,
            })
            return
        }

        // Handle standard errors
        log.Printf("Internal error: %v", err)
        c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Internal server error",
        })
    }

    // Register routes...
    r.Serve()
}
```

### Validation Errors

Validation is a common task. Here's a pattern for handling validation errors with helpful, structured responses:

```go
type ValidationErrorResponse struct {
    Message string            `json:"message"`
    Errors  map[string]string `json:"errors"`
}

func (uc *UserController) Create(c *router.Context) error {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return &AppError{
            Status:  http.StatusBadRequest,
            Message: "Invalid JSON",
            Code:    "INVALID_JSON",
        }
    }

    // Validate fields
    errors := make(map[string]string)
    if user.Email == "" {
        errors["email"] = "Email is required"
    }
    if user.Name == "" {
        errors["name"] = "Name is required"
    }
    if len(user.Password) < 8 {
        errors["password"] = "Password must be at least 8 characters"
    }

    if len(errors) > 0 {
        // Return validation errors
        return c.JSON(http.StatusBadRequest, ValidationErrorResponse{
            Message: "Validation failed",
            Errors:  errors,
        })
    }

    // Create user...
    created := createUser(user)
    return c.JSON(http.StatusCreated, created)
}
```

### Error Handling Best Practices

Here are some patterns that'll make your error handling cleaner and more maintainable.

**1. Return errors early**

```go
// ✓ Good: Early return
func (uc *UserController) Show(c *router.Context) error {
    id := c.Param("id")

    user, err := database.GetUser(id)
    if err != nil {
        return NotFoundError("User not found")
    }

    return c.JSON(http.StatusOK, user)
}

// ❌ Bad: Nested conditions
func (uc *UserController) Show(c *router.Context) error {
    id := c.Param("id")

    user, err := database.GetUser(id)
    if err == nil {
        return c.JSON(http.StatusOK, user)
    } else {
        return NotFoundError("User not found")
    }
}
```

**2. Use descriptive error messages:**

```go
// ✓ Good: Descriptive
return NotFoundError("Post with ID " + id + " not found")

// ❌ Bad: Vague
return NotFoundError("Not found")
```

**3. Don't expose internal errors to clients:**

```go
// ✓ Good: Hide internal details
r.ErrorHandler = func(c *router.Context, err error) {
    if c.IsHeaderWritten() {
        return
    }

    // Log the actual error for debugging
    log.Printf("Error: %v", err)

    // Send generic message to client
    c.JSON(http.StatusInternalServerError, map[string]string{
        "error": "Internal server error",
    })
}

// ❌ Bad: Exposes internal details
return fmt.Errorf("database connection failed: %v", dbErr)
```

**4. Handle errors at the right level:**

```go
// ✓ Good: Handle specific errors in handler, generic in ErrorHandler
r.Get("/posts/:id", func(c *router.Context) error {
    id := c.Param("id")

    post, err := database.GetPost(id)
    if err == sql.ErrNoRows {
        return NotFoundError("Post not found")
    }
    if err != nil {
        return err  // Let ErrorHandler handle unexpected errors
    }

    return c.JSON(http.StatusOK, post)
})
```

**5. Use middleware for common error scenarios:**

```go
// Authentication middleware
func AuthMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *router.Context) error {
        token := c.Header("Authorization")
        if token == "" {
            return UnauthorizedError("Missing authorization token")
        }

        user, err := validateToken(token)
        if err != nil {
            return UnauthorizedError("Invalid token")
        }

        c.Set("user", user)
        return next(c)
    }
}

// Use the middleware
r.Get("/profile", userController.Profile,
    router.WithMiddleware(AuthMiddleware))
```
