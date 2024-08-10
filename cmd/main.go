package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/joho/godotenv"
)

var (
	Version     string
	Commit      string
	showVersion bool
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
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s Commit: %s\n", Version, Commit)
		return
	}

	_ = godotenv.Load(envfile)

	giteaServer := getGlobalValue("gitea_server")
	giteaToken := getGlobalValue("gitea_token")
	giteaSkip := getGlobalValue("gitea_skip_verify")
	secrets := getGlobalValue("secrets")
	orgs := getGlobalValue("orgs")
	repos := getGlobalValue("repos")
	dryRun := toBool(getGlobalValue("dry_run"))

	if giteaServer == "" || giteaToken == "" {
		slog.Error("missing gitea server or token")
		return
	}

	allsecrets := getDataFromEnv(strings.Split(secrets, ","))
	if len(allsecrets) == 0 {
		slog.Error("can't find any secrets")
		return
	}

	slog.Info("gitea server", "value", giteaServer)
	if dryRun {
		slog.Warn("[DRY_RUN='true'] No changes will be written to secrets")
	}

	// init gitea client
	ctx := withContextFunc(context.Background(), func() {})
	g, err := NewGitea(
		ctx,
		giteaServer,
		giteaToken,
		toBool(giteaSkip),
		slog.New(slog.NewTextHandler(os.Stdout, nil)),
	)
	if err != nil {
		slog.Error("failed to init gitea client", "error", err)
		return
	}

	// update gitea org secrets
	orgsList := strings.Split(orgs, ",")
	for _, org := range orgsList {
		if org == "" {
			continue
		}
		for k, v := range allsecrets {
			if dryRun {
				slog.Info("update org secrets", "org", org, "secret", k)
				continue
			}
			_, err := g.client.CreateOrgActionSecret(org, gsdk.CreateSecretOption{
				Name: k,
				Data: v,
			})
			if err != nil {
				slog.Error("failed to update org secrets", "org", org, "error", err)
				break
			}
			slog.Info("update org secrets", "org", org, "secret", k)
		}
	}

	// update gitea repo secrets
	reposList := strings.Split(repos, ",")
	for _, repo := range reposList {
		if repo == "" {
			continue
		}
		// check if the repo is in the format "org/repo"
		val := strings.Split(repo, "/")
		if len(val) != 2 {
			slog.Error("invalid repo format", "repo", repo)
			continue
		}
		for k, v := range allsecrets {
			if dryRun {
				slog.Info("update repo secrets", "repo", repo, "secret", k)
				continue
			}
			_, err := g.client.CreateRepoActionSecret(val[0], val[1], gsdk.CreateSecretOption{
				Name: k,
				Data: v,
			})
			if err != nil {
				slog.Error("failed to update repo secrets", "repo", repo, "error", err)
				break
			}
			slog.Info("update repo secrets", "repo", repo, "secret", k)
		}
	}
}
