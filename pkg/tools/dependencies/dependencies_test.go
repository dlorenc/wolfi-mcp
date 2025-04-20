package dependencies

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestDependenciesTool(t *testing.T) {
	// Create tool
	tool := New()
	
	// Check tool name
	if tool.GetTool().Name != "package_dependencies" {
		t.Errorf("Expected tool name to be 'package_dependencies', got '%s'", tool.GetTool().Name)
	}
	
	// Create mock repository
	mockPackages := []*apk.Package{
		{
			Name: "alpine-base", 
			Version: "3.15.0", 
			Dependencies: []string{"lib-package=2.0.0"},
		},
		{
			Name: "lib-package", 
			Version: "2.0.0",
			Provides: []string{"lib-capability=2.0"},
		},
		{
			Name: "empty-package", 
			Version: "1.0.0",
			Dependencies: []string{},
		},
		{
			Name: "special-dep-package", 
			Version: "1.0.0",
			Dependencies: []string{"so:libssl.so.1.1", "lib-package"},
		},
	}
	repo := apkindex.NewRepository(mockPackages)
	
	// Get handler
	handler := tool.GetHandler(repo)
	
	// Test case with dependencies
	req := mcp.CallToolRequest{}
	req.Params.Name = "package_dependencies"
	req.Params.Arguments = map[string]interface{}{
		"package": "alpine-base",
	}
	
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	
	// Verify it's not an error
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}
	
	// Verify we have content
	if len(result.Content) == 0 {
		t.Errorf("Expected non-empty result content")
	}

	// Test case with no dependencies
	req = mcp.CallToolRequest{}
	req.Params.Name = "package_dependencies"
	req.Params.Arguments = map[string]interface{}{
		"package": "empty-package",
	}
	
	result, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify it's not an error
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}

	// Test case with special dependencies
	req = mcp.CallToolRequest{}
	req.Params.Name = "package_dependencies"
	req.Params.Arguments = map[string]interface{}{
		"package": "special-dep-package",
	}
	
	result, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify it's not an error
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}

	// Test case with nonexistent package
	req = mcp.CallToolRequest{}
	req.Params.Name = "package_dependencies"
	req.Params.Arguments = map[string]interface{}{
		"package": "nonexistent-package",
	}
	
	result, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify it's not an error, but should indicate package not found
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}

	// Convert to JSON to check for "not found"
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "not found") {
		t.Errorf("Expected 'not found' message for nonexistent package, got: %s", jsonStr)
	}
}