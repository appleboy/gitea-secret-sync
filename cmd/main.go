package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	// Check if there is an environment variable with the format "INPUT_<KEY>"
	if value := os.Getenv("INPUT_" + key); value != "" {
		return value // Return the value of the "INPUT_<KEY>" environment variable
	}

	// If the "INPUT_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}

func getSecretsFromEnv(secrets []string) map[string]string {
	secretsMap := make(map[string]string)
	for _, secret := range secrets {
		val := getGlobalValue(secret)
		if val == "" {
			continue
		}
		secretsMap[secret] = val
	}
	return secretsMap
}

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.Parse()

	_ = godotenv.Load(envfile)

	giteaServer := getGlobalValue("gitea_server")
	giteaToken := getGlobalValue("gitea_token")
	giteaSkip := getGlobalValue("gitea_skip_verify")
	secrets := getGlobalValue("secrets")

	if giteaServer == "" || giteaToken == "" {
		slog.Error("missing gitea server or token")
		return
	}

	allsecrets := getSecretsFromEnv(strings.Split(secrets, ","))
	if len(allsecrets) == 0 {
		slog.Error("can't find any secrets")
		return
	}

	slog.Info("gitea server", "value", giteaServer)

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
}
