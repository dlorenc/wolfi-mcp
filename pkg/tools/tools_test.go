package tools

import (
	"context"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/mark3labs/mcp-go/mcp"
)

// mockServer is used for testing tools registration
type mockServer struct {
	addedTools map[string]bool
}

func newMockServer() *mockServer {
	return &mockServer{
		addedTools: make(map[string]bool),
	}
}

// AddTool implements the interface required by RegisterAll
func (s *mockServer) AddTool(tool mcp.Tool, handler ToolHandler) {
	s.addedTools[tool.Name] = true
}

// mockTool is used for testing
type mockTool struct {
	BaseTool
	name string
}

func newMockTool(name string) *mockTool {
	tool := mcp.NewTool(name,
		mcp.WithDescription("A mock tool for testing"),
	)
	
	return &mockTool{
		BaseTool: BaseTool{Tool: tool},
		name:     name,
	}
}

func (t *mockTool) GetHandler(repo *apkindex.Repository) ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Mock tool response for " + t.name), nil
	}
}

func TestBaseTool(t *testing.T) {
	toolName := "test-tool"
	mcpTool := mcp.NewTool(toolName)
	baseTool := BaseTool{Tool: mcpTool}
	
	got := baseTool.GetTool()
	if got.Name != toolName {
		t.Errorf("GetTool().Name = %q, want %q", got.Name, toolName)
	}
}

func TestRegisterAll(t *testing.T) {
	// Create mock server and repository
	srv := newMockServer()
	repo := apkindex.NewRepository([]*apk.Package{})
	
	// Create mock tools
	tool1 := newMockTool("tool1")
	tool2 := newMockTool("tool2")
	tool3 := newMockTool("tool3")
	
	// Register all tools
	RegisterAll(srv, repo, tool1, tool2, tool3)
	
	// Verify all tools were registered
	for _, name := range []string{"tool1", "tool2", "tool3"} {
		if !srv.addedTools[name] {
			t.Errorf("Tool %q was not registered", name)
		}
	}
}