package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const logDebug = true

func dbg(msg string) {
	if logDebug {
		slog.Debug(msg)
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "configuration file path")
	flag.StringVar(&configPath, "c", "", "configuration file path (shorthand)")
	var mode string
	flag.StringVar(&mode, "mode", "mark3labs", "server implementation mode: mark3labs or gosdk")
	var logFilePath string
	flag.StringVar(&logFilePath, "log-file", "", "log file path")
	flag.Parse()

	if logFilePath != "" {
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	slog.Info("logging initialized", "log-file", logFilePath)

	slog.Info("mcpcmds starting", "pid", os.Getpid(), "mode", mode)

	var config Config
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			slog.Error("failed to read config file", "error", err, "path", configPath)
			fmt.Printf("Failed to read config file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &config); err != nil {
			slog.Error("failed to parse config file", "error", err, "path", configPath)
			fmt.Printf("Failed to parse config file: %v\n", err)
			os.Exit(1)
		}
		slog.Info("config loaded", "path", configPath, "tools", len(config.Tools))
	}

	switch mode {
	case "gosdk":
		runGoSDK(config)
	default:
		runMark3Labs(config)
	}
}
