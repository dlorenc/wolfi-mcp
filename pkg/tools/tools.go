package tools

import (
	"context"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolHandler is the type for tool handler functions
type ToolHandler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)

// Tool defines the interface for MCP tools
type Tool interface {
	// GetTool returns the MCP tool definition
	GetTool() mcp.Tool

	// GetHandler returns the handler function for the tool
	GetHandler(repo *apkindex.Repository) ToolHandler
}

// BaseTool provides common functionality for all tools
type BaseTool struct {
	Tool mcp.Tool
}

// GetTool returns the MCP tool definition
func (b *BaseTool) GetTool() mcp.Tool {
	return b.Tool
}

// RegisterAll registers all available tools with the server
func RegisterAll(srv interface {
	AddTool(mcp.Tool, ToolHandler)
}, repo *apkindex.Repository, tools ...Tool) {
	for _, tool := range tools {
		srv.AddTool(tool.GetTool(), tool.GetHandler(repo))
	}
}
