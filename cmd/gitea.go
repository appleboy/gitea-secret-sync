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

	gsdk "code.gitea.io/sdk/gitea"
	"sync-secrets/core"
)

// gitea is a struct that holds the gitea client and implements core.GiteaClient interface.
type gitea struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	timeout    time.Duration
	client     *gsdk.Client
	logger     *slog.Logger
}

// Ensure gitea implements core.GiteaClient interface
var _ core.GiteaClient = (*gitea)(nil)

// GiteaOption is a function type for configuring gitea client options.
type GiteaOption func(*gitea)

// WithSkipVerify sets whether to skip TLS certificate verification.
func WithSkipVerify(skip bool) GiteaOption {
	return func(g *gitea) {
		g.skipVerify = skip
	}
}

// WithLogger sets the logger for the gitea client.
func WithLogger(logger *slog.Logger) GiteaOption {
	return func(g *gitea) {
		g.logger = logger
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) GiteaOption {
	return func(g *gitea) {
		g.timeout = timeout
	}
}

// validateInputs performs comprehensive validation of input parameters.
func (g *gitea) validateInputs() error {
	if g.server == "" {
		return errors.New("gitea server URL is required")
	}
	if g.token == "" {
		return errors.New("gitea token is required")
	}
	// Gitea tokens are typically 40 characters (hex format)
	if len(g.token) < 40 {
		return errors.New("gitea token appears invalid (expected 40 characters)")
	}
	return nil
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
	if g.skipVerify {
		g.logger.Warn("TLS certificate verification disabled - this reduces security")
		tlsConfig.InsecureSkipVerify = true
	}

	// Set default timeout if not configured
	timeout := g.timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: timeout,
	}
}

// init initializes the gitea client.
func (g *gitea) init() error {
	// Use extracted validation method
	if err := g.validateInputs(); err != nil {
		return err
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
	}

	// Use extracted HTTP client creation method
	httpClient := g.createHTTPClient()
	opts = append(opts, gsdk.SetHTTPClient(httpClient))

	client, err := gsdk.NewClient(g.server, opts...)
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
	g := &gitea{
		ctx:     ctx,
		server:  server,
		token:   token,
		logger:  slog.Default(),   // default logger
		timeout: 30 * time.Second, // default timeout
	}

	// Apply options
	for _, opt := range opts {
		opt(g)
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}

// CreateOrgActionSecret creates or updates an organization action secret
func (g *gitea) CreateOrgActionSecret(org string, opt gsdk.CreateSecretOption) (interface{}, error) {
	if g.client == nil {
		return nil, errors.New("gitea client not initialized")
	}
	return g.client.CreateOrgActionSecret(org, opt)
}

// CreateRepoActionSecret creates or updates a repository action secret
func (g *gitea) CreateRepoActionSecret(owner, repo string, opt gsdk.CreateSecretOption) (interface{}, error) {
	if g.client == nil {
		return nil, errors.New("gitea client not initialized")
	}
	return g.client.CreateRepoActionSecret(owner, repo, opt)
}

// Ping verifies the connection to Gitea server and token validity
func (g *gitea) Ping() error {
	if g.client == nil {
		return errors.New("gitea client not initialized")
	}

	// Use a lightweight endpoint to test connectivity
	_, _, err := g.client.GetMyUserInfo()
	if err != nil {
		return fmt.Errorf("failed to ping gitea server: %w", err)
	}

	g.logger.Info("successfully connected to gitea server", "server", g.server)
	return nil
}

// Close performs cleanup operations
func (g *gitea) Close() error {
	// Add any cleanup logic here if needed
	g.logger.Debug("closing gitea client")
	return nil
}
