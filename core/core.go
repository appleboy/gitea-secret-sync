package core

import (
	gsdk "code.gitea.io/sdk/gitea"
)

// Retrier defines the interface for retry mechanisms using generics
type Retrier[T any] interface {
	// DoWithRetry executes an operation that returns T with retry logic
	DoWithRetry(operation func() (T, error)) (T, error)
	// IsRetryable determines if an error/response combination is retryable
	IsRetryable(resp T, err error) bool
}

// GiteaRetrier is a type alias for Retrier with gsdk.Response for backward compatibility
type GiteaRetrier = Retrier[*gsdk.Response]

// GiteaClient defines the interface for Gitea operations
type GiteaClient interface {
	// Organization secrets operations
	CreateOrgActionSecret(org string, opt gsdk.CreateSecretOption) (*gsdk.Response, error)

	// Repository secrets operations
	CreateRepoActionSecret(owner, repo string, opt gsdk.CreateSecretOption) (*gsdk.Response, error)

	// Health check
	Ping() error

	// Close and cleanup
	Close() error
}
