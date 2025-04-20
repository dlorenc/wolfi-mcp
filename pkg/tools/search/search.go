package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool implements the package search tool
type Tool struct {
	tools.BaseTool
}

// New creates a new search tool
func New() *Tool {
	tool := mcp.NewTool("search_packages",
		mcp.WithDescription("Search for packages in the Alpine package database"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The package name to search for (supports partial matches)"),
		),
	)

	return &Tool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

// GetHandler returns the handler function for the search tool
func (t *Tool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := strings.ToLower(request.Params.Arguments["query"].(string))
		results := repo.Search(query)
		
		if len(results) == 0 {
			return mcp.NewToolResultText("No packages found matching your query."), nil
		}

		// Format the results as text
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d packages matching '%s':\n\n", len(results), query))
		
		for i, pkg := range results {
			sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, pkg.Name, pkg.Version))
			if pkg.Description != "" {
				sb.WriteString(fmt.Sprintf("   Description: %s\n", pkg.Description))
			}
			sb.WriteString("\n")
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}