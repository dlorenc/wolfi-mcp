package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dlorenc/wolfi-mcp/pkg/apkindex"
	"github.com/dlorenc/wolfi-mcp/pkg/server"
	"github.com/dlorenc/wolfi-mcp/pkg/tools"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/dependencies"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/graph"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/info"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/search"
	"github.com/dlorenc/wolfi-mcp/pkg/tools/versions"
)

const (
	defaultWolfiURL = "https://packages.wolfi.dev/os/%s/APKINDEX.tar.gz"
	cacheSubDir     = "wolfi-mcp" // Application-specific subdirectory in the cache
	cacheFile       = "APKINDEX.tar.gz"
)

func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded data: %w", err)
	}

	return nil
}

// getUserCacheDir returns the standard cache directory for the current OS
func getUserCacheDir() (string, error) {
	// Try to use standard directories
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Caches/wolfi-mcp/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		return filepath.Join(homeDir, "Library", "Caches", cacheSubDir), nil
	
	case "windows":
		// Windows: %LOCALAPPDATA%\wolfi-mcp\cache\
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			// Fallback if LOCALAPPDATA is not set
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("could not get user home directory: %w", err)
			}
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, cacheSubDir, "cache"), nil
	
	default:
		// Linux and others: Follow XDG Base Directory Specification
		// Check XDG_CACHE_HOME environment variable first
		xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
		if xdgCacheHome != "" {
			return filepath.Join(xdgCacheHome, cacheSubDir), nil
		}
		
		// Fallback to ~/.cache
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		return filepath.Join(homeDir, ".cache", cacheSubDir), nil
	}
}

func getAPKIndexPath(indexPath string) (string, error) {
	// If a specific path is provided, use it
	if indexPath != "" {
		absPath, err := filepath.Abs(indexPath)
		if err != nil {
			return "", fmt.Errorf("error getting absolute path: %w", err)
		}
		return absPath, nil
	}

	// Get the user's cache directory
	cacheDir, err := getUserCacheDir()
	if err != nil {
		return "", fmt.Errorf("error determining cache directory: %w", err)
	}

	// Otherwise, use or create a cached version
	cacheFilePath := filepath.Join(cacheDir, cacheFile)
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}
	
	// Detect architecture for download URL (aarch64 or x86_64)
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}
	
	// Download the index file
	url := fmt.Sprintf(defaultWolfiURL, arch)
	fmt.Printf("Downloading Wolfi APKINDEX from %s...\n", url)
	
	if err := downloadFile(url, cacheFilePath); err != nil {
		return "", fmt.Errorf("error downloading index file: %w", err)
	}
	
	absPath, err := filepath.Abs(cacheFilePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}
	
	return absPath, nil
}

func main() {
	// Define command line flags
	indexPath := flag.String("index", "", "Path to APKINDEX.tar.gz file (if not provided, downloads from Wolfi repository)")
	flag.Parse()

	// Get the APKINDEX file path (download if needed)
	absPath, err := getAPKIndexPath(*indexPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create a new index loader
	loader := &apkindex.FileIndexLoader{}

	// Load the APK index
	fmt.Printf("Loading APK index from %s...\n", absPath)
	packages, err := loader.LoadIndex(absPath)
	if err != nil {
		fmt.Printf("Error loading APK index: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d packages\n", len(packages))

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
		graph.New(),
	}

	// Register all tools with the server
	tools.RegisterAll(srv, repo, allTools...)

	// Start the server
	fmt.Println("Starting MCP server...")
	if err := srv.Serve(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}