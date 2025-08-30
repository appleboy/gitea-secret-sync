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
	debugMode   bool
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

// fatalError logs an error and exits with code 1
func fatalError(msg string, args ...interface{}) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
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
	description := getGlobalValue("description")
	dryRun := toBool(getGlobalValue("dry_run"))
	debugFromEnv := toBool(getGlobalValue("debug"))

	// Use debug mode from command line flag OR environment variable
	if !debugMode {
		debugMode = debugFromEnv
	}

	// Configure slog level based on debug mode
	var logLevel slog.Level
	if debugMode {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Set the default logger
	slog.SetDefault(logger)

	slog.Info("gitea server", "value", giteaServer)

	if debugMode {
		slog.Debug("=== Debug Mode: Configuration Values ===")
		slog.Debug("--- Command Line Flags ---")
		slog.Debug("flag.env-file", "value", envfile)
		slog.Debug("flag.version", "value", showVersion)
		slog.Debug("flag.debug", "value", debugMode)
		slog.Debug("--- Environment Variables ---")
		slog.Debug("gitea_server", "value", giteaServer)
		slog.Debug("gitea_token", "value", "***HIDDEN***")
		slog.Debug("gitea_skip_verify", "value", giteaSkip)
		slog.Debug("secrets", "value", secrets)
		slog.Debug("orgs", "value", orgs)
		slog.Debug("repos", "value", repos)
		slog.Debug("description", "value", description)
		slog.Debug("dry_run", "value", dryRun)
		slog.Debug("debug", "value", debugMode)
		slog.Debug("=========================================")
	}

	if giteaServer == "" || giteaToken == "" {
		fatalError("missing gitea server or token")
	}

	allsecrets := getDataFromEnv(strings.Split(secrets, ","))
	if len(allsecrets) == 0 {
		fatalError("can't find any secrets")
	}

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
		logger,
	)
	if err != nil {
		fatalError("failed to init gitea client", "error", err)
	}

	// update gitea org secrets
	orgsList := strings.Split(orgs, ",")
	for _, org := range orgsList {
		org = strings.TrimSpace(org)
		if org == "" {
			continue
		}
		for k, v := range allsecrets {
			if dryRun {
				slog.Info("update org secrets", "org", org, "secret", k)
				continue
			}
			_, err := g.client.CreateOrgActionSecret(org, gsdk.CreateSecretOption{
				Name:        k,
				Data:        v,
				Description: description,
			})
			if err != nil {
				slog.Error(
					"failed to update org secrets",
					"org", org,
					"secret", k,
					"error", err,
				)
				continue
			}
			slog.Info("update org secrets", "org", org, "secret", k)
		}
	}

	// update gitea repo secrets
	reposList := strings.Split(repos, ",")
	for _, repo := range reposList {
		repo = strings.TrimSpace(repo)
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
				Name:        k,
				Data:        v,
				Description: description,
			})
			if err != nil {
				slog.Error(
					"failed to update repo secrets",
					"repo", repo,
					"secret", k,
					"error", err,
				)
				continue
			}
			slog.Info("update repo secrets", "repo", repo, "secret", k)
		}
	}
}
