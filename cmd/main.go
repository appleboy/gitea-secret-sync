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

// setupGracefulShutdown sets up graceful shutdown handling with proper cleanup
func setupGracefulShutdown(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	cleanup := func() {
		signal.Stop(c)
		cancel()
	}

	go func() {
		defer cleanup()
		select {
		case <-ctx.Done():
			return
		case sig := <-c:
			slog.Info("Received shutdown signal", "signal", sig)
			cancel()
		}
	}()

	return ctx, cleanup
}

// logError logs an error using structured logging and returns a simple error for the caller to handle
func logError(msg string, args ...interface{}) error {
	slog.Error(msg, args...)
	return fmt.Errorf("%s", msg)
}

// validateRepoFormat validates that repo is in "org/repo" format
func validateRepoFormat(repo string) bool {
	if repo == "" {
		return false
	}
	parts := strings.Split(repo, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// validateConfiguration validates required configuration parameters
func validateConfiguration(giteaServer, giteaToken string, secrets map[string]string) error {
	if giteaServer == "" || giteaToken == "" {
		return logError("missing gitea server or token")
	}
	if len(secrets) == 0 {
		return logError("can't find any secrets")
	}
	return nil
}

// processOrgs processes organization secrets
func processOrgs(g *gitea, orgs string, secrets map[string]string, description string, dryRun bool) {
	orgsList := splitByCommaOrNewline(orgs)
	for _, org := range orgsList {
		org = strings.TrimSpace(org)
		if org == "" {
			continue
		}
		for k, v := range secrets {
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
}

// processRepos processes repository secrets
func processRepos(g *gitea, repos string, secrets map[string]string, description string, dryRun bool) {
	reposList := splitByCommaOrNewline(repos)
	for _, repo := range reposList {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		// check if the repo is in the format "org/repo"
		if !validateRepoFormat(repo) {
			slog.Error("invalid repo format, expected 'org/repo'", "repo", repo)
			continue
		}
		val := strings.Split(repo, "/")

		for k, v := range secrets {
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

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s Commit: %s\n", Version, Commit)
		return nil
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

	allsecrets := getDataFromEnv(splitByCommaOrNewline(secrets))
	if err := validateConfiguration(giteaServer, giteaToken, allsecrets); err != nil {
		return err
	}

	if dryRun {
		slog.Warn("[DRY_RUN='true'] No changes will be written to secrets")
	}

	// init gitea client with graceful shutdown
	ctx, cleanup := setupGracefulShutdown(context.Background())
	defer cleanup()

	g, err := NewGitea(
		ctx,
		giteaServer,
		giteaToken,
		toBool(giteaSkip),
		logger,
	)
	if err != nil {
		return logError("failed to init gitea client", "error", err)
	}

	// update gitea org secrets
	processOrgs(g, orgs, allsecrets, description, dryRun)

	// update gitea repo secrets
	processRepos(g, repos, allsecrets, description, dryRun)

	return nil
}
