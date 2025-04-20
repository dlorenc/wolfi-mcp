package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/server"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/dependencies"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/info"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/search"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/versions"
)

func main() {
	// Get the absolute path to the APKINDEX file
	absPath, err := filepath.Abs("APKINDEX.tar.gz")
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Create a new index loader
	loader := &apkindex.FileIndexLoader{}

	// Load the APK index
	packages, err := loader.LoadIndex(absPath)
	if err != nil {
		fmt.Printf("Error loading APK index: %v\n", err)
		os.Exit(1)
	}

	// Create a new repository with the loaded packages
	repo := apkindex.NewRepository(packages)

	// Create a new server with default configuration
	srv := server.New(server.DefaultConfig())

	// Create all tools
	allTools := []tools.Tool{
		search.New(),
		info.New(),
		dependencies.New(),
		versions.New(),
	}

	// Register all tools with the server
	tools.RegisterAll(srv, repo, allTools...)

	// Start the server
	if err := srv.Serve(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
