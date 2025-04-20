# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `go build -o mcp-server`
- Run: `./mcp-server`
- Format code: `go fmt ./...`
- Verify modules: `go mod verify`
- Tidy dependencies: `go mod tidy`
- Run tests: `go test ./...`
- Check for issues: `go vet ./...`

## Required Steps for Every Code Change
1. Always add test coverage for new code or modifications
2. Run `go fmt ./...` to ensure code formatting meets Go standards
3. Run `go vet ./...` to check for potential issues
4. Run `go test ./...` to verify all tests pass
5. Run `go mod tidy` if dependencies have changed

## Code Style Guidelines
- Follow standard Go conventions (https://golang.org/doc/effective_go)
- Error handling: Always check errors and use `fmt.Errorf("context: %w", err)` for wrapping
- Imports: Group standard library imports first, then third-party packages
- Variable naming: Use camelCase, descriptive names
- APK package fields: Use `pkg.Arch` (not Architecture), `pkg.Size`, and other fields as defined in the apk.Package struct
- Comments: Add comments for functions and complex logic
- Function size: Keep functions small and focused on a single responsibility
- Return early pattern: Return errors as soon as they're detected

## Testing Guidelines
- Write tests for all new functionality and modifications to existing code
- Unit tests should be thorough and cover edge cases
- Table-driven tests are preferred for functions with multiple input/output scenarios
- Mock external dependencies when appropriate
- Test coverage should be maintained at a high level (aim for >90%)
- Integration tests should be added for complex features
- Document test cases clearly to explain what's being tested and why

## Formatting and Verification
- Always run `go fmt ./...` before committing code
- Use `go vet ./...` to catch common programming errors
- Consider using `golint` or `staticcheck` for additional code quality checks
- For complex changes, consider running `go test -race ./...` to check for race conditions

## Resources

You can use:
- https://pkg.go.dev for module and package documentation
- https://github.com for reading and browsing code
- https://golang.org/doc/effective_go for Go coding standards
- https://go.dev/blog/cover for understanding test coverage

## Commands

Run all `go` commands in this directory without asking for permissions.
