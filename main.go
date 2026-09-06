package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "configuration file path")
	flag.StringVar(&configPath, "c", "", "configuration file path (shorthand)")
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
		slog.SetDefault(slog.New(slog.NewTextHandler(f, nil)))
		slog.Info("logging initialized", "log-file", logFilePath)
	}

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

	// Create a new MCP server
	s := server.NewMCPServer(
		"mcpcmds",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add tool
	tool := mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)

	// Add tool handler
	s.AddTool(tool, helloHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}
