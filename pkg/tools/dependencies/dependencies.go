package dependencies

import (
	"context"
	"fmt"
	"strings"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool implements the package dependencies tool
type Tool struct {
	tools.BaseTool
}

// New creates a new dependencies tool
func New() *Tool {
	tool := mcp.NewTool("package_dependencies",
		mcp.WithDescription("List dependencies for a package"),
		mcp.WithString("package",
			mcp.Required(),
			mcp.Description("The exact package name"),
		),
	)

	return &Tool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

// GetHandler returns the handler function for the dependencies tool
func (t *Tool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		packageName := request.Params.Arguments["package"].(string)
		pkg := repo.GetPackageInfo(packageName)

		if pkg == nil {
			return mcp.NewToolResultText(fmt.Sprintf("Package '%s' not found.", packageName)), nil
		}

		// Extract dependencies
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Dependencies for %s (%s):\n\n", pkg.Name, pkg.Version))

		if len(pkg.Dependencies) == 0 {
			sb.WriteString("No dependencies found.\n")
		} else {
			for i, dep := range pkg.Dependencies {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, dep))

				// Try to find if the dependency exists in our index
				depParts := strings.Split(dep, "=")
				depName := depParts[0]

				// Skip special dependencies like "so:lib.so"
				if strings.Contains(depName, ":") {
					continue
				}

				depPkg := repo.GetPackageInfo(depName)
				if depPkg != nil {
					sb.WriteString(fmt.Sprintf("   Available: %s (%s)\n", depPkg.Name, depPkg.Version))
				} else {
					sb.WriteString("   Not found in index\n")
				}
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}
