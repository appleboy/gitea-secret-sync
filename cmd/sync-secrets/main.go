package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	gsdk "code.gitea.io/sdk/gitea"
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

func main() {
	giteaServer := flag.String("gitea-server", "", "gitea server url")
	giteaToken := flag.String("gitea-token", "", "gitea access token")
	giteaSkip := flag.Bool("gitea-skip-verify", false, "skip verify gitea server")
	giteaOrg := flag.String("gitea-org", "", "gitea organization")

	flag.Parse()

	// init gitea client
	ctx := withContextFunc(context.Background(), func() {})
	g := &gitea{
		ctx:        ctx,
		server:     PtrToString(giteaServer),
		token:      PtrToString(giteaToken),
		skipVerify: PtrToBool(giteaSkip),
		logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	err := g.init()
	if err != nil {
		slog.Error("failed to init gitea client", "error", err)
		return
	}

	rows, _, err := g.client.ListOrgActionSecret(PtrToString(giteaOrg), gsdk.ListOrgActionSecretOption{
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
