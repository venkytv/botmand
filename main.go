package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/venkytv/botmand/backend"
	"github.com/venkytv/botmand/conversation"
	"github.com/venkytv/botmand/globals"

	"github.com/urfave/cli/v2"
)

var version = "local-build"
var date = "unknown"

// Set up signal handling
func handleSignals(ctx context.Context, cm *conversation.Manager, cfg *cli.Context, done chan bool) {
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

		break
	}
	done <- true

}

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Error getting user home directory: %v", err)
	}

	defaultTokenFile := path.Join(homedir, ".slack.token")
	defaultBotDirectory := path.Join(homedir, "botmand-engines")

	app := &cli.App{
		Name:    globals.BotName,
		Usage:   "A slack bot for running conversations",
		Version: fmt.Sprintf("%s (%s)", version, date),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-directory",
				Usage:   "directory containing config files",
				Value:   defaultBotDirectory,
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:  "slack-backend-token",
				Usage: "api token for slack backend",
			},
			&cli.StringFlag{
				Name:    "slack-backend-token-file",
				Usage:   "file containing slack backend api token",
				Value:   defaultTokenFile,
				Aliases: []string{"t"},
			},
			&cli.BoolFlag{
				Name:    "enable-metrics",
				Usage:   "enable prometheus-style metrics",
				Aliases: []string{"m"},
			},
			&cli.IntFlag{
				Name:    "metrics-port",
				Usage:   "metrics port",
				Value:   2112,
				Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "print debug messages",
				Aliases: []string{"d"},
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("version") {
				fmt.Printf("%s %s (%s)", globals.BotName, version, date)
				return nil
			}
			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}
			if c.Bool("enable-metrics") {
				go func() {
					http.Handle("/metrics", promhttp.Handler())
					err := http.ListenAndServe(fmt.Sprintf(":%d", c.Int("metrics-port")), nil)
					if errors.Is(err, http.ErrServerClosed) {
						logrus.Info("Metrics server shutdown")
					} else {
						logrus.Warnf("Error starting metrics server: %s", err)
					}
				}()
			}

			apiToken := c.String("slack-backend-token")
			if len(apiToken) < 1 {
				// Load API token from file
				apiTokenFile := c.String("slack-backend-token-file")
				content, err := ioutil.ReadFile(apiTokenFile)
				if err != nil {
					logrus.Fatalf("Failed to open slack token file: %s: %s", apiTokenFile, err)
				}
				apiToken = strings.TrimSpace(string(content))
			}

			api := backend.NewSlackApi(apiToken, c.Bool("debug"))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			beqs := backend.NewBackendQueues()
			be := backend.NewSlackBackend(&api, &beqs)
			cm := conversation.NewManager(ctx, c, be, beqs)

			done := make(chan bool)
			go handleSignals(ctx, cm, c, done)
			go cm.Start(ctx)

			<-done
			logrus.Debug("Exiting")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
