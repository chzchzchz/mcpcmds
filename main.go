package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

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
		server.WithToolCapabilities(true),
		server.WithTitle(config.Title),
		server.WithDescription(config.Desc),
	)

	// Add tools from config
	addTools(s, config)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func toolHandler(tool Tool) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := make(map[string]string)
		for _, a := range tool.Required {
			val, err := request.RequireString(a.Name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args[a.Name] = val
		}
		for _, a := range tool.Optional {
			args[a.Name] = mcp.ParseString(request, a.Name, "")
		}

		cmdArgs := make([]string, len(tool.Command))
		for i, arg := range tool.Command {
			cmdArgs[i] = replacePlaceholders(arg, args)
		}

		command := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := command.Output()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(output)), nil
	}
}

func replacePlaceholders(s string, args map[string]string) string {
	var b strings.Builder
	for {
		start := strings.Index(s, "${")
		if start == -1 {
			b.WriteString(s)
			break
		}
		end := strings.Index(s[start+2:], "}")
		if end == -1 {
			b.WriteString(s)
			break
		}
		key := s[start+2 : start+2+end]
		b.WriteString(s[:start])
		if val, ok := args[key]; ok {
			b.WriteString(val)
		}
		s = s[start+2+end+1:]
	}
	return b.String()
}

func addTools(s *server.MCPServer, config Config) {
	for _, tool := range config.Tools {
		opts := []mcp.ToolOption{mcp.WithDescription(tool.Desc)}
		for _, a := range tool.Required {
			opts = append(opts, mcp.WithString(a.Name, mcp.Required(), mcp.Description(a.Desc)))
		}
		for _, a := range tool.Optional {
			opts = append(opts, mcp.WithString(a.Name, mcp.Description(a.Desc)))
		}
		s.AddTool(mcp.NewTool(tool.Name, opts...), toolHandler(tool))
	}
}
