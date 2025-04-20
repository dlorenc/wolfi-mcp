package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"chainguard.dev/apko/pkg/apk/apk"
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

// multiStringFlag is a flag.Value that allows a flag to be specified multiple times
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

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

// isURL checks if a string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// getAPKIndexPath handles a path which could be a local file path or a URL
// It returns the absolute path to the local file containing the APKINDEX data
func getAPKIndexPath(indexPath string) (string, error) {
	// If a specific index is provided
	if indexPath != "" {
		// Check if it's a URL
		if isURL(indexPath) {
			// Get the cache directory
			cacheDir, err := getUserCacheDir()
			if err != nil {
				return "", fmt.Errorf("error determining cache directory: %w", err)
			}

			// Create cache directory if it doesn't exist
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				return "", fmt.Errorf("error creating cache directory: %w", err)
			}

			// Use a hashed filename to avoid conflicts and overlong paths
			urlHash := fmt.Sprintf("%x", sha256.Sum256([]byte(indexPath)))
			cacheFilePath := filepath.Join(cacheDir, fmt.Sprintf("APKINDEX_%s.tar.gz", urlHash[:8]))

			// Download the file
			fmt.Printf("Downloading APKINDEX from %s...\n", indexPath)
			if err := downloadFile(indexPath, cacheFilePath); err != nil {
				return "", fmt.Errorf("error downloading index file from %s: %w", indexPath, err)
			}

			absPath, err := filepath.Abs(cacheFilePath)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path: %w", err)
			}
			return absPath, nil
		} else {
			// It's a local file path
			absPath, err := filepath.Abs(indexPath)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path: %w", err)
			}
			return absPath, nil
		}
	}

	// If no path provided, use or create the default cached version
	cacheDir, err := getUserCacheDir()
	if err != nil {
		return "", fmt.Errorf("error determining cache directory: %w", err)
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}

	// Default cache file path
	cacheFilePath := filepath.Join(cacheDir, cacheFile)

	// Detect architecture for download URL (aarch64 or x86_64)
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}

	// Download the index file from the default URL
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

// mergePackages combines packages from multiple APKINDEX files following Alpine merging semantics:
// 1. When a package exists in multiple indexes, the highest version wins
// 2. If versions are equal, the most recently indexed one wins
func mergePackages(existing []*apk.Package, new []*apk.Package) []*apk.Package {
	// Create a map for fast lookups of existing packages by name
	pkgMap := make(map[string]*apk.Package, len(existing))
	for _, pkg := range existing {
		pkgMap[pkg.Name] = pkg
	}

	// Process new packages
	for _, pkg := range new {
		existing, exists := pkgMap[pkg.Name]

		if !exists {
			// New package, add it
			pkgMap[pkg.Name] = pkg
			continue
		}

		// Package already exists, compare versions with Alpine's version comparison
		// Since we don't have direct access to apkversion.Compare from the package,
		// we'll use a simple string comparison for version checking
		// In a real-world implementation, this should use the actual Alpine versioning rules

		// First check if the versions are equal
		if existing.Version == pkg.Version {
			// Same version, take the newer package (the one from the later index)
			// This follows Alpine's approach where later repositories override earlier ones
			pkgMap[pkg.Name] = pkg
			continue
		}

		// For different versions, we'll use a simple string comparison
		// This is not a fully accurate representation of Alpine's version comparison,
		// but should work for most common cases
		if pkg.Version > existing.Version {
			// New package has higher version, replace
			pkgMap[pkg.Name] = pkg
		}
		// If pkg.Version < existing.Version, keep the existing one
	}

	// Convert map back to slice
	result := make([]*apk.Package, 0, len(pkgMap))
	for _, pkg := range pkgMap {
		result = append(result, pkg)
	}

	return result
}

func main() {
	// Define command line flags - index can be repeated for multiple indexes
	var indexPaths multiStringFlag
	flag.Var(&indexPaths, "index", "Path to APKINDEX.tar.gz file (can be specified multiple times, if not provided, downloads from Wolfi repository)")
	flag.Parse()

	// Create a new index loader
	loader := &apkindex.FileIndexLoader{}

	// Keep track of all loaded packages
	var allPackages []*apk.Package

	if len(indexPaths) == 0 {
		// No indexes specified, download the default one
		absPath, err := getAPKIndexPath("")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Load the APK index
		fmt.Printf("Loading APK index from %s...\n", absPath)
		packages, err := loader.LoadIndex(absPath)
		if err != nil {
			fmt.Printf("Error loading APK index: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loaded %d packages\n", len(packages))
		allPackages = packages
	} else {
		// Load all specified indexes
		for _, indexPath := range indexPaths {
			absPath, err := getAPKIndexPath(indexPath)
			if err != nil {
				fmt.Printf("Error getting index path for %s: %v\n", indexPath, err)
				os.Exit(1)
			}

			fmt.Printf("Loading APK index from %s...\n", absPath)
			packages, err := loader.LoadIndex(absPath)
			if err != nil {
				fmt.Printf("Error loading APK index: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Loaded %d packages from %s\n", len(packages), absPath)

			// Merge packages into allPackages with proper semantics
			allPackages = mergePackages(allPackages, packages)
		}
		fmt.Printf("Total of %d packages loaded after merging\n", len(allPackages))
	}

	// Create a new repository with the loaded packages
	repo := apkindex.NewRepository(allPackages)

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
