package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool implements the package graph tool
type Tool struct {
	tools.BaseTool
}

// New creates a new graph tool
func New() *Tool {
	tool := mcp.NewTool("package_graph",
		mcp.WithDescription("Query the package dependency graph using provides and requires relationships"),
		mcp.WithString("package",
			mcp.Required(),
			mcp.Description("The package name to start the graph query from"),
		),
		mcp.WithString("query_type",
			mcp.Required(),
			mcp.Description("The type of query to perform: 'requires', 'provides', 'depends_on', 'required_by', 'what_provides'"),
		),
		mcp.WithString("depth",
			mcp.Description("Maximum depth of the graph traversal (default: 1)"),
		),
	)

	return &Tool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

// GetHandler returns the handler function for the graph tool
func (t *Tool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		packageName := request.Params.Arguments["package"].(string)
		queryType := request.Params.Arguments["query_type"].(string)

		// Parse depth parameter if provided
		depth := 1
		if depthStr, ok := request.Params.Arguments["depth"]; ok {
			if depthVal, ok := depthStr.(string); ok {
				if _, err := fmt.Sscanf(depthVal, "%d", &depth); err != nil {
					return mcp.NewToolResultError(
						fmt.Sprintf("Invalid depth value '%s': %v", depthVal, err),
					), nil
				}
			}
		}

		// Ensure depth is between 1 and 5 (for performance and readability)
		if depth < 1 {
			depth = 1
		} else if depth > 5 {
			depth = 5
		}

		var sb strings.Builder

		// Special case for what_provides query type
		if strings.ToLower(queryType) == "what_provides" {
			// Shows what packages provide a certain capability
			sb.WriteString(fmt.Sprintf("Packages that provide %s:\n\n", packageName))
			providers := findPackagesProviding(repo, packageName)
			if len(providers) == 0 {
				sb.WriteString("No packages found that provide this capability.\n")
			} else {
				// Sort for predictable output
				sort.Strings(providers)
				for i, provider := range providers {
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, provider))
				}
			}
			return mcp.NewToolResultText(sb.String()), nil
		}

		// For all other query types, get the package
		pkg := repo.GetPackageInfo(packageName)
		if pkg == nil {
			return mcp.NewToolResultText(fmt.Sprintf("Package '%s' not found.", packageName)), nil
		}

		switch strings.ToLower(queryType) {
		case "requires", "dependencies":
			// Shows what dependencies a package has (direct requirements)
			sb.WriteString(fmt.Sprintf("Dependencies required by %s (%s):\n\n", pkg.Name, pkg.Version))
			if len(pkg.Dependencies) == 0 {
				sb.WriteString("No dependencies found.\n")
			} else {
				for i, dep := range pkg.Dependencies {
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, dep))
				}
			}

		case "provides":
			// Shows what capabilities a package provides
			sb.WriteString(fmt.Sprintf("Capabilities provided by %s (%s):\n\n", pkg.Name, pkg.Version))
			if len(pkg.Provides) == 0 {
				sb.WriteString("No explicit provides found.\n")
				sb.WriteString(fmt.Sprintf("This package implicitly provides: %s=%s\n", pkg.Name, pkg.Version))
			} else {
				for i, provide := range pkg.Provides {
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, provide))
				}
			}

		case "depends_on":
			// Recursive dependency graph
			sb.WriteString(fmt.Sprintf("Dependency graph for %s (%s) with depth %d:\n\n", pkg.Name, pkg.Version, depth))
			visited := make(map[string]bool)
			getDependencyGraph(repo, pkg, &sb, "", 0, depth, visited)

		case "required_by":
			// Shows what packages depend on this package
			sb.WriteString(fmt.Sprintf("Packages that depend on %s:\n\n", pkg.Name))
			requiredBy := findPackagesRequiring(repo, pkg.Name)
			if len(requiredBy) == 0 {
				sb.WriteString("No packages found that depend on this package.\n")
			} else {
				// Sort for predictable output
				sort.Strings(requiredBy)
				for i, dep := range requiredBy {
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, dep))
				}
			}

		default:
			return mcp.NewToolResultError(
				fmt.Sprintf("Unknown query type: %s. Supported types: requires, provides, depends_on, required_by, what_provides", queryType),
			), nil
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// getDependencyGraph recursively builds a dependency tree for visualization
func getDependencyGraph(repo *apkindex.Repository, pkg *apk.Package, sb *strings.Builder, prefix string, currentDepth, maxDepth int, visited map[string]bool) {
	if currentDepth > maxDepth {
		return
	}

	// Mark this package as visited to avoid cycles
	visited[pkg.Name] = true

	// Print this package
	if currentDepth == 0 {
		sb.WriteString(fmt.Sprintf("%s (%s)\n", pkg.Name, pkg.Version))
	} else {
		sb.WriteString(fmt.Sprintf("%s├─ %s (%s)\n", prefix, pkg.Name, pkg.Version))
	}

	// Don't continue if we've reached max depth
	if currentDepth == maxDepth {
		return
	}

	// Process dependencies
	childPrefix := prefix + "│  "
	for i, depStr := range pkg.Dependencies {
		// Extract the package name from the dependency string
		depParts := strings.Split(depStr, "=")
		depName := depParts[0]

		// Skip special dependencies like "so:lib.so"
		if strings.Contains(depName, ":") {
			continue
		}

		// Skip if we already visited this package (avoid cycles)
		if visited[depName] {
			sb.WriteString(fmt.Sprintf("%s├─ %s [already visited]\n", childPrefix, depName))
			continue
		}

		// Get the dependency package
		depPkg := repo.GetPackageInfo(depName)

		// If found, recursively process it
		if depPkg != nil {
			getDependencyGraph(repo, depPkg, sb, childPrefix, currentDepth+1, maxDepth, visited)
		} else {
			sb.WriteString(fmt.Sprintf("%s├─ %s [not found in index]\n", childPrefix, depName))
		}

		// Add a new line after each top-level dependency except the last one
		if currentDepth == 0 && i < len(pkg.Dependencies)-1 {
			sb.WriteString("\n")
		}
	}
}

// findPackagesRequiring finds all packages that depend on the given package name
func findPackagesRequiring(repo *apkindex.Repository, packageName string) []string {
	var requiringPackages []string
	allPackages := repo.GetAllPackages()

	for _, pkg := range allPackages {
		for _, dep := range pkg.Dependencies {
			// Extract the package name from the dependency string (handle version constraints)
			depParts := strings.Split(dep, "=")
			depName := depParts[0]

			// Also handle other constraints like >= or >
			if idx := strings.IndexAny(depName, "<>"); idx != -1 {
				depName = strings.TrimSpace(depName[:idx])
			}

			// Skip special dependencies like "so:lib.so"
			if strings.Contains(depName, ":") {
				continue
			}

			if depName == packageName {
				requiringPackages = append(requiringPackages, pkg.Name)
				break
			}
		}
	}

	return requiringPackages
}

// findPackagesProviding finds all packages that provide the given capability
func findPackagesProviding(repo *apkindex.Repository, capability string) []string {
	var providingPackages []string
	allPackages := repo.GetAllPackages()

	for _, pkg := range allPackages {
		// Check if the package name itself matches the capability
		if pkg.Name == capability {
			providingPackages = append(providingPackages, pkg.Name)
			continue
		}

		// Check the provides list
		for _, provide := range pkg.Provides {
			// Extract the capability name (handle version constraints)
			provideParts := strings.Split(provide, "=")
			provideName := provideParts[0]

			if provideName == capability {
				providingPackages = append(providingPackages, pkg.Name)
				break
			}
		}
	}

	return providingPackages
}
