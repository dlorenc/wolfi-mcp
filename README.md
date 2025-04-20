# Wolfi MCP Package Database

An MCP (Model Context Protocol) server for querying and interacting with Alpine-based Linux distribution package databases.

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

# Use an APKINDEX from a URL (will be downloaded and cached)
./mcp-server -index https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz

# Use multiple APKINDEX files (will be merged following Alpine merging semantics)
./mcp-server -index /path/to/first/APKINDEX.tar.gz -index /path/to/second/APKINDEX.tar.gz

# Use a mix of local files and URLs
./mcp-server -index /path/to/local/APKINDEX.tar.gz -index https://example.com/repo/APKINDEX.tar.gz
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

### Multiple APKINDEX Support

The server supports loading multiple APKINDEX files by using the `-index` flag multiple times:

```bash
./mcp-server -index first.tar.gz -index second.tar.gz -index third.tar.gz
```

The `-index` flag can accept:
- Local file paths: `-index /path/to/APKINDEX.tar.gz`
- URLs: `-index https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz`
- A mix of both: `-index local.tar.gz -index https://example.com/APKINDEX.tar.gz`

When a URL is provided, the server will:
1. Download the APKINDEX file from the specified URL
2. Cache it locally in the standard cache directory (using a filename based on the URL hash)
3. Use the cached file for subsequent runs (unless the cache is cleared)

When multiple APKINDEX files are provided, they are merged following Alpine Linux package repository semantics:

1. If a package appears in multiple index files:
   - The highest version wins
   - For identical versions, the most recently indexed one (rightmost in command line arguments) takes precedence

This allows combining packages from different repositories or overlaying custom packages on top of the base distribution.

## Using with Claude Code

This MCP server is designed to work with Claude Code via the Model Context Protocol (MCP). Here's how to set it up:

### Prerequisites

- Go 1.24 or higher
- Claude Code CLI
- Git (to clone the repository)

### Installation and Setup

1. **Clone and build the server**:
   ```bash
   git clone https://github.com/dlorenc/wolfi-mcp.git
   cd wolfi-mcp
   go build -o mcp-server
   ```

2. **Register the MCP server with Claude Code**:
   
   Use the `claude mcp add` command to register the MCP server:
   
   ```bash
   claude mcp add wolfi -- ./mcp-server
   ```
   
   You can also specify a custom APKINDEX file:
   
   ```bash
   claude mcp add wolfi -- ./mcp-server -index /path/to/your/APKINDEX.tar.gz
   ```
   
   With a URL:
   
   ```bash
   claude mcp add wolfi -- ./mcp-server -index https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz
   ```
   
   Or with multiple APKINDEX files (mix of local files and URLs):
   
   ```bash
   claude mcp add wolfi -- ./mcp-server -index /path/to/first/APKINDEX.tar.gz -index https://example.com/custom/APKINDEX.tar.gz
   ```
   
   To make the server available in all projects:
   
   ```bash
   claude mcp add wolfi -s user -- ./mcp-server
   ```

3. **Verify the server is registered**:
   
   ```bash
   claude mcp list
   ```

### Using the Wolfi MCP Server Tools

Once the server is registered, you can access its tools in Claude Code conversations with this format:

```
/tool mcp__wolfi__tool_name --parameter=value
```

For example:

```
/tool mcp__wolfi__search_packages --query=python
/tool mcp__wolfi__package_info --package=python3
/tool mcp__wolfi__package_dependencies --package=python3
/tool mcp__wolfi__compare_versions --package=python3
/tool mcp__wolfi__package_graph --package=python3 --query_type=depends_on
```

### Example Session

Here's an example of how Claude Code might use these tools in a conversation:

```
User: What packages are available for Python in Wolfi?

Claude: I'll search the Wolfi package database for Python packages.

/tool mcp__wolfi__search_packages --query=python

Based on the search results, I found several Python-related packages in the Wolfi repository:
- python3 (3.11.6)
- python3-dev (3.11.6)
- python3-doc (3.11.6)
- ...

User: What are the dependencies of python3?

Claude: Let me check the dependencies for the python3 package.

/tool mcp__wolfi__package_dependencies --package=python3

The python3 package has the following dependencies:
1. ca-certificates
2. libcrypto3
3. ...
```

### Managing the MCP Server

- **List registered servers**: `claude mcp list`
- **Get server details**: `claude mcp get wolfi`
- **Remove the server**: `claude mcp remove wolfi`

## Troubleshooting

### Common Issues

- **"Package not found" errors**: This could indicate that the package name is misspelled or the package is not available in the Wolfi repository.
- **Connection errors when downloading**: Check your internet connection and firewall settings.
- **Permission errors**: Ensure you have write permissions to the cache directory.

### Debugging

To see more information about what the server is doing:

1. Run the MCP server directly:
   ```bash
   ./mcp-server
   ```
   This will show download progress and package loading information.

2. Check the cache directories if you're experiencing issues:
   - Linux: `~/.cache/wolfi-mcp/`
   - macOS: `~/Library/Caches/wolfi-mcp/`
   - Windows: `%LOCALAPPDATA%\wolfi-mcp\cache\`

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