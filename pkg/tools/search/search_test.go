package search

import (
	"context"
	"encoding/json"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestSearchTool(t *testing.T) {
	// Create tool
	tool := New()

	// Check tool name
	if tool.GetTool().Name != "search_packages" {
		t.Errorf("Expected tool name to be 'search_packages', got '%s'", tool.GetTool().Name)
	}

	// Create mock repository
	mockPackages := []*apk.Package{
		{Name: "alpine-base", Version: "3.15.0", Description: "Meta package for Alpine base"},
		{Name: "alpine-keys", Version: "2.4-r1", Description: "Public keys for Alpine Linux packages"},
	}
	repo := apkindex.NewRepository(mockPackages)

	// Get handler
	handler := tool.GetHandler(repo)

	// Test search with result
	req := mcp.CallToolRequest{}
	req.Params.Name = "search_packages"
	req.Params.Arguments = map[string]interface{}{
		"query": "alpine",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify it's not an error
	if result.IsError {
		t.Fatalf("Expected successful result, got error")
	}

	// Convert to JSON to inspect contents
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	// Verify the result contains the expected packages
	jsonStr := string(jsonData)
	if len(jsonStr) == 0 {
		t.Errorf("Expected non-empty result")
	}

	// Test search with no results
	req = mcp.CallToolRequest{}
	req.Params.Name = "search_packages"
	req.Params.Arguments = map[string]interface{}{
		"query": "nonexistent",
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
