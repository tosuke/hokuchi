package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

var (
	flagHelp      bool
	flagHttpAddr  string
	flagLogLevel  string
	flagDataPath  string
	flagCachePath string
)

func init() {
	flag.BoolVar(&flagHelp, "help", false, "print usage and exit")
	flag.StringVar(&flagHttpAddr, "http.address", "127.0.0.1:8080", "HTTP server listen address")
	flag.StringVar(&flagLogLevel, "log.level", "info", "logging level")
	flag.StringVar(&flagDataPath, "data.path", "/var/lib/hokuchi", "data directory")
	flag.StringVar(&flagCachePath, "cache.path", "/tmp", "cache directory")
}

type config struct {
	LogLevel   slog.Level
	HttpAddr   string
	AssetsPath string
	DataPath   string
	CachePath  string
}

func initConfig() config {
	httpAddr := os.Getenv("HOKUCHI_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = flagHttpAddr
	}

	logLevelStr := os.Getenv("HOKUCHI_LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = flagLogLevel
	}
	logLevel := new(slog.Level)
	if err := logLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
		fmt.Printf("cannot parse log level: %s\n", logLevelStr)
		*logLevel = slog.LevelInfo
	}

	dataPath := os.Getenv("HOKUCHI_DATA_PATH")
	if dataPath == "" {
		dataPath = flagDataPath
	}

	cachePath := os.Getenv("HOKUCHI_CACHE_PATH")
	if cachePath == "" {
		cachePath = flagCachePath
	}

	return config{
		HttpAddr:   httpAddr,
		LogLevel:   *logLevel,
		AssetsPath: "assets",
		DataPath:   dataPath,
		CachePath:  cachePath,
	}
}
