# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `go build -o mcp-server`
- Run: `./mcp-server`
- Format code: `go fmt ./...`
- Verify modules: `go mod verify`
- Tidy dependencies: `go mod tidy`

## Code Style Guidelines
- Follow standard Go conventions (https://golang.org/doc/effective_go)
- Error handling: Always check errors and use `fmt.Errorf("context: %w", err)` for wrapping
- Imports: Group standard library imports first, then third-party packages
- Variable naming: Use camelCase, descriptive names
- APK package fields: Use `pkg.Arch` (not Architecture), `pkg.Size`, and other fields as defined in the apk.Package struct
- Comments: Add comments for functions and complex logic
- Function size: Keep functions small and focused on a single responsibility
- Return early pattern: Return errors as soon as they're detected

## Resources

You can use:
- https://pkg.go.dev for module and package documentation
- https://github.com for reading and browsing code

## Commands

Run all `go` commands in this directory without asking for permissions.
