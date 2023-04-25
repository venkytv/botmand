package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/duh-uh/teabot/backend"
	"github.com/duh-uh/teabot/conversation"
	"github.com/duh-uh/teabot/globals"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/venkytv/go-config"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Error getting user home directory: %v", err)
	}

	defaultTokenFile := path.Join(homedir, ".botters.token")
	defaultBotDirectory := path.Join(homedir, "teabot-engines")

	// Load config
	flag.Bool("enable-metrics", false, "enable prometheus-style metrics")
	flag.Int("metrics-port", 2112, "metrics port")
	flag.Bool("debug", false, "print debug messages")
	flag.String("config-directory", defaultBotDirectory, "config directory")
	flag.String("slack-backend-token", "", "api token for slack backend")
	flag.String("slack-backend-token-file", defaultTokenFile, "file containing slack backend api token")

	cfg := config.Load(nil, globals.BotName)

	debug := false
	if cfg.GetBool("debug") {
		debug = true
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		logrus.Debug("Caught signal: ", sig)
		done <- true
	}()

	// Metrics interface
	if cfg.GetBool("enable-metrics") {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("metrics-port")), nil)
			if errors.Is(err, http.ErrServerClosed) {
				logrus.Info("Metrics server shutdown")
			} else {
				logrus.Warnf("Error starting metrics server: %s", err)
			}
		}()
	}

	apiToken := cfg.GetString("slack-backend-token")
	if len(apiToken) < 1 {
		// Load API token from file
		apiTokenFile := cfg.GetString("slack-backend-token-file")
		content, err := ioutil.ReadFile(apiTokenFile)
		if err != nil {
			logrus.Fatal("Failed to open config file:", apiTokenFile, err)
		}
		apiToken = strings.TrimSpace(string(content))
	}

	api := backend.NewSlackApi(apiToken, debug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	beqs := backend.NewBackendQueues()
	be := backend.NewSlackBackend(&api, &beqs)
	cm := conversation.NewManager(ctx, cfg, be, beqs)

	// Set up signal handling
	go func(ctx context.Context, cm *conversation.Manager, cfg *config.Config) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		for {
			select {
			case sig := <-sigs:
				logrus.Debug("Caught signal: ", sig)
				switch sig {
				case syscall.SIGHUP:
					logrus.Info("Reloading engines")
					cm.LoadEngines(ctx, cfg)
					continue
				}

			case <-ctx.Done():
				logrus.Debug("Shutting down")
			}

			os.Exit(0)
		}

	}(ctx, cm, cfg)

	go cm.Start(ctx)

	<-done
	logrus.Debug("Exiting")
}
