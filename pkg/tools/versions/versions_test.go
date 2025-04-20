package versions

import (
	"context"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestVersionsTool(t *testing.T) {
	// Create tool
	tool := New()
	
	// Check tool name
	if tool.GetTool().Name != "compare_versions" {
		t.Errorf("Expected tool name to be 'compare_versions', got '%s'", tool.GetTool().Name)
	}
	
	// Create mock repository with multiple versions
	mockPackages := []*apk.Package{
		{
			Name: "alpine-base", 
			Version: "3.15.0", 
			Arch: "x86_64",
			Size: 1024,
		},
		{
			Name: "alpine-base", 
			Version: "3.14.0", 
			Arch: "x86_64",
			Size: 1000,
		},
		{
			Name: "alpine-base", 
			Version: "3.16.0", 
			Arch: "aarch64",
			Size: 1048,
			Origin: "alpine",
		},
	}
	repo := apkindex.NewRepository(mockPackages)
	
	// Get handler
	handler := tool.GetHandler(repo)
	
	// Test with multiple versions
	req := mcp.CallToolRequest{}
	req.Params.Name = "compare_versions"
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
	req.Params.Name = "compare_versions"
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
}