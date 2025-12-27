# Gitea Secret Sync

[English](README.md) | [繁體中文](README.zh-tw.md) | [简体中文](README.zh-cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/gitea-secret-sync)](https://goreportcard.com/report/github.com/appleboy/gitea-secret-sync)
[![Lint and Testing](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml)

A powerful CLI tool for synchronizing secrets across Gitea organizations and repositories. This tool enables batch updates of action secrets to multiple Gitea organizations and repositories simultaneously, with support for dry-run mode and SSL verification options.

## Features

- 🔐 **Batch Secret Management**: Sync multiple secrets to organizations and repositories at once
- 🏢 **Organization Support**: Update secrets for multiple Gitea organizations
- 📁 **Repository Support**: Update secrets for specific repositories (format: `org/repo`)
- 🔍 **Dry Run Mode**: Preview changes without actually updating secrets
- 🔒 **SSL Configuration**: Option to skip SSL verification for self-hosted Gitea instances
- 📁 **Environment File Support**: Load configuration from `.env` files
- 📊 **Structured Logging**: Clear logging with structured output
- 🛡️ **Graceful Shutdown**: Handle interruption signals gracefully

## Installation

### From Source

```bash
go install github.com/appleboy/gitea-secret-sync/cmd@latest
```

### From Release

Download the latest binary from the [releases page](https://github.com/appleboy/gitea-secret-sync/releases).

## Usage

### Command Line Options

```bash
gitea-secret-sync [options]
```

**Options:**

- `-env-file string`: Read in a file of environment variables (default: ".env")
- `-version`: Show version information and exit

### Environment Variables

The tool supports two formats for environment variables:

1. Direct format: `VARIABLE_NAME=value`
2. Input format: `INPUT_VARIABLE_NAME=value` (takes precedence over direct format)

#### Required Configuration

| Variable       | Description                                             | Example                     |
| -------------- | ------------------------------------------------------- | --------------------------- |
| `GITEA_SERVER` | Gitea server URL                                        | `https://gitea.example.com` |
| `GITEA_TOKEN`  | Gitea access token                                      | `your_gitea_token_here`     |
| `SECRETS`      | Comma or newline-separated list of secret names to sync | `SECRET1,SECRET2,SECRET3`   |

#### Optional Configuration

| Variable            | Description                                                | Default | Example                 |
| ------------------- | ---------------------------------------------------------- | ------- | ----------------------- |
| `GITEA_SKIP_VERIFY` | Skip SSL certificate verification                          | `false` | `true`                  |
| `ORGS`              | Comma or newline-separated list of organizations to update | -       | `org1,org2,org3`        |
| `REPOS`             | Comma or newline-separated list of repositories to update  | -       | `org1/repo1,org2/repo2` |
| `DRY_RUN`           | Enable dry-run mode (preview only)                         | `false` | `true`                  |

#### Secret Values

For each secret name listed in `SECRETS`, you need to provide the corresponding environment variable:

```bash
# If SECRETS=SECRET1,SECRET2,SECRET3
SECRET1=actual_secret_value_1
SECRET2=actual_secret_value_2
SECRET3=actual_secret_value_3
```

### Input Format Support

The tool supports flexible input formats for `SECRETS`, `ORGS`, and `REPOS` variables:

#### Comma-separated (traditional format)

```bash
SECRETS=SECRET1,SECRET2,SECRET3
ORGS=org1,org2,org3
REPOS=org1/repo1,org2/repo2,org3/repo3
```

#### Newline-separated format

```bash
SECRETS="SECRET1
SECRET2
SECRET3"

ORGS="org1
org2
org3"

REPOS="org1/repo1
org2/repo2
org3/repo3"
```

#### Mixed format (comma and newline)

```bash
SECRETS="SECRET1,SECRET2
SECRET3"
ORGS="org1,org2
org3"
```

**Note:** Extra whitespace and empty items are automatically handled and filtered out.

## Configuration Examples

### Using .env File

Create a `.env` file in your project directory:

```bash
# Gitea Configuration
GITEA_SERVER=https://gitea.example.com
GITEA_TOKEN=your_gitea_access_token
GITEA_SKIP_VERIFY=false

# Secrets to sync (comma-separated format)
SECRETS=DATABASE_URL,API_KEY,JWT_SECRET

# Alternative: newline-separated format
# SECRETS="DATABASE_URL
# API_KEY
# JWT_SECRET"

# Secret values
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk-1234567890abcdef
JWT_SECRET=super-secret-jwt

# Target organizations and repositories (comma-separated)
ORGS=my-org,another-org
REPOS=my-org/repo1,my-org/repo2,another-org/repo3

# Alternative: newline-separated format
# ORGS="my-org
# another-org"
# REPOS="my-org/repo1
# my-org/repo2
# another-org/repo3"

# Optional: Enable dry-run mode
DRY_RUN=false
```

### Using Environment Variables

```bash
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_gitea_access_token
export SECRETS=DATABASE_URL,API_KEY,JWT_SECRET
export DATABASE_URL=postgresql://user:pass@localhost/db
export API_KEY=sk-1234567890abcdef
export JWT_SECRET=super-secret-jwt
export ORGS=my-org,another-org
export REPOS=my-org/repo1,my-org/repo2
```

### Using INPUT\_ Prefix (GitHub Actions Style)

```bash
export INPUT_GITEA_SERVER=https://gitea.example.com
export INPUT_GITEA_TOKEN=your_gitea_access_token
export INPUT_SECRETS=DATABASE_URL,API_KEY,JWT_SECRET
export INPUT_DATABASE_URL=postgresql://user:pass@localhost/db
export INPUT_API_KEY=sk-1234567890abcdef
export INPUT_JWT_SECRET=super-secret-jwt
export INPUT_ORGS=my-org,another-org
export INPUT_REPOS=my-org/repo1,my-org/repo2
```

## Usage Examples

### Basic Usage

```bash
# Load configuration from .env file and sync secrets
./gitea-secret-sync

# Use custom environment file
./gitea-secret-sync -env-file production.env
```

### Dry Run Mode

```bash
# Preview what would be changed without actually updating
DRY_RUN=true ./gitea-secret-sync
```

### Organization Only

```bash
# Update secrets only for specific organizations
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export ORGS=org1,org2

./gitea-secret-sync
```

### Repository Only

```bash
# Update secrets only for specific repositories
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export REPOS=org1/repo1,org1/repo2,org2/repo3

./gitea-secret-sync
```

### Using Newline-Separated Format

```bash
# Example using newline-separated format for better readability
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS="DATABASE_URL
API_KEY
JWT_SECRET
WEBHOOK_SECRET"
export DATABASE_URL=postgresql://user:pass@localhost/db
export API_KEY=sk-1234567890abcdef
export JWT_SECRET=super-secret-jwt
export WEBHOOK_SECRET=whsec_1234567890
export ORGS="production-org
staging-org
development-org"

./gitea-secret-sync
```

## Getting Your Gitea Token

1. Log in to your Gitea instance
2. Go to **Settings** → **Applications** → **Generate New Token**
3. Select the required scopes:
   - `write:organization` (for organization secrets)
   - `write:repository` (for repository secrets)
4. Copy the generated token and use it as `GITEA_TOKEN`

## Error Handling

The tool includes comprehensive error handling:

- **Invalid Configuration**: Missing required variables will cause the program to exit with an error message
- **Network Issues**: Connection problems are logged with detailed error information
- **API Errors**: Gitea API errors are captured and logged with context
- **Invalid Repository Format**: Repository names must be in `org/repo` format
- **Graceful Shutdown**: The tool handles SIGINT and SIGTERM signals gracefully

## Logging

The tool uses structured logging with the following log levels:

- `INFO`: Normal operation information
- `WARN`: Warnings (e.g., dry-run mode notifications)
- `ERROR`: Error conditions that prevent operation

Log format includes timestamp, level, message, and relevant context fields.

## Development

### Building from Source

```bash
git clone https://github.com/appleboy/gitea-secret-sync.git
cd gitea-secret-sync
go build -o gitea-secret-sync ./cmd
```

### Running Tests

```bash
go test ./cmd
```

### Project Structure

```txt
cmd/
├── gitea.go           # Gitea client wrapper with SSL configuration
├── main.go            # Main application entry point and logic
├── util.go            # Utility functions for environment variables and string parsing
├── util_test.go       # Unit tests for utility functions
└── util_split_test.go # Unit tests for string splitting functionality
```

### Key Components

- **`gitea.go`**: Implements the Gitea client wrapper with SSL configuration support
- **`main.go`**: Contains the main application logic, signal handling, and batch processing
- **`util.go`**: Provides utility functions for reading environment variables with INPUT\_ prefix support and flexible string parsing (comma/newline separation)
- **`util_test.go`**: Unit tests for utility functions
- **`util_split_test.go`**: Unit tests for the new string splitting functionality

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Security Considerations

- **Token Security**: Store your Gitea tokens securely and never commit them to version control
- **SSL Verification**: Only disable SSL verification (`GITEA_SKIP_VERIFY=true`) for trusted internal networks
- **Dry Run**: Always test with `DRY_RUN=true` before making actual changes
- **Secret Values**: Ensure secret values are properly escaped and don't contain special characters that might cause parsing issues

## Troubleshooting

### Common Issues

#### "missing gitea server or token"

- Ensure `GITEA_SERVER` and `GITEA_TOKEN` environment variables are set

#### "invalid repo format"

- Repository names must be in `org/repo` format (e.g., `myorg/myrepo`)

#### "failed to update secrets"

- Check if your Gitea token has sufficient permissions
- Verify that the organization or repository exists
- Ensure the Gitea server is accessible

#### SSL certificate errors

- Set `GITEA_SKIP_VERIFY=true` for self-signed certificates (not recommended for production)
- Or properly configure SSL certificates

### Getting Help

If you encounter issues:

1. Enable dry-run mode to check configuration: `DRY_RUN=true`
2. Check the logs for detailed error messages
3. Verify your Gitea token permissions
4. Test connectivity to your Gitea server

## Related Projects

- [Gitea](https://gitea.io/) - A painless self-hosted Git service
- [Gitea SDK](https://code.gitea.io/sdk) - Go SDK for Gitea API

---

Made with ❤️ for the Gitea community
