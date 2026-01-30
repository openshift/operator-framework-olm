package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	server "github.com/openshift/operator-framework-olm/pkg/lifecycle-server"
)

const (
	defaultFBCPath     = "/catalog/configs"
	defaultListenAddr  = ":8080"
	defaultHealthAddr  = ":8081"
	shutdownTimeout    = 10 * time.Second
)

func main() {
	cmd := newStartCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "encountered an error while executing the binary: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	fbcPath, err := cmd.Flags().GetString("fbc-path")
	if err != nil {
		return err
	}
	listenAddr, err := cmd.Flags().GetString("listen")
	if err != nil {
		return err
	}
	healthAddr, err := cmd.Flags().GetString("health")
	if err != nil {
		return err
	}

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.Info("starting lifecycle-server")

	// Load lifecycle data from FBC
	log.WithField("path", fbcPath).Info("loading lifecycle data from FBC")
	data, err := server.LoadLifecycleData(fbcPath)
	if err != nil {
		log.WithError(err).Warn("failed to load lifecycle data, starting with empty data")
		data = make(server.LifecycleIndex)
	}
	log.WithFields(logrus.Fields{
		"blobCount":    server.CountBlobs(data),
		"versionCount": len(data),
		"versions":     server.ListVersions(data),
	}).Info("loaded lifecycle data")

	// Create HTTP handler
	handler := server.NewHandler(data, log)

	// Create health handler
	healthHandler := http.NewServeMux()
	healthHandler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Create HTTP servers
	apiServer := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}
	healthServer := &http.Server{
		Addr:    healthAddr,
		Handler: healthHandler,
	}

	// Start servers
	errCh := make(chan error, 2)
	go func() {
		log.WithField("addr", listenAddr).Info("starting API server")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("API server error: %w", err)
		}
	}()
	go func() {
		log.WithField("addr", healthAddr).Info("starting health server")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("health server error: %w", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.WithField("signal", sig).Info("received shutdown signal")
	case err := <-errCh:
		log.WithError(err).Error("server error")
		return err
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Info("shutting down servers")
	if err := apiServer.Shutdown(ctx); err != nil {
		log.WithError(err).Error("API server shutdown error")
	}
	if err := healthServer.Shutdown(ctx); err != nil {
		log.WithError(err).Error("health server shutdown error")
	}

	return nil
}
