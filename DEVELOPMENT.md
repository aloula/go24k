# Go24K Development Setup

## Working with Integration Tests

This project includes integration tests that require special build tags to be included during compilation.

### VS Code Setup

The project includes VS Code configuration in `.vscode/settings.json` that automatically includes the `integration` build tag. This allows gopls (Go Language Server) to properly parse and understand all test files.

### Manual Testing

To run integration tests manually:

```bash
# Run integration tests
go test -tags=integration

# Run all tests including integration
go test -tags=integration ./...

# Build with integration tests included
go build -tags=integration
```

### Build Tags Explained

- `integration_test.go` uses `//go:build integration` tag
- These tests require FFmpeg and the compiled binary
- They test end-to-end functionality with real files
- Regular unit tests run without special tags

### Troubleshooting

If you see "No packages found" errors in VS Code:

1. Reload VS Code window (Ctrl+Shift+P → "Developer: Reload Window")
2. Check that `.vscode/settings.json` exists and contains build flags
3. Manually set build tags in VS Code settings if needed

### Editor Configuration

For other editors, add these flags to your Go tooling configuration:
- Build flags: `-tags=integration`
- Environment: `GOFLAGS=-tags=integration`

## Code Quality and Linting

### golangci-lint Setup

The project uses `golangci-lint` for comprehensive code analysis. Install it using:

```bash
# Automatic installation (recommended)
./install-linter.sh

# Manual installation
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Running Linters

```bash
# Modern linter (recommended)
make lint-modern
# or directly:
golangci-lint run

# Traditional linter (via test.sh)
make lint

# Quick analysis
golangci-lint run --fast

# Auto-fix issues
golangci-lint run --fix
```

### Linter Configuration

The project includes `.golangci.yml` with optimized settings:
- Focuses on important issues (errcheck, gosimple, etc.)
- Relaxed rules for test files
- Appropriate complexity limits for video processing code
- Security checks for production code

### Pre-commit Checks

Before committing code:

```bash
# Full check pipeline
make check

# Or step by step:
make fmt          # Format code
make lint-modern  # Static analysis
make test-unit    # Unit tests
```