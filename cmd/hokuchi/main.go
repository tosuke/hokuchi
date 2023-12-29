package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/tosuke/hokuchi/flatcar"
	"github.com/tosuke/hokuchi/server"
	"github.com/tosuke/hokuchi/slogerr"
	"github.com/tosuke/hokuchi/storage"
)

func main() {
	status := run(os.Args)
	if status != 0 {
		os.Exit(status)
	}
}

func run(args []string) int {
	flag.CommandLine.Parse(args[1:])
	if flagHelp {
		flag.Usage()
		return 0
	}

	cfg := initConfig()

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     cfg.LogLevel,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)

	storage := storage.NewFSStorage(cfg.DataPath, cfg.CachePath)
	defer storage.Close()

	server := &server.Server{
		Logger:     logger,
		AssetsPath: cfg.AssetsPath,

		Flatcar: flatcar.New(flatcar.Option{
			RequestConcurrency: 8,
		}),
		Storage: storage,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	slog.Info(fmt.Sprintf("starting HTTP server on %s", cfg.HttpAddr))
	hs := &http.Server{
		Addr:    cfg.HttpAddr,
		Handler: server.HTTPHandler(),
	}
	defer hs.Close()
	go func() {
		if err := hs.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			cancel(fmt.Errorf("failed to start server: %w", err))
		}
		cancel(nil)
	}()

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		slog.Error("failed to start", slogerr.Err(err))
		return 1
	}

	// graceful shutdown
	slog.Info("shutting down gracefully")
	gracefulCtx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	if err := hs.Shutdown(gracefulCtx); err != nil {
		slog.Error("failed to shutdown the server", slogerr.Err(err))
		return 1
	}

	return 0
}
