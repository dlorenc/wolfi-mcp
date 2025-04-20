package info

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestInfoTool(t *testing.T) {
	// Create tool
	tool := New()
	
	// Check tool name
	if tool.GetTool().Name != "package_info" {
		t.Errorf("Expected tool name to be 'package_info', got '%s'", tool.GetTool().Name)
	}
	
	// Create mock repository
	mockPackages := []*apk.Package{
		{
			Name: "alpine-base", 
			Version: "3.15.0", 
			Description: "Meta package for Alpine base",
			Dependencies: []string{"alpine-baselayout", "alpine-keys"},
			Arch: "x86_64",
			Size: 1024,
		},
	}
	repo := apkindex.NewRepository(mockPackages)
	
	// Get handler
	handler := tool.GetHandler(repo)
	
	// Test with existing package
	req := mcp.CallToolRequest{}
	req.Params.Name = "package_info"
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
	
	// Test with nonexistent package
	req = mcp.CallToolRequest{}
	req.Params.Name = "package_info"
	req.Params.Arguments = map[string]interface{}{
		"package": "nonexistent-package",
	}
	
	result, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	
	// Verify it's not an error
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}

	// Check for 'not found' message
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "not found") {
		t.Errorf("Expected 'not found' message for nonexistent package, got: %s", jsonStr)
	}
}