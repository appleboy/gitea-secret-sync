package core

import (
	gsdk "code.gitea.io/sdk/gitea"
)

// Retrier defines the interface for retry mechanisms
type Retrier interface {
	// DoWithRetry executes an operation that returns Response with retry logic
	DoWithRetry(operation func() (*gsdk.Response, error)) (*gsdk.Response, error)
	// IsRetryable determines if an error/response combination is retryable
	IsRetryable(resp *gsdk.Response, err error) bool
}

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
