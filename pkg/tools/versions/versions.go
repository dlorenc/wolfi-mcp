package versions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool implements the package versions tool
type Tool struct {
	tools.BaseTool
}

// New creates a new versions tool
func New() *Tool {
	tool := mcp.NewTool("compare_versions",
		mcp.WithDescription("Compare versions of packages"),
		mcp.WithString("package",
			mcp.Required(),
			mcp.Description("The package name to compare versions for"),
		),
	)

	return &Tool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

// GetHandler returns the handler function for the versions tool
func (t *Tool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		packageName := request.Params.Arguments["package"].(string)
		versions := repo.GetPackageVersions(packageName)
		
		if len(versions) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No versions found for package '%s'.", packageName)), nil
		}

		// Sort versions
		sort.Slice(versions, func(i, j int) bool {
			// This is a simplified version comparison
			// For proper semantic versioning, consider using a dedicated package
			return versions[i].Version < versions[j].Version
		})

		// Format the versions
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Versions of %s:\n\n", packageName))
		
		for i, pkg := range versions {
			sb.WriteString(fmt.Sprintf("%d. Version: %s\n", i+1, pkg.Version))
			sb.WriteString(fmt.Sprintf("   Architecture: %s\n", pkg.Arch))
			sb.WriteString(fmt.Sprintf("   Size: %d bytes\n", pkg.Size))
			if pkg.Origin != "" {
				sb.WriteString(fmt.Sprintf("   Origin: %s\n", pkg.Origin))
			}
			sb.WriteString("\n")
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}