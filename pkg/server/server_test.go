package server

import (
	"context"
	"testing"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/search"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestAddTool tests the tool registration
func TestAddTool(t *testing.T) {
	// Create a server
	srv := New(DefaultConfig())
	
	// Create a mock repo
	mockRepo := apkindex.NewRepository(nil)
	
	// Create and register a tool
	searchTool := search.New()
	
	// Add the tool to the server
	srv.AddTool(searchTool.GetTool(), searchTool.GetHandler(mockRepo))
	
	// We can't easily verify the internal state of the server,
	// but at least we can verify that the code doesn't panic
}

// mockTool is used for testing
type mockTool struct {
	tools.BaseTool
}

func newMockTool() *mockTool {
	tool := mcp.NewTool("mock_tool",
		mcp.WithDescription("A mock tool for testing"),
	)
	
	return &mockTool{
		BaseTool: tools.BaseTool{Tool: tool},
	}
}

func (t *mockTool) GetHandler(repo *apkindex.Repository) tools.ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Mock tool response"), nil
	}
}

// TestRegisterAll tests registering multiple tools
func TestRegisterAll(t *testing.T) {
	// Create a server
	srv := New(DefaultConfig())
	
	// Create a mock repo
	mockRepo := apkindex.NewRepository(nil)
	
	// Create multiple tools
	mockTool1 := newMockTool()
	mockTool2 := newMockTool()
	
	// Register all tools
	tools.RegisterAll(srv, mockRepo, mockTool1, mockTool2)
	
	// Again, we can't easily verify the internal state,
	// but we can verify the code executes without errors
}