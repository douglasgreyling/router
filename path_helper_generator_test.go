package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractParameters(t *testing.T) {
	tests := []struct {
		pattern  string
		expected []RouteParam
	}{
		{
			pattern:  "/users/:id",
			expected: []RouteParam{{Name: "id", Type: "string"}},
		},
		{
			pattern:  "/users/:user_id/posts/:post_id",
			expected: []RouteParam{{Name: "user_id", Type: "string"}, {Name: "post_id", Type: "string"}},
		},
		{
			pattern:  "/",
			expected: []RouteParam{},
		},
		{
			pattern:  "/static/path",
			expected: []RouteParam{},
		},
		{
			pattern:  "/:a/:b/:c",
			expected: []RouteParam{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "c", Type: "string"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := extractParameters(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d parameters, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i].Name != tt.expected[i].Name {
					t.Errorf("parameter %d: expected name %q, got %q", i, tt.expected[i].Name, result[i].Name)
				}
				if result[i].Type != tt.expected[i].Type {
					t.Errorf("parameter %d: expected type %q, got %q", i, tt.expected[i].Type, result[i].Type)
				}
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_show", "UserShow"},
		{"user_post_index", "UserPostIndex"},
		{"simple", "Simple"},
		{"api-product-list", "ApiProductList"},
		{"mixed_case-test", "MixedCaseTest"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMakeParamList(t *testing.T) {
	tests := []struct {
		name     string
		params   []RouteParam
		expected string
	}{
		{
			name:     "no parameters",
			params:   []RouteParam{},
			expected: "",
		},
		{
			name:     "single parameter",
			params:   []RouteParam{{Name: "id", Type: "string"}},
			expected: "id string",
		},
		{
			name:     "multiple parameters",
			params:   []RouteParam{{Name: "user_id", Type: "string"}, {Name: "post_id", Type: "string"}},
			expected: "user_id string, post_id string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeParamList(tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMakeParamNames(t *testing.T) {
	tests := []struct {
		name     string
		params   []RouteParam
		expected string
	}{
		{
			name:     "no parameters",
			params:   []RouteParam{},
			expected: "",
		},
		{
			name:     "single parameter",
			params:   []RouteParam{{Name: "id", Type: "string"}},
			expected: "id",
		},
		{
			name:     "multiple parameters",
			params:   []RouteParam{{Name: "user_id", Type: "string"}, {Name: "post_id", Type: "string"}},
			expected: "user_id, post_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeParamNames(tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPathHelperGeneratorAddRoute(t *testing.T) {
	cg := NewPathHelperGenerator()

	cg.AddRoute("home", "/", "GET")
	cg.AddRoute("user_show", "/users/:id", "GET")
	cg.AddRoute("user_post", "/users/:user_id/posts/:post_id", "GET")

	if len(cg.routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(cg.routes))
	}

	// Check first route
	if cg.routes[0].Name != "home" {
		t.Errorf("expected route name 'home', got %q", cg.routes[0].Name)
	}
	if len(cg.routes[0].Parameters) != 0 {
		t.Errorf("expected 0 parameters for home route, got %d", len(cg.routes[0].Parameters))
	}

	// Check second route
	if cg.routes[1].Name != "user_show" {
		t.Errorf("expected route name 'user_show', got %q", cg.routes[1].Name)
	}
	if len(cg.routes[1].Parameters) != 1 {
		t.Errorf("expected 1 parameter for user_show route, got %d", len(cg.routes[1].Parameters))
	}
	if cg.routes[1].Parameters[0].Name != "id" {
		t.Errorf("expected parameter name 'id', got %q", cg.routes[1].Parameters[0].Name)
	}

	// Check third route
	if len(cg.routes[2].Parameters) != 2 {
		t.Errorf("expected 2 parameters for user_post route, got %d", len(cg.routes[2].Parameters))
	}
}

func TestPathHelperGeneratorGenerate(t *testing.T) {
	cg := NewPathHelperGenerator()
	cg.AddRoute("home", "/", "GET")
	cg.AddRoute("user_show", "/users/:id", "GET")
	cg.AddRoute("user_post", "/users/:user_id/posts/:post_id", "GET")

	// Create temporary directory
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "routes.go")

	// Generate code
	err := cg.Generate("routes", outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("output file was not created")
	}

	// Read and check content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check for package declaration
	if !strings.Contains(contentStr, "package routes") {
		t.Errorf("generated code missing package declaration")
	}

	// Check for standalone functions
	expectedFunctions := []string{
		"func HomePath(query ...url.Values) string",
		"func UserShowPath(id string, query ...url.Values) string",
		"func UserPostPath(user_id string, post_id string, query ...url.Values) string",
		"func HomeURL(host string, query ...url.Values) string",
		"func UserShowURL(host string, id string, query ...url.Values) string",
		"func UserPostURL(host string, user_id string, post_id string, query ...url.Values) string",
	}

	for _, fn := range expectedFunctions {
		if !strings.Contains(contentStr, fn) {
			t.Errorf("generated code missing function %s", fn)
		}
	}
}

func TestPathHelperGeneratorGenerateWithEmptyRoutes(t *testing.T) {
	cg := NewPathHelperGenerator()

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "routes.go")

	err := cg.Generate("routes", outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// File should not exist when there are no routes
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Errorf("expected no file to be created for empty routes, but file exists")
	}
}

func TestPathHelperGeneratorGenerateComplexRoutes(t *testing.T) {
	cg := NewPathHelperGenerator()
	cg.AddRoute("multi_param", "/foo/:a/bar/:b/baz/:c", "GET")
	cg.AddRoute("api_product", "/api/v1/products/:id", "GET")
	cg.AddRoute("nested_resource", "/orgs/:org_id/teams/:team_id/members/:id", "GET")

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "routes.go")

	err := cg.Generate("routes", outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check for complex route standalone functions with correct parameter counts
	if !strings.Contains(contentStr, "func MultiParamPath(a string, b string, c string, query ...url.Values) string") {
		t.Errorf("generated code has incorrect signature for MultiParamPath")
	}

	if !strings.Contains(contentStr, "func ApiProductPath(id string, query ...url.Values) string") {
		t.Errorf("generated code has incorrect signature for ApiProductPath")
	}

	if !strings.Contains(contentStr, "func NestedResourcePath(org_id string, team_id string, id string, query ...url.Values) string") {
		t.Errorf("generated code has incorrect signature for NestedResourcePath")
	}
}
