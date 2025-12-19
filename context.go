package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// responseWriter wraps http.ResponseWriter to track response state
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader captures the status code and tracks that headers were written
func (w *responseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures WriteHeader is called and tracks that response started
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Status returns the HTTP status code that was written
func (w *responseWriter) Status() int {
	return w.status
}

// Context provides a convenient interface for handling HTTP requests and responses.
// It wraps http.ResponseWriter and *http.Request with helper methods for common tasks
// like sending JSON, parsing parameters, setting headers, and storing request-scoped values.
//
// Key features:
//   - Access route parameters: c.Param("id")
//   - Read query parameters: c.Query("page")
//   - Send responses: c.JSON(), c.String(), c.HTML()
//   - Store request-scoped data: c.Set()/c.Get()
//   - Access raw request/response: c.Request, c.Writer
//
// Example:
//
//	func handler(c *router.Context) error {
//	    // Get route parameter
//	    id := c.Param("id")
//	
//	    // Get query parameter with existence check
//	    if page, ok := c.Query("page"); ok {
//	        // parameter exists
//	    }
//	
//	    // Store and retrieve values
//	    c.Set("user_id", 123)
//	    if userID, ok := c.GetInt("user_id"); ok {
//	        // value exists and is an int
//	    }
//	
//	    // Send JSON response
//	    return c.JSON(200, map[string]string{"id": id})
//	}
type Context struct {
	Writer  *responseWriter
	Request *http.Request
	Params  Params
	store   map[string]interface{}
	index   int // for middleware chain
}

// newContext creates a new Context instance
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  &responseWriter{ResponseWriter: w, status: http.StatusOK},
		Request: r,
		Params:  make(Params),
		store:   make(map[string]interface{}),
		index:   -1,
	}
}

// IsHeaderWritten returns true if response headers have been sent to the client.
// Once headers are written, the status code and headers cannot be changed.
//
// This is primarily useful when implementing custom error handlers to avoid
// attempting to modify the response after headers have been sent:
//
//	r.ErrorHandler = func(c *Context, err error) {
//	    if c.IsHeaderWritten() {
//	        // Log error, can't modify response
//	        return
//	    }
//	    c.JSON(500, map[string]string{"error": err.Error()})
//	}
func (c *Context) IsHeaderWritten() bool {
	return c.Writer.wroteHeader
}

// Param returns a route parameter by name
func (c *Context) Param(name string) string {
	return c.Params[name]
}

// Query returns a URL query parameter by name.
// Returns (value, true) if the parameter exists, or ("", false) if it doesn't.
func (c *Context) Query(name string) (string, bool) {
	values := c.Request.URL.Query()
	if !values.Has(name) {
		return "", false
	}
	return values.Get(name), true
}

// QueryDefault returns a URL query parameter or a default value if it doesn't exist.
func (c *Context) QueryDefault(name, defaultValue string) string {
	if value, ok := c.Query(name); ok {
		return value
	}
	return defaultValue
}

// Set stores a value in the context
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Get retrieves a value from the context.
// Returns (value, true) if the key exists, or (nil, false) if it doesn't.
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.store[key]
	return val, ok
}

// GetString retrieves a string value from the context.
// Returns (value, true) if the key exists and is a string, or ("", false) otherwise.
func (c *Context) GetString(key string) (string, bool) {
	val, ok := c.store[key].(string)
	return val, ok
}

// GetInt retrieves an int value from the context.
// Returns (value, true) if the key exists and is an int, or (0, false) otherwise.
func (c *Context) GetInt(key string) (int, bool) {
	val, ok := c.store[key].(int)
	return val, ok
}

// GetBool retrieves a bool value from the context.
// Returns (value, true) if the key exists and is a bool, or (false, false) otherwise.
func (c *Context) GetBool(key string) (bool, bool) {
	val, ok := c.store[key].(bool)
	return val, ok
}

// JSON sends a JSON response
func (c *Context) JSON(status int, data interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(data)
}

// String sends a plain text response
func (c *Context) String(status int, format string, values ...interface{}) error {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(status)
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// HTML sends an HTML response
func (c *Context) HTML(status int, html string) error {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(html))
	return err
}

// Data sends raw bytes as response
func (c *Context) Data(status int, contentType string, data []byte) error {
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write(data)
	return err
}

// NoContent sends a response with no body
func (c *Context) NoContent(status int) error {
	c.Writer.WriteHeader(status)
	return nil
}

// Redirect sends a redirect response
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status > 308 {
		return fmt.Errorf("invalid redirect status code: %d", status)
	}
	http.Redirect(c.Writer, c.Request, url, status)
	return nil
}

// BindJSON binds JSON request body to a struct
func (c *Context) BindJSON(obj interface{}) error {
	if c.Request.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(obj)
}

// Body returns the request body as bytes
func (c *Context) Body() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// Method returns the HTTP method
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the request path
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Header returns a request header value
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Cookie returns a cookie by name
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie sets a cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// Status sets the response status code
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// GetStatus returns the HTTP status code that was written (or will be written).
// Returns 200 if no status has been explicitly set.
func (c *Context) GetStatus() int {
	return c.Writer.Status()
}

// ClientIP returns the client's IP address
func (c *Context) ClientIP() string {
	// Check X-Forwarded-For header first
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// Check X-Real-IP header
	if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fall back to RemoteAddr
	return c.Request.RemoteAddr
}
