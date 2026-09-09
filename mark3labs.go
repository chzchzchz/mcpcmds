package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runMark3Labs(config Config) {
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

	// Start the stdio server with logging
	if err := serveStdioWithLogging(s); err != nil {
		slog.Error("Server error", "error", err)
		fmt.Printf("Server error: %v\n", err)
	}
	slog.Info("mcpcmds exiting")
}

// loggingWriter wraps an io.Writer to log all outgoing data.
type loggingWriter struct {
	io.Writer
}

func (w *loggingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if n > 0 {
		dbg(fmt.Sprintf("stdout send %d bytes: %s", n, string(p[:n])))
	}
	return n, err
}

// loggingReader wraps an io.Reader to log all incoming data.
type loggingReader struct {
	io.Reader
}

func (r *loggingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		dbg(fmt.Sprintf("stdin recv %d bytes: %s", n, string(p[:n])))
	}
	if err != nil && err != io.EOF {
		dbg(fmt.Sprintf("stdin read error: %v", err))
	}
	return n, err
}

// serveStdioWithLogging starts the stdio server with full request/response logging.
func serveStdioWithLogging(srv *server.MCPServer) error {
	stdioSrv := server.NewStdioServer(srv)
	stdioSrv.SetErrorLogger(log.New(os.Stderr, "", log.LstdFlags))
	dbg("stdio server created")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		dbg("signal received, shutting down")
		cancel()
	}()

	dbg("stdio server listening on stdin")
	dbg("calling Listen...")
	return stdioSrv.Listen(ctx, &loggingReader{Reader: os.Stdin}, &loggingWriter{Writer: os.Stdout})
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
