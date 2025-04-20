package server

import (
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Config holds server configuration options
type Config struct {
	Name    string
	Version string
}

// DefaultConfig returns a default server configuration
func DefaultConfig() Config {
	return Config{
		Name:    "Alpine Package Database",
		Version: "1.0.0",
	}
}

// Server represents the MCP server for the package database
type Server struct {
	config Config
	server *server.MCPServer
}

// New creates a new Server instance
func New(config Config) *Server {
	mcpServer := server.NewMCPServer(
		config.Name,
		config.Version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	return &Server{
		config: config,
		server: mcpServer,
	}
}

// AddTool adds a tool and its handler to the server
func (s *Server) AddTool(tool mcp.Tool, handler tools.ToolHandler) {
	s.server.AddTool(tool, server.ToolHandlerFunc(handler))
}

// Serve starts the server to handle stdin/stdout communication
func (s *Server) Serve() error {
	return server.ServeStdio(s.server)
}