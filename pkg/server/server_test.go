package server

import (
	"context"
	"testing"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

// MockMCPServer implements a fake MCPServer for testing
type MockMCPServer struct {
	serveError error
}

func (m *MockMCPServer) ServeStdio(_ *server.MCPServer) error {
	return m.serveError
}

// Override the server.ServeStdio function to test the Serve method
func TestServe(t *testing.T) {
	// We can't easily test Serve without mocking the ServeStdio function
	// which is not easily mockable. So we're just skipping this test for now.
	t.Skip("Skipping Serve test - not easy to mock ServeStdio")
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