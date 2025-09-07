package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"sync-secrets/core"
	"sync-secrets/retry"

	gsdk "code.gitea.io/sdk/gitea"
)

// gitea is a struct that holds the gitea client and implements core.GiteaClient interface.
type gitea struct {
	ctx     context.Context
	config  *config
	client  *gsdk.Client
	logger  *slog.Logger
	retrier core.Retrier
}

// Ensure gitea implements core.GiteaClient interface
var _ core.GiteaClient = (*gitea)(nil)

// GiteaOption is a function type for configuring gitea client options.
type GiteaOption func(*gitea)

// WithConfig sets the configuration for the gitea client.
func WithConfig(cfg *config) GiteaOption {
	return func(g *gitea) {
		g.config = cfg
	}
}

// WithLogger sets the logger for the gitea client.
func WithLogger(logger *slog.Logger) GiteaOption {
	return func(g *gitea) {
		g.logger = logger
	}
}

// WithSkipVerify sets whether to skip TLS certificate verification.
func WithSkipVerify(skip bool) GiteaOption {
	return func(g *gitea) {
		if g.config == nil {
			g.config = &config{}
		}
		g.config.SkipVerify = skip
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) GiteaOption {
	return func(g *gitea) {
		if g.config == nil {
			g.config = &config{}
		}
		g.config.Timeout = timeout
	}
}

// WithRetrier sets the retry mechanism for the gitea client.
func WithRetrier(retrier core.Retrier) GiteaOption {
	return func(g *gitea) {
		g.retrier = retrier
	}
}

// createHTTPClient creates and configures the HTTP client with proper TLS settings.
func (g *gitea) createHTTPClient() *http.Client {
	certs, err := x509.SystemCertPool()
	if err != nil {
		g.logger.Warn("failed to load system cert pool, using empty pool", "error", err)
		certs = x509.NewCertPool()
	}

	tlsConfig := &tls.Config{
		RootCAs: certs,
	}

	// Only skip verification if explicitly needed and log warning
	if g.config.SkipVerify {
		g.logger.Warn("TLS certificate verification disabled - this reduces security")
		tlsConfig.InsecureSkipVerify = true
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: g.config.Timeout,
	}
}

// init initializes the gitea client.
func (g *gitea) init() error {
	// Validate configuration
	if err := g.config.Validate(); err != nil {
		return err
	}

	g.config.Server = strings.TrimRight(g.config.Server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.config.Token),
	}

	// Use extracted HTTP client creation method
	httpClient := g.createHTTPClient()
	opts = append(opts, gsdk.SetHTTPClient(httpClient))

	client, err := gsdk.NewClient(g.config.Server, opts...)
	if err != nil {
		return fmt.Errorf("failed to create gitea client: %w", err)
	}

	g.client = client
	return nil
}

// NewGitea creates a new instance of the gitea struct with options pattern.
func NewGitea(
	ctx context.Context,
	server string,
	token string,
	opts ...GiteaOption,
) (core.GiteaClient, error) {
	// Create default config
	cfg := &config{
		Server:     server,
		Token:      token,
		Timeout:    30 * time.Second, // default timeout
		RetryCount: 3,                // default retry count
	}

	g := &gitea{
		ctx:    ctx,
		config: cfg,
		logger: slog.Default(), // default logger
	}

	// Apply options
	for _, opt := range opts {
		opt(g)
	}

	// Initialize default retrier if not provided
	if g.retrier == nil {
		g.retrier = retry.NewDefaultRetrier(&retry.RetryConfig{
			MaxRetries: g.config.RetryCount,
			Logger:     g.logger,
		})
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}

// CreateOrgActionSecret creates or updates an organization action secret
func (g *gitea) CreateOrgActionSecret(org string, opt gsdk.CreateSecretOption) (*gsdk.Response, error) {
	if g.client == nil {
		return nil, errors.New("gitea client not initialized")
	}

	return g.retrier.DoWithRetry(func() (*gsdk.Response, error) {
		return g.client.CreateOrgActionSecret(org, opt)
	})
}

// CreateRepoActionSecret creates or updates a repository action secret
func (g *gitea) CreateRepoActionSecret(owner, repo string, opt gsdk.CreateSecretOption) (*gsdk.Response, error) {
	if g.client == nil {
		return nil, errors.New("gitea client not initialized")
	}

	return g.retrier.DoWithRetry(func() (*gsdk.Response, error) {
		return g.client.CreateRepoActionSecret(owner, repo, opt)
	})
}

// Ping verifies the connection to Gitea server and token validity
func (g *gitea) Ping() error {
	if g.client == nil {
		return errors.New("gitea client not initialized")
	}

	_, err := g.retrier.DoWithRetry(func() (*gsdk.Response, error) {
		_, resp, err := g.client.GetMyUserInfo()
		return resp, err
	})
	if err != nil {
		return fmt.Errorf("failed to ping gitea server: %w", err)
	}

	g.logger.Info("successfully connected to gitea server", "server", g.config.Server)
	return nil
}

// Close performs cleanup operations
func (g *gitea) Close() error {
	// Add any cleanup logic here if needed
	g.logger.Debug("closing gitea client")
	return nil
}
