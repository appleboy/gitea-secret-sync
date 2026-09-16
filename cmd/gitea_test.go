package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sync-secrets/retry"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/require"
)

func TestGiteaCreateActionSecretAllowsLongName(t *testing.T) {
	t.Parallel()

	const secretData = "secret-value"
	secretName := strings.Repeat("A", 31)

	tests := []struct {
		name string
		path string
		call func(*gitea, gsdk.CreateOrUpdateSecretOption) (*gsdk.Response, error)
	}{
		{
			name: "organization secret",
			path: "/api/v1/orgs/example/actions/secrets/" + secretName,
			call: func(client *gitea, opt gsdk.CreateOrUpdateSecretOption) (*gsdk.Response, error) {
				return client.CreateOrgActionSecret("example", secretName, opt)
			},
		},
		{
			name: "repository secret",
			path: "/api/v1/repos/example/project/actions/secrets/" + secretName,
			call: func(client *gitea, opt gsdk.CreateOrUpdateSecretOption) (*gsdk.Response, error) {
				return client.CreateRepoActionSecret("example", "project", secretName, opt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type request struct {
				method string
				path   string
				body   string
			}
			requestCh := make(chan request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				requestCh <- request{method: r.Method, path: r.URL.Path, body: string(body)}
				w.WriteHeader(http.StatusCreated)
			}))
			t.Cleanup(server.Close)

			sdkClient, err := gsdk.NewClient(server.URL, gsdk.SetGiteaVersion(""))
			require.NoError(t, err)
			client := &gitea{
				client:  sdkClient,
				retrier: retry.NewGiteaRetrier(nil),
			}

			_, err = tt.call(client, gsdk.CreateOrUpdateSecretOption{Data: secretData})
			require.NoError(t, err)

			got := <-requestCh
			require.Equal(t, http.MethodPut, got.method)
			require.Equal(t, tt.path, got.path)
			require.JSONEq(t, `{"data":"secret-value","description":""}`, got.body)
		})
	}
}
