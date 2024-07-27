package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/joho/godotenv"
)

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key) // Convert key to uppercase

	// If the "PLUGIN_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.Parse()

	_ = godotenv.Load(envfile)

	giteaServer := getGlobalValue("gitea_server")
	giteaToken := getGlobalValue("gitea_token")
	giteaSkip := getGlobalValue("gitea_skip_verify")
	giteaOrg := getGlobalValue("gitea_org")

	// init gitea client
	ctx := withContextFunc(context.Background(), func() {})
	g := &gitea{
		ctx:        ctx,
		server:     giteaServer,
		token:      giteaToken,
		skipVerify: ToBool(giteaSkip),
		logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	err := g.init()
	if err != nil {
		slog.Error("failed to init gitea client", "error", err)
		return
	}

	rows, _, err := g.client.ListOrgActionSecret(giteaOrg, gsdk.ListOrgActionSecretOption{
		ListOptions: gsdk.ListOptions{
			Page:     1,
			PageSize: 10,
		},
	})
	if err != nil {
		slog.Error("failed to list org action secrets", "error", err)
		return
	}

	for _, row := range rows {
		slog.Info(
			"get secret",
			"name",
			row.Name,
			"value",
			row.Data,
			"created",
			row.Created,
		)
	}
}
