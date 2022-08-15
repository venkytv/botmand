package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/duh-uh/teabot/backend"
	"github.com/duh-uh/teabot/conversation"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

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
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// XXX: Get port from argv
		err := http.ListenAndServe(":2112", nil)
		if errors.Is(err, http.ErrServerClosed) {
			logrus.Info("Metrics server shutdown")
		} else {
			logrus.Warnf("Error starting metrics server: %s", err)
		}
	}()

	homedir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Error getting user home directory: %v", err)
	}

	tokenFile := path.Join(homedir, ".botters.token")
	api := backend.NewSlackApi(tokenFile)

	beqs := backend.NewBackendQueues()
	be := backend.NewSlackBackend(&api, &beqs)
	cm := conversation.NewManager(be, beqs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cm.Start(ctx)

	<-done
	logrus.Debug("Exiting")
}
