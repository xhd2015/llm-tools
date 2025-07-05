package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhd2015/llm-tools/tools/batch_read_file"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"LLM Tools Batch Read File Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Add the batch_read_file tool using the exact schema from batch_read_file.GetToolDefinition()
	batchReadFileTool := mcp.NewTool("batch_read_file",
		mcp.WithDescription(`Read the contents of multiple files in a single batch operation. This tool improves efficiency by reading multiple files at once instead of making separate read_file calls.
Each file in the batch can have individual line range settings, and the tool respects the same line limits as read_file (max 250 lines, min 200 lines for partial reads).

Key features:
- Batch processing of multiple files
- Individual line range control per file
- Global and per-file line limits
- Error handling with continue-on-error option
- Optional outline generation
- Structured response with success/error counts

This tool is particularly useful when you need to read multiple related files (e.g., examining imports, comparing implementations, or gathering context from multiple source files).`),
	)

	batchReadFileTool.InputSchema.Properties = map[string]any{
		"files": map[string]any{
			"type":        "array",
			"description": "Array of file read requests, each with its own target file and line range settings",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_file": map[string]any{
						"type":        "string",
						"description": "The path of the file to read. You can use either a relative path in the workspace or an absolute path.",
					},
					"should_read_entire_file": map[string]any{
						"type":        "string", // Note: should be boolean but keeping as string for compatibility
						"description": "Whether to read the entire file. When true, start_line_one_indexed and end_line_one_indexed_inclusive are ignored. When false, line range parameters are required. Defaults to false.",
					},
					"start_line_one_indexed": map[string]any{
						"type":        "number",
						"description": "The one-indexed line number to start reading from (inclusive). Required when should_read_entire_file is false. Ignored when should_read_entire_file is true.",
					},
					"end_line_one_indexed_inclusive": map[string]any{
						"type":        "number",
						"description": "The one-indexed line number to end reading at (inclusive). Required when should_read_entire_file is false. Ignored when should_read_entire_file is true.",
					},
					"max_lines": map[string]any{
						"type":        "number",
						"description": "Optional per-file maximum lines limit. Overrides global_max_lines for this file.",
					},
				},
				"required": []string{"target_file"},
			},
		},
		"global_max_lines": map[string]any{
			"type":        "number",
			"description": "Global maximum lines per file (default: 250). Can be overridden per file.",
		},
		"global_min_lines": map[string]any{
			"type":        "number",
			"description": "Global minimum lines per file for partial reads (default: 200). Applied when expanding ranges.",
		},
		"continue_on_error": map[string]any{
			"type":        "string", // Note: should be boolean but keeping as string for compatibility
			"description": "Whether to continue processing other files if one fails (default: true).",
		},
		"include_outline": map[string]any{
			"type":        "string", // Note: should be boolean but keeping as string for compatibility
			"description": "Whether to include outline in responses (default: true).",
		},
		"explanation": map[string]any{
			"type":        "string",
			"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
		},
	}
	batchReadFileTool.InputSchema.Required = []string{"files"}

	// Add the tool handler
	s.AddTool(batchReadFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get the explanation
		explanation, err := request.RequireString("explanation")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing required explanation: %v", err)), nil
		}

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
			Files:       fileRequests,
			Explanation: explanation,
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
