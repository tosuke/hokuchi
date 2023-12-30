package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"braces.dev/errtrace"
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
	defer server.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	slog.Info(fmt.Sprintf("starting HTTP server on %s", cfg.HttpAddr))
	go func() {
		err := server.Start(cfg.HttpAddr)
		cancel(errtrace.Wrap(err))
	}()

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		slog.Error("Error starting", slogerr.Err(err))
		return 1
	}

	// graceful shutdown
	slog.Info("shutting down gracefully")
	gracefulCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if err := server.Shutdown(gracefulCtx); err != nil {
		slog.Error("Error shutting down", slogerr.Err(err))
		return 1
	}

	return 0
}
