package info

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool implements the package info tool
type Tool struct {
	tools.BaseTool
}

// New creates a new info tool
func New() *Tool {
	tool := mcp.NewTool("package_info",
		mcp.WithDescription("Get detailed information about a specific package"),
		mcp.WithString("package",
			mcp.Required(),
			mcp.Description("The exact package name"),
		),
	)

	return &Tool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

// GetHandler returns the handler function for the info tool
func (t *Tool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		packageName := request.Params.Arguments["package"].(string)
		pkg := repo.GetPackageInfo(packageName)

		if pkg == nil {
			return mcp.NewToolResultText(fmt.Sprintf("Package '%s' not found.", packageName)), nil
		}

		// Format the package details as JSON for a more structured response
		details, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error formatting package details: %v", err)), nil
		}

		return mcp.NewToolResultText(string(details)), nil
	}
}
