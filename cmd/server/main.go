// Animetsu API server entrypoint.

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/animetsu/api/internal/config"
	"github.com/animetsu/api/internal/logger"
	"github.com/animetsu/api/internal/server"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	app := server.New(cfg, log)

	// Start HTTP server in a goroutine so we can wait for shutdown signals.
	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", ":"+cfg.Port).Msg("animetsu-api listening")
		if err := app.Listen(":" + cfg.Port); err != nil {
			errCh <- err
		}
	}()

	// Wait for SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			log.Fatal().Err(err).Msg("server crashed")
		}
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
		os.Exit(1)
	}
	log.Info().Msg("bye")
}
