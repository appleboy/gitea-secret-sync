package retry

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"sync-secrets/core"

	gsdk "code.gitea.io/sdk/gitea"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries int
	Logger     *slog.Logger
}

// DefaultRetrier implements the core.Retrier interface with default retry logic using generics
type DefaultRetrier[T any] struct {
	config *RetryConfig
}

// NewDefaultRetrier creates a new DefaultRetrier with the given configuration
func NewDefaultRetrier[T any](config *RetryConfig) *DefaultRetrier[T] {
	if config == nil {
		config = &RetryConfig{
			MaxRetries: 3,
			Logger:     slog.Default(),
		}
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	return &DefaultRetrier[T]{
		config: config,
	}
}

// DoWithRetry executes an operation that returns T with retry logic
func (r *DefaultRetrier[T]) DoWithRetry(operation func() (T, error)) (T, error) {
	var lastResp T

	for i := 0; i < r.config.MaxRetries; i++ {
		resp, err := operation()
		lastResp = resp

		// Check if error/response is retryable
		if !r.IsRetryable(resp, err) {
			return resp, err
		}

		if i < r.config.MaxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			r.config.Logger.Warn("operation failed, retrying",
				"attempt", i+1,
				"backoff", backoff,
				"error", err)
			time.Sleep(backoff)
		}
	}
	return lastResp, fmt.Errorf("operation failed after %d retries", r.config.MaxRetries)
}

// IsRetryable determines if an error/response is retryable (default implementation)
func (r *DefaultRetrier[T]) IsRetryable(resp T, err error) bool {
	// Default implementation only checks for network-level errors
	if err != nil {
		// Check for network timeout errors
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}

		// Check for HTTP status codes that indicate temporary issues
		errMsg := err.Error()
		if strings.Contains(errMsg, "500") ||
			strings.Contains(errMsg, "502") ||
			strings.Contains(errMsg, "503") ||
			strings.Contains(errMsg, "504") {
			return true
		}

		// Check for connection refused or network unreachable
		if strings.Contains(errMsg, "connection refused") ||
			strings.Contains(errMsg, "network is unreachable") ||
			strings.Contains(errMsg, "no such host") {
			return true
		}
	}

	return false
}

// GiteaRetrier is a concrete implementation for Gitea responses
type GiteaRetrier struct {
	config *RetryConfig
}

// Ensure GiteaRetrier implements core.GiteaRetrier interface
var _ core.GiteaRetrier = (*GiteaRetrier)(nil)

// IsRetryable determines if an error/response is retryable for Gitea responses
func (g *GiteaRetrier) IsRetryable(resp *gsdk.Response, err error) bool {
	// Check if operation succeeded (2xx status code and no error)
	if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false // Success - no need to retry
	}

	// If there's an error, check if it's retryable
	if err != nil {
		// Check for network timeout errors
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}

		// Check for HTTP status codes that indicate temporary issues
		errMsg := err.Error()
		if strings.Contains(errMsg, "500") ||
			strings.Contains(errMsg, "502") ||
			strings.Contains(errMsg, "503") ||
			strings.Contains(errMsg, "504") {
			return true
		}

		// Check for connection refused or network unreachable
		if strings.Contains(errMsg, "connection refused") ||
			strings.Contains(errMsg, "network is unreachable") ||
			strings.Contains(errMsg, "no such host") {
			return true
		}
	}

	// If no error but response indicates server issues, retry
	if resp != nil {
		switch resp.StatusCode {
		case 500, 502, 503, 504: // Server errors
			return true
		case 408, 429: // Request timeout, too many requests
			return true
		case 401, 403, 404, 422: // Client errors - don't retry
			return false
		}

		// Non-2xx responses should generally be retried except for client errors
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Retry on server errors (5xx) and some 4xx (but not client auth/not found)
			return resp.StatusCode >= 500 || resp.StatusCode == 408 || resp.StatusCode == 429
		}
	}

	return false
}

// DoWithRetry executes an operation that returns *gsdk.Response with retry logic
func (g *GiteaRetrier) DoWithRetry(operation func() (*gsdk.Response, error)) (*gsdk.Response, error) {
	var lastResp *gsdk.Response

	for i := 0; i < g.config.MaxRetries; i++ {
		resp, err := operation()
		lastResp = resp

		// Check if error/response is retryable
		if !g.IsRetryable(resp, err) {
			return resp, err
		}

		if i < g.config.MaxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			g.config.Logger.Warn("operation failed, retrying",
				"attempt", i+1,
				"backoff", backoff,
				"error", err)
			time.Sleep(backoff)
		}
	}
	return lastResp, fmt.Errorf("operation failed after %d retries", g.config.MaxRetries)
}

// NewGiteaRetrier creates a new GiteaRetrier with the given configuration
func NewGiteaRetrier(config *RetryConfig) *GiteaRetrier {
	if config == nil {
		config = &RetryConfig{
			MaxRetries: 3,
			Logger:     slog.Default(),
		}
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	return &GiteaRetrier{
		config: config,
	}
}
