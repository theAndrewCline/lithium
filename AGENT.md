# Lithium - Go Project Agent Guide

## Build/Test Commands
- `go build` - Build the application
- `go run main.go` - Run the application directly
- `go test ./...` - Run all tests recursively
- `go test -run TestName` - Run a specific test by name
- `go test -v ./...` - Run tests with verbose output
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Run Go static analysis
- `go mod tidy` - Clean up module dependencies

## Architecture
- Simple Go module named "lithium"
- Single binary application with main.go entry point
- Binary output excluded via .gitignore (bin/ directory)
- No external dependencies currently

## Code Style Guidelines
- Follow standard Go conventions (gofmt, go vet)
- Use Go modules for dependency management
- Standard Go naming: PascalCase for exported, camelCase for unexported
- Error handling: explicit error returns, no panics in library code
- Imports: standard library first, then third-party, then local packages
- Use meaningful variable names, avoid abbreviations
- Functions should be focused and single-purpose
