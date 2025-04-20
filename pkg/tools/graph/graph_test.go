package graph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGraphTool(t *testing.T) {
	// Create tool
	tool := New()
	
	// Check tool name
	if tool.GetTool().Name != "package_graph" {
		t.Errorf("Expected tool name to be 'package_graph', got '%s'", tool.GetTool().Name)
	}
	
	// Create mock repository with dependency relationships
	mockPackages := []*apk.Package{
		{
			Name: "base-package", 
			Version: "1.0.0", 
			Dependencies: []string{"lib-package=2.0.0"},
		},
		{
			Name: "lib-package", 
			Version: "2.0.0",
			Provides: []string{"lib-capability=2.0"},
		},
		{
			Name: "app-package", 
			Version: "3.0.0",
			Dependencies: []string{"lib-package>=1.5.0", "base-package"},
		},
	}
	repo := apkindex.NewRepository(mockPackages)
	
	// Get handler
	handler := tool.GetHandler(repo)
	
	// Test the findPackagesProviding function directly
	t.Run("findPackagesProviding", func(t *testing.T) {
		// Find packages providing a capability that exists
		providers := findPackagesProviding(repo, "lib-capability")
		if len(providers) != 1 || providers[0] != "lib-package" {
			t.Errorf("Expected [lib-package], got %v", providers)
		}
		
		// Find packages providing a capability that doesn't exist
		providers = findPackagesProviding(repo, "nonexistent-capability")
		if len(providers) != 0 {
			t.Errorf("Expected empty slice, got %v", providers)
		}
		
		// Find packages by name match (implicit provides)
		providers = findPackagesProviding(repo, "lib-package")
		if len(providers) != 1 || providers[0] != "lib-package" {
			t.Errorf("Expected [lib-package], got %v", providers)
		}
	})
	
	// Test different query types
	testCases := []struct {
		name              string
		pkg               string
		queryType         string
		depth             string
		checkText         string
		expectedErrorFlag bool
	}{
		{
			name:      "requires",
			pkg:       "base-package",
			queryType: "requires",
			checkText: "lib-package",
		},
		{
			name:      "provides",
			pkg:       "lib-package",
			queryType: "provides",
			checkText: "lib-capability",
		},
		{
			name:      "depends_on",
			pkg:       "app-package",
			queryType: "depends_on",
			checkText: "base-package",
		},
		{
			name:      "depends_on with depth",
			pkg:       "app-package",
			queryType: "depends_on",
			depth:     "3",
			checkText: "base-package",
		},
		{
			name:      "required_by",
			pkg:       "lib-package",
			queryType: "required_by",
			checkText: "base-package",
		},
		{
			name:      "what_provides",
			pkg:       "lib-capability",
			queryType: "what_provides",
			checkText: "Packages that provide lib-capability",
		},
		{
			name:              "invalid query type",
			pkg:               "base-package",
			queryType:         "invalid_type",
			checkText:         "Unknown query type",
			expectedErrorFlag: true,
		},
		{
			name:              "depends_on with invalid depth",
			pkg:               "app-package",
			queryType:         "depends_on",
			depth:             "invalid",
			checkText:         "base-package",
			expectedErrorFlag: true,
		},
		{
			name:              "package not found",
			pkg:               "nonexistent-package",
			queryType:         "depends_on",
			checkText:         "not found",
			expectedErrorFlag: false,
		},
		{
			name:              "what_provides with package that doesn't exist",
			pkg:               "nonexistent-capability", 
			queryType:         "what_provides",
			checkText:         "No packages found that provide",
			expectedErrorFlag: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := mcp.CallToolRequest{}
			req.Params.Name = "package_graph"
			args := map[string]interface{}{
				"package":    tc.pkg,
				"query_type": tc.queryType,
			}
			
			// Add depth parameter if specified
			if tc.depth != "" {
				args["depth"] = tc.depth
			}
			
			req.Params.Arguments = args
			
			result, err := handler(context.Background(), req)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}
			
			// Verify error flag is as expected
			if result.IsError != tc.expectedErrorFlag {
				t.Fatalf("Expected IsError=%v, got %v", tc.expectedErrorFlag, result.IsError)
			}
			
			// Skip further checks for error responses
			if result.IsError {
				return
			}
			
			// Verify we have content
			if len(result.Content) == 0 {
				t.Errorf("Expected non-empty result content")
			}
			
			// Convert to JSON to check for expected text
			jsonData, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("Failed to marshal result: %v", err)
			}
			
			jsonStr := string(jsonData)
			if !strings.Contains(jsonStr, tc.checkText) {
				t.Errorf("Expected result to contain '%s', got: %s", tc.checkText, jsonStr)
			}
		})
	}
}