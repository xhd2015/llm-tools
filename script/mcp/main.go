package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/llm-tools/tools/batch_read_file"
)

const help = `
mcp-batch-read-file

Usage: mcp-batch-read-file

Options:
  -h, --help                       show help
  -v,--verbose                     show verbose info  

Examples:
  mcp-batch-read-file help         show help message

  # inspect mcp
  echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | mcp-batch-read-file

Local development:
  go build -o $GOPATH/bin/mcp-batch-read-file ./
`

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Create a new MCP server
	s := server.NewMCPServer(
		"LLM Tools Batch Read File Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	toolDef := batch_read_file.GetToolDefinition()

	// Add the batch_read_file tool using the exact schema from batch_read_file.GetToolDefinition()
	batchReadFileTool := mcp.NewTool(toolDef.Name,
		mcp.WithDescription(toolDef.Description),
	)

	batchReadFileTool.InputSchema.Type = string(toolDef.Parameters.Type)
	batchReadFileTool.InputSchema.Required = toolDef.Parameters.Required
	batchReadFileTool.InputSchema.Properties = toolDef.Parameters.PropertiesToMap()

	// Add the tool handler
	s.AddTool(batchReadFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get the explanation
		explanation := request.GetString("explanation", "")

		workspaceRoot := request.GetString("workspace_root", "")

		// Get all arguments
		args := request.GetArguments()

		// Check for files parameter
		filesArg, ok := args["files"]
		if !ok {
			return mcp.NewToolResultError(`Missing 'files' parameter. Expected structure:
{
  "explanation": "Why you're reading these files",
  "files": [
    {
      "target_file": "path/to/file.go",
      "should_read_entire_file": true
    }
  ],
  "continue_on_error": true,
  "include_outline": true
}`), nil
		}

		// Convert files argument to proper structure
		var fileRequests []batch_read_file.FileReadRequest
		if err := convertToFileRequests(filesArg, &fileRequests); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid files parameter: %v", err)), nil
		}

		// Build the BatchReadFileRequest
		req := batch_read_file.BatchReadFileRequest{
			WorkspaceRoot: workspaceRoot,
			Files:         fileRequests,
			Explanation:   explanation,
		}

		// Set optional parameters with defaults
		req.ContinueOnError = true
		req.IncludeOutline = true

		// Override with provided values if they exist
		if val, ok := args["global_max_lines"]; ok {
			if maxLines, ok := val.(float64); ok {
				req.GlobalMaxLines = int(maxLines)
			}
		}
		if val, ok := args["global_min_lines"]; ok {
			if minLines, ok := val.(float64); ok {
				req.GlobalMinLines = int(minLines)
			}
		}
		if val, ok := args["continue_on_error"]; ok {
			if continueOnError, ok := val.(bool); ok {
				req.ContinueOnError = continueOnError
			}
		}
		if val, ok := args["include_outline"]; ok {
			if includeOutline, ok := val.(bool); ok {
				req.IncludeOutline = includeOutline
			}
		}

		// Call the type-safe BatchReadFile function
		response, err := batch_read_file.BatchReadFile(req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Batch read file failed: %v", err)), nil
		}

		// Convert response to JSON for output
		jsonOutput, err := json.Marshal(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		// Return the result as text
		return mcp.NewToolResultText(string(jsonOutput)), nil
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handle(args []string) error {
	args, err := flags.New().Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		cmd := args[0]
		args = args[1:]
		if cmd == "--help" || cmd == "help" {
			fmt.Print(strings.TrimPrefix(help, "\n"))
			os.Exit(0)
			return nil
		}
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra arguments: %s", strings.Join(args, ","))
		}
	}
	return nil
}

// convertToFileRequests converts the files argument to []FileReadRequest
func convertToFileRequests(filesArg interface{}, fileRequests *[]batch_read_file.FileReadRequest) error {
	// Convert to JSON and back to ensure proper type conversion
	jsonBytes, err := json.Marshal(filesArg)
	if err != nil {
		return fmt.Errorf("failed to marshal files argument: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, fileRequests); err != nil {
		return fmt.Errorf("failed to unmarshal files argument: %w", err)
	}

	return nil
}
