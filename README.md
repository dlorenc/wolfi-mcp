# Wolfi MCP Package Database

An MCP (Model Calling Protocol) server for querying and interacting with Alpine-based Linux distribution package databases.

## Features

- Search for packages by name (partial matching supported)
- Get detailed information about specific packages
- List dependencies for packages
- Compare versions of packages

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
./mcp-server
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

## Package Database

The server uses an APKINDEX.tar.gz file which contains the package database information. 
This file follows the Alpine Linux repository format and is parsed using the chainguard-dev/apko 
Go module.

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