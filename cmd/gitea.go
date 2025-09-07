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

	gsdk "code.gitea.io/sdk/gitea"
)

// gitea is a struct that holds the gitea client.
type gitea struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	client     *gsdk.Client
	logger     *slog.Logger
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

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
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

// NewGitea creates a new instance of the gitea struct.
func NewGitea(
	ctx context.Context,
	server string,
	token string,
	skipVerify bool,
	logger *slog.Logger,
) (*gitea, error) {
	g := &gitea{
		ctx:        ctx,
		server:     server,
		token:      token,
		skipVerify: skipVerify,
		logger:     logger,
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}
