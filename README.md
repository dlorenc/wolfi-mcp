# Wolfi MCP Package Database

An MCP (Model Calling Protocol) server for querying and interacting with Alpine-based Linux distribution package databases.

## Features

- Search for packages by name (partial matching supported)
- Get detailed information about specific packages
- List dependencies for packages
- Compare versions of packages
- Query the package dependency graph with different relationship types:
  - What a package requires
  - What capabilities a package provides
  - Recursive dependency graphs
  - Reverse dependency lookup
  - Capability provider lookup

## Requirements

- Go 1.24 or higher

## Installation

```bash
# Clone the repository
git clone https://github.com/dlorenc/wolfi-mcp.git
cd wolfi-mcp

# Build the server
go build -o mcp-server
```

## Usage

The server reads from standard input and writes to standard output following the MCP protocol.

```bash
# Start the server with default settings (downloads the latest APKINDEX from Wolfi)
./mcp-server

# Use a specific APKINDEX file
./mcp-server -index /path/to/your/APKINDEX.tar.gz
```

### Available Tools

The server provides the following tools:

1. **search_packages** - Search for packages in the database
   - Parameter: `query` - The package name to search for (supports partial matches)

2. **package_info** - Get detailed information about a specific package
   - Parameter: `package` - The exact package name

3. **package_dependencies** - List dependencies for a package
   - Parameter: `package` - The exact package name

4. **compare_versions** - Compare versions of packages
   - Parameter: `package` - The package name to compare versions for

5. **package_graph** - Query the package dependency graph using provides and requires relationships
   - Parameter: `package` - The package name to start the graph query from
   - Parameter: `query_type` - The type of query to perform: 
     - `requires` - Show what a package requires directly
     - `provides` - Show what capabilities a package provides
     - `depends_on` - Show a recursive dependency graph
     - `required_by` - Show what packages depend on this package
     - `what_provides` - Show what packages provide a certain capability
   - Parameter: `depth` (optional) - Maximum depth for recursive queries (default: 1, max: 5)

## Package Database

The server uses an APKINDEX.tar.gz file which contains the package database information. 
This file follows the Alpine Linux repository format and is parsed using the chainguard-dev/apko 
Go module.

By default, the server automatically downloads the latest APKINDEX.tar.gz from the Wolfi repository:
- https://packages.wolfi.dev/os/aarch64/APKINDEX.tar.gz (on ARM64 systems)
- https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz (on x86_64 systems)

The downloaded file is cached in a standard OS-specific location to avoid unnecessary downloads on restart:
- Linux: `$XDG_CACHE_HOME/wolfi-mcp/APKINDEX.tar.gz` (defaults to `~/.cache/wolfi-mcp/APKINDEX.tar.gz`)
- macOS: `~/Library/Caches/wolfi-mcp/APKINDEX.tar.gz`
- Windows: `%LOCALAPPDATA%\wolfi-mcp\cache\APKINDEX.tar.gz`

You can override this behavior and use a specific APKINDEX file by using the `-index` flag.

## Development

```bash
# Verify modules
go mod verify

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.