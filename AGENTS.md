# AGENTS.md — AI Agent Instructions for migtools/udistribution

## Project Overview
`udistribution` is a Go library that provides a client to interface with storage drivers of [distribution/distribution](https://github.com/distribution/distribution) (the Docker registry) without requiring a listening HTTP server. This enables direct access to container image registries' storage backends, which is used in the OADP ecosystem for image backup and restore operations.

- **Primary Language**: Go
- **Module**: `github.com/migtools/udistribution`
- **Default Branch**: `main`

## Build Instructions
```bash
# Build the library
make build
# or directly:
go build ./...
```

## Test Instructions
```bash
# Run all tests
make test
# or directly:
go test ./...

# Run specific tests
go test ./pkg/... -run TestName

# Format code
make fmt
# or directly:
go fmt ./...

# Vet code
go vet ./...
```

## Code Conventions
- Library-only project (no main binary)
- All public APIs are in `pkg/` directory
- Uses distribution/distribution storage driver interfaces
- Standard Go error handling patterns
- Follow Go conventions for exported vs unexported identifiers

## Project Structure
```
pkg/           - All library code
  image/       - Container image operations
  client/      - Registry client implementation
  storage/     - Storage driver integration
go.mod         - Module definition
go.sum         - Dependency checksums
```

## CI/CD
- GitHub Actions workflows in `.github/workflows/`:
  - `go.yml` — Go build and test
  - `go-pr.yml` — PR validation
  - `codeql-analysis.yml` — CodeQL security analysis
- Reproduce CI locally:
  ```bash
  make fmt
  make test
  make build
  ```

## Common Tasks

### Adding support for a new storage driver
1. Implement the storage driver adapter in `pkg/`
2. Register the driver in the client initialization
3. Add tests for the new driver
4. Update documentation

### Working with container image layers
- Image operations are in `pkg/image/`
- Follow the distribution/distribution interfaces for compatibility
