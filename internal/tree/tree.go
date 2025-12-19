package tree

import (
	"fmt"
	"strings"
)

// NodeType represents the type of node in the radix tree
type NodeType uint8

const (
	Static   NodeType = iota // static route segment
	Param                    // :param - matches a single segment
	Wildcard                 // *wildcard - matches everything after
)

// Node represents a node in the radix tree
type Node struct {
	// The path segment this node represents
	Path string

	// Type of node (static, param, wildcard)
	NType NodeType

	// The full pattern if this node ends a route
	Pattern string

	// Handlers for different HTTP methods (stored as interface{})
	Handlers map[string]interface{}

	// Child nodes
	Children []*Node

	// Parameter name if this is a param or wildcard node
	ParamName string

	// Middleware chain for this specific route (stored as []interface{})
	Middleware []interface{}
}

// Tree manages route trees for each HTTP method
type Tree struct {
	roots map[string]*Node
}

// New creates a new Tree instance
func New() *Tree {
	return &Tree{
		roots: make(map[string]*Node),
	}
}

// AddRoute adds a route to the radix tree
func (t *Tree) AddRoute(method, path string, handler interface{}, middleware []interface{}) error {
	if len(path) == 0 || path[0] != '/' {
		return fmt.Errorf("invalid route path %q for %s: path must begin with '/'", path, method)
	}

	// Ensure root node exists for this method
	if t.roots[method] == nil {
		t.roots[method] = &Node{
			Path:     "/",
			Handlers: make(map[string]interface{}),
			Children: make([]*Node, 0),
		}
	}

	root := t.roots[method]

	if path == "/" {
		root.Handlers[method] = handler
		root.Pattern = path
		root.Middleware = middleware
		return nil
	}

	// Remove leading and trailing slashes, split path
	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")

	// Validate no duplicate parameter names
	paramNames := make(map[string]int)
	for i, segment := range segments {
		if len(segment) > 0 && (segment[0] == ':' || segment[0] == '*') {
			paramName := segment[1:]
			if firstIndex, exists := paramNames[paramName]; exists {
				return fmt.Errorf("duplicate parameter %q in route %s /%s: first occurrence at segment %d, duplicate at segment %d", paramName, method, path, firstIndex, i)
			}
			paramNames[paramName] = i
		}
	}

	current := root
	for i, segment := range segments {
		// Determine node type
		nType := Static
		paramName := ""

		if len(segment) > 0 {
			if segment[0] == ':' {
				nType = Param
				paramName = segment[1:]
			} else if segment[0] == '*' {
				nType = Wildcard
				paramName = segment[1:]
			}
		}

		// Look for existing child with matching segment
		var next *Node
		for _, child := range current.Children {
			if child.Path == segment && child.NType == nType {
				next = child
				break
			}
		}

		// Create new node if no match found
		if next == nil {
			next = &Node{
				Path:      segment,
				NType:     nType,
				ParamName: paramName,
				Handlers:  make(map[string]interface{}),
				Children:  make([]*Node, 0),
			}
			current.Children = append(current.Children, next)
		}

		// If this is the last segment, set the handler
		if i == len(segments)-1 {
			next.Handlers[method] = handler
			next.Pattern = "/" + strings.Join(segments, "/")
			next.Middleware = middleware
		}

		current = next
	}

	return nil
}

// Find finds a matching route in the tree and returns handler, params, and middleware
func (t *Tree) Find(method, path string) (interface{}, map[string]string, []interface{}) {
	root := t.roots[method]
	if root == nil {
		return nil, nil, nil
	}

	if path == "/" {
		if handler, ok := root.Handlers[method]; ok {
			return handler, nil, root.Middleware
		}
		return nil, nil, nil
	}

	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")
	params := make(map[string]string)

	handler, middleware := search(root, segments, 0, params, method)
	return handler, params, middleware
}

// search recursively searches for a matching route
func search(n *Node, segments []string, index int, params map[string]string, method string) (interface{}, []interface{}) {
	// If we've matched all segments, check if this node has a handler
	if index == len(segments) {
		if handler, ok := n.Handlers[method]; ok {
			return handler, n.Middleware
		}
		return nil, nil
	}

	segment := segments[index]

	// Try children in order: static > param > wildcard
	for _, child := range n.Children {
		switch child.NType {
		case Static:
			if child.Path == segment {
				if handler, middleware := search(child, segments, index+1, params, method); handler != nil {
					return handler, middleware
				}
			}
		case Param:
			params[child.ParamName] = segment
			if handler, middleware := search(child, segments, index+1, params, method); handler != nil {
				return handler, middleware
			}
			delete(params, child.ParamName) // backtrack
		case Wildcard:
			// Wildcard matches everything remaining
			params[child.ParamName] = strings.Join(segments[index:], "/")
			if handler, ok := child.Handlers[method]; ok {
				return handler, child.Middleware
			}
		}
	}

	return nil, nil
}

// HasMethod checks if any HTTP method has a handler for the given path
func (t *Tree) HasMethod(path string) bool {
	for method := range t.roots {
		handler, _, _ := t.Find(method, path)
		if handler != nil {
			return true
		}
	}
	return false
}

// GetMethods returns all HTTP methods that have handlers for the given path
func (t *Tree) GetMethods(path string) []string {
	methods := make([]string, 0)
	for method := range t.roots {
		handler, _, _ := t.Find(method, path)
		if handler != nil {
			methods = append(methods, method)
		}
	}
	return methods
}
