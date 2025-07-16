package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/llm-tools/tools/batch_read_file"
	"github.com/xhd2015/llm-tools/tools/create_file"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/delete_file"
	"github.com/xhd2015/llm-tools/tools/edit_file"
	"github.com/xhd2015/llm-tools/tools/file_search"
	"github.com/xhd2015/llm-tools/tools/get_workspace_root"
	"github.com/xhd2015/llm-tools/tools/grep_search"
	"github.com/xhd2015/llm-tools/tools/list_dir"
	"github.com/xhd2015/llm-tools/tools/mcp_client"
	"github.com/xhd2015/llm-tools/tools/read_file"
	"github.com/xhd2015/llm-tools/tools/rename_file"
	"github.com/xhd2015/llm-tools/tools/run_terminal_cmd"
	"github.com/xhd2015/llm-tools/tools/search_replace"
	"github.com/xhd2015/llm-tools/tools/send_answer"
	"github.com/xhd2015/llm-tools/tools/todo_write"
	"github.com/xhd2015/llm-tools/tools/web_search"
	"github.com/xhd2015/llm-tools/tools/write_file"
)

const help = `
llm-tools-mcp

Usage: llm-tools-mcp [OPTIONS]

Options:
  -h, --help                       show help
  -v,--verbose                     show verbose info
  --port PORT                      serve via HTTP/SSE on specified port (default: stdio)

Examples:
  llm-tools-mcp help               show help message
  llm-tools-mcp                    serve via stdio (default)
  llm-tools-mcp --port 8080        serve via HTTP/SSE on port 8080

  # inspect mcp via stdio
  echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | llm-tools-mcp

  # inspect mcp via HTTP/SSE
  curl http://localhost:8080/sse

Local development:
  go build -o $GOPATH/bin/llm-tools-mcp ./
`

// ToolRegistry defines a tool with its definition and execution function
type ToolRegistry struct {
	GetDefinition   func() defs.ToolDefinition
	ExecuteFromJSON func(string) (string, error)
}

// toolRegistry contains all available tools
var toolRegistry = map[string]ToolRegistry{
	"get_workspace_root": {
		GetDefinition:   get_workspace_root.GetToolDefinition,
		ExecuteFromJSON: get_workspace_root.ExecuteFromJSON,
	},
	"batch_read_file": {
		GetDefinition:   batch_read_file.GetToolDefinition,
		ExecuteFromJSON: batch_read_file.ExecuteFromJSON,
	},
	"write_file": {
		GetDefinition:   write_file.GetToolDefinition,
		ExecuteFromJSON: write_file.ExecuteFromJSON,
	},
	"read_file": {
		GetDefinition:   read_file.GetToolDefinition,
		ExecuteFromJSON: read_file.ExecuteFromJSON,
	},
	"grep_search": {
		GetDefinition:   grep_search.GetToolDefinition,
		ExecuteFromJSON: grep_search.ExecuteFromJSON,
	},
	"list_dir": {
		GetDefinition:   list_dir.GetToolDefinition,
		ExecuteFromJSON: list_dir.ExecuteFromJSON,
	},
	"run_terminal_cmd": {
		GetDefinition:   run_terminal_cmd.GetToolDefinition,
		ExecuteFromJSON: run_terminal_cmd.ExecuteFromJSON,
	},
	"create_file": {
		GetDefinition:   create_file.GetToolDefinition,
		ExecuteFromJSON: create_file.ExecuteFromJSON,
	},
	"rename_file": {
		GetDefinition:   rename_file.GetToolDefinition,
		ExecuteFromJSON: rename_file.ExecuteFromJSON,
	},
	"edit_file": {
		GetDefinition:   edit_file.GetToolDefinition,
		ExecuteFromJSON: edit_file.ExecuteFromJSON,
	},
	"search_replace": {
		GetDefinition:   search_replace.GetToolDefinition,
		ExecuteFromJSON: search_replace.ExecuteFromJSON,
	},
	"send_answer": {
		GetDefinition:   send_answer.GetToolDefinition,
		ExecuteFromJSON: send_answer.ExecuteFromJSON,
	},
	"mcp_client": {
		GetDefinition:   mcp_client.GetToolDefinition,
		ExecuteFromJSON: mcp_client.ExecuteFromJSON,
	},
	"file_search": {
		GetDefinition:   file_search.GetToolDefinition,
		ExecuteFromJSON: file_search.ExecuteFromJSON,
	},
	"delete_file": {
		GetDefinition:   delete_file.GetToolDefinition,
		ExecuteFromJSON: delete_file.ExecuteFromJSON,
	},
	"web_search": {
		GetDefinition:   web_search.GetToolDefinition,
		ExecuteFromJSON: web_search.ExecuteFromJSON,
	},
	"todo_write": {
		GetDefinition:   todo_write.GetToolDefinition,
		ExecuteFromJSON: todo_write.ExecuteFromJSON,
	},
}

func main() {
	port, err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"LLM Tools MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Add all tools from registry
	for toolName, toolReg := range toolRegistry {
		addTool(s, toolName, toolReg)
	}

	// Start the server based on port configuration
	if port > 0 {
		sse := server.NewSSEServer(s, server.WithBaseURL(fmt.Sprintf("http://localhost:%d", port)))
		err = sse.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatalf("SSE server error: %v", err)
		}
	} else {
		// Serve via stdio (default)
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Stdio server error: %v", err)
		}
	}
}

func handle(args []string) (int, error) {
	var port int
	args, err := flags.New().Help("-h,--help", help).
		Int("--port", &port).
		Parse(args)
	if err != nil {
		return 0, err
	}
	if len(args) > 0 {
		cmd := args[0]
		args = args[1:]
		if cmd == "--help" || cmd == "help" {
			fmt.Print(strings.TrimPrefix(help, "\n"))
			os.Exit(0)
			return 0, nil
		}
		if len(args) > 0 {
			return 0, fmt.Errorf("unrecognized extra arguments: %s", strings.Join(args, ","))
		}
	}
	return port, nil
}

// serveHTTP starts an HTTP server with SSE support
func serveHTTP(s *server.MCPServer, port int) error {
	mux := http.NewServeMux()

	// SSE endpoint
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Send initial SSE message
		fmt.Fprintf(w, "data: %s\n\n", `{"jsonrpc": "2.0", "method": "notifications/initialized"}`)

		// Keep connection alive
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		flusher.Flush()

		// Handle the SSE connection
		// For now, we'll keep it simple and just maintain the connection
		// In a full implementation, this would handle MCP protocol messages
		select {
		case <-r.Context().Done():
			return
		}
	})

	// Messages endpoint for POST requests
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For now, return a simple response
		// In a full implementation, this would handle MCP protocol messages
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "SSE transport not fully implemented",
			"message": "Use stdio transport for full functionality",
		})
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"server":  "LLM Tools MCP Server",
			"version": "1.0.0",
		})
	})

	// Root endpoint with server info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "LLM Tools MCP Server",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"sse":      "/sse",
				"messages": "/messages",
				"health":   "/health",
			},
			"tools": len(toolRegistry),
			"note":  "SSE transport is basic implementation. Use stdio transport for full MCP functionality.",
		})
	})

	addr := ":" + strconv.Itoa(port)
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("SSE endpoint: http://localhost:%d/sse", port)
	log.Printf("Messages endpoint: http://localhost:%d/messages", port)
	log.Printf("Health check: http://localhost:%d/health", port)
	log.Printf("Note: SSE transport is a basic implementation. Use stdio transport for full MCP functionality.")

	return http.ListenAndServe(addr, mux)
}

// addTool adds a single tool to the MCP server using the generic pattern
func addTool(s *server.MCPServer, toolName string, toolReg ToolRegistry) {
	toolDef := toolReg.GetDefinition()

	tool := mcp.NewTool(toolDef.Name, mcp.WithDescription(toolDef.Description))
	tool.InputSchema.Type = string(toolDef.Parameters.Type)
	tool.InputSchema.Required = toolDef.Parameters.Required
	tool.InputSchema.Properties = toolDef.Parameters.PropertiesToMap()

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		jsonBytes, err := json.Marshal(args)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal arguments: %v", err)), nil
		}

		result, err := toolReg.ExecuteFromJSON(string(jsonBytes))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tool execution failed: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}
