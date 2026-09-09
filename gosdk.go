package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func runGoSDK(config Config) {
	impl := &mcp.Implementation{
		Name:        "mcpcmds",
		Title:       config.Title,
		Description: config.Desc,
		Version:     "1.0.0",
	}

	s := mcp.NewServer(impl, &mcp.ServerOptions{
		HasTools: true,
	})

	addToolsGoSDK(s, config)

	ctx := context.Background()
	if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.Error("Server error", "error", err)
		fmt.Printf("Server error: %v\n", err)
	}
	slog.Info("mcpcmds exiting")
}

func addToolsGoSDK(s *mcp.Server, config Config) {
	for _, tool := range config.Tools {
		t := &mcp.Tool{
			Name:        tool.Name,
			Title:       tool.Title,
			Description: tool.Desc,
			InputSchema: buildInputSchema(tool),
		}
		s.AddTool(t, toolHandlerGoSDK(tool))
	}
}

func buildInputSchema(tool Tool) map[string]any {
	props := make(map[string]any)
	required := make([]string, 0)

	for _, a := range tool.Required {
		props[a.Name] = map[string]any{
			"type":        "string",
			"description": a.Desc,
		}
		required = append(required, a.Name)
	}
	for _, a := range tool.Optional {
		props[a.Name] = map[string]any{
			"type":        "string",
			"description": a.Desc,
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func toolHandlerGoSDK(tool Tool) func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := make(map[string]string)

		if request.Params != nil && request.Params.Arguments != nil {
			raw := make(map[string]any)
			if err := json.Unmarshal(request.Params.Arguments, &raw); err == nil {
				for _, a := range tool.Required {
					if v, ok := raw[a.Name].(string); ok {
						args[a.Name] = v
					}
				}
				for _, a := range tool.Optional {
					if v, ok := raw[a.Name].(string); ok {
						args[a.Name] = v
					} else if v, ok := raw[a.Name]; ok {
						args[a.Name] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		cmdArgs := make([]string, len(tool.Command))
		for i, arg := range tool.Command {
			cmdArgs[i] = replacePlaceholders(arg, args)
		}

		command := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := command.Output()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
		}, nil
	}
}
