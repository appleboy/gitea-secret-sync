package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

// init initializes the gitea client.
func (g *gitea) init() (err error) {
	if g.server == "" || g.token == "" {
		return errors.New("mission gitea server or token")
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
	}

	// add new http client for skip verify
	certs, _ := x509.SystemCertPool()
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certs,
				InsecureSkipVerify: g.skipVerify,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
	opts = append(opts, gsdk.SetHTTPClient(httpClient))

	g.client, err = gsdk.NewClient(g.server, opts...)
	if err != nil {
		return err
	}

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
