package core

import (
	gsdk "code.gitea.io/sdk/gitea"
)

// GiteaClient defines the interface for Gitea operations
type GiteaClient interface {
	// Organization secrets operations
	CreateOrgActionSecret(org string, opt gsdk.CreateSecretOption) (interface{}, error)

	// Repository secrets operations
	CreateRepoActionSecret(owner, repo string, opt gsdk.CreateSecretOption) (interface{}, error)

	// Health check
	Ping() error

	// Close and cleanup
	Close() error
}
