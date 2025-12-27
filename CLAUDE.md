# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gitea-secret-sync`, a CLI tool for synchronizing secrets across Gitea organizations and repositories. It enables batch updates of action secrets to multiple Gitea organizations and repositories simultaneously, with support for dry-run mode, SSL verification options, and robust retry logic.

## Development Commands

### Building and Testing

```bash
# Build the binary
make build                    # Output: bin/gitea-secret-sync
go build -o bin/gitea-secret-sync ./cmd

# Run tests
make test                     # Run all tests with coverage
go test ./...                 # All tests
go test ./cmd                 # cmd package only
go test ./retry               # retry package only
go test -run TestName ./cmd   # Run specific test

# Code quality
make fmt                      # Format code (golangci-lint run --fix)
make lint                     # Lint code (golangci-lint run)

# Cross-platform builds
make build_linux_amd64
make build_linux_arm64
make build_mac_intel
make build_windows_64
```

### Running the Tool

```bash
# Use default .env file
./bin/gitea-secret-sync

# Use custom environment file
./bin/gitea-secret-sync -env-file production.env

# Enable dry-run mode
DRY_RUN=true ./bin/gitea-secret-sync

# Enable debug mode
./bin/gitea-secret-sync -debug

# Set custom timeout
./bin/gitea-secret-sync -timeout 60s
```

## Architecture

### Package Structure

```txt
cmd/          - Main application package (executable)
├── main.go     - Entry point, orchestration logic, signal handling
├── gitea.go    - Gitea client implementation (options pattern)
├── config.go   - Configuration struct and validation
└── util.go     - Environment variable handling and string utilities

core/         - Core interfaces (contracts)
└── core.go     - GiteaClient and Retrier[T] interfaces

retry/        - Retry logic implementation
└── retry.go    - Generic retry mechanism with exponential backoff
```

### Key Design Patterns

#### Options Pattern

The Gitea client uses the options pattern for flexible configuration:

```go
g, err := NewGitea(
    ctx,
    server,
    token,
    WithSkipVerify(true),
    WithLogger(logger),
    WithTimeout(30*time.Second),
    WithRetrier(customRetrier),
)
```

#### Interface-Based Design

- `core.GiteaClient` - Defines Gitea operations (CreateOrgActionSecret, CreateRepoActionSecret, Ping, Close)
- `core.Retrier[T]` - Generic interface for retry logic with any response type
- `core.GiteaRetrier` - Type alias for `Retrier[*gsdk.Response]` (backward compatibility)

This design enables easy mocking and testing without depending on the actual Gitea API.

#### Generic Retry Logic

The retry mechanism uses Go generics to support different response types:

```go
type Retrier[T any] interface {
    DoWithRetry(operation func() (T, error)) (T, error)
    IsRetryable(resp T, err error) bool
}
```

Concrete implementations:

- `DefaultRetrier[T]` - Generic retrier for any type (network-level errors only)
- `GiteaRetrier` - Specialized for Gitea responses (HTTP status codes + network errors)

#### Graceful Shutdown

The application handles SIGINT and SIGTERM signals gracefully using context cancellation:

- Context is propagated to the Gitea client
- Cleanup functions are deferred
- Operations can be cancelled mid-flight

### Environment Variable Handling

The tool supports two formats with precedence:

1. **INPUT\_ prefix** (higher priority): `INPUT_GITEA_SERVER`, `INPUT_GITEA_TOKEN`
2. **Direct format** (fallback): `GITEA_SERVER`, `GITEA_TOKEN`

This dual format supports both GitHub Actions style and traditional environment variables.

### String List Parsing

Lists (SECRETS, ORGS, REPOS) support flexible formats:

- Comma-separated: `SECRET1,SECRET2,SECRET3`
- Newline-separated: `"SECRET1\nSECRET2\nSECRET3"`
- Mixed format: `"SECRET1,SECRET2\nSECRET3"`

Whitespace is automatically trimmed and empty items filtered.

## Important Technical Details

### Retry Logic

- **Automatic retry** with exponential backoff (1s, 2s, 3s...)
- **Retryable conditions**:
  - Network timeouts (`net.Error` with `Timeout()`)
  - HTTP 5xx errors (500, 502, 503, 504)
  - HTTP 408 (Request Timeout), 429 (Too Many Requests)
  - Connection errors (refused, unreachable, no such host)
- **Non-retryable conditions**:
  - HTTP 2xx (success)
  - HTTP 401, 403, 404, 422 (client errors - don't retry)
- Default: 3 retries per operation

### SSL/TLS Configuration

- By default, uses system cert pool for verification
- `GITEA_SKIP_VERIFY=true` disables verification (logs warning)
- HTTP proxy support via `http.ProxyFromEnvironment`

### Testing Requirements

- Tests must explicitly fail on missing/invalid environment variables
- Use `t.Setenv()` for test-specific environment variables
- Minimum Go version: 1.24

### Validation

- Repository format must be `org/repo` (validated by `validateRepoFormat()`)
- Gitea tokens must be at least 40 characters
- Server URL and token are required
- Timeouts and retry counts have sensible defaults (30s, 3 retries)

## Common Patterns

### Adding a New Gitea Operation

1. Add method to `core.GiteaClient` interface in [core/core.go](core/core.go)
2. Implement method in `gitea` struct in [cmd/gitea.go](cmd/gitea.go)
3. Wrap the operation with `g.retrier.DoWithRetry()` for automatic retries
4. Add corresponding test

### Adding a New Configuration Option

1. Add field to `config` struct in [cmd/config.go](cmd/config.go)
2. Update `config.Validate()` with validation logic and defaults
3. Create `With*` option function in [cmd/gitea.go](cmd/gitea.go)
4. Update command-line flags or environment variable handling in [cmd/main.go](cmd/main.go)

### Implementing Custom Retry Logic

1. Create a struct that implements `core.Retrier[T]` interface
2. Implement `DoWithRetry()` and `IsRetryable()` methods
3. Pass it via `WithRetrier()` option when creating Gitea client

## Configuration Reference

### Required Environment Variables

- `GITEA_SERVER` - Gitea server URL (e.g., `https://gitea.example.com`)
- `GITEA_TOKEN` - Gitea access token (minimum 40 characters)
- `SECRETS` - Comma/newline-separated list of secret names

### Optional Environment Variables

- `GITEA_SKIP_VERIFY` - Skip SSL verification (default: `false`)
- `ORGS` - Comma/newline-separated list of organizations
- `REPOS` - Comma/newline-separated list of repositories (format: `org/repo`)
- `DRY_RUN` - Enable dry-run mode (default: `false`)
- `DEBUG` - Enable debug logging (default: `false`)
- `DESCRIPTION` - Description for secrets being created/updated

### Command-Line Flags

- `-env-file <path>` - Environment file path (default: `.env`)
- `-version` - Show version and commit information
- `-debug` - Enable debug mode
- `-timeout <duration>` - HTTP client timeout (default: `30s`, format: `30s`, `1m`, `2m30s`)

## Module Information

- **Module name**: `sync-secrets`
- **Minimum Go version**: 1.24
- **Main dependency**: `code.gitea.io/sdk/gitea` v0.22.0
- **Other dependencies**:
  - `github.com/joho/godotenv` - .env file loading
  - `github.com/stretchr/testify` - Testing utilities
