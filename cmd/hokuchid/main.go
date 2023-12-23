package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/tosuke/hokuchi/server"
)

func main() {
	flags := struct {
		help       bool
		httpAddr   string
		assetsPath string
	}{}
	flag.BoolVar(&flags.help, "help", false, "print usage and exit")
	flag.StringVar(&flags.httpAddr, "http.address", "127.0.0.1:8080", "HTTP server listen address")
	flag.StringVar(&flags.assetsPath, "assets.path", "assets", "path to assets")
	flag.Parse()

	if flags.help {
		flag.Usage()
		return
	}

	httpAddr := flags.httpAddr
    assetsPath := flags.assetsPath

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)

	server := &server.Server{
		Logger: logger,
        AssetsPath: assetsPath,
	}

	slog.InfoContext(ctx, fmt.Sprintf("starting HTTP server on %s", httpAddr))
	if err := server.Start(ctx, httpAddr); err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}
