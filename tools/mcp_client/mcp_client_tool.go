package mcp_client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools mcp_client communicates with external MCP servers to use custom tools

Usage: llm-tools mcp_client [OPTIONS]

Options:
  --server-command <cmd>       command to execute the MCP server (required)
  --server-args <args>         arguments to pass to the MCP server (can be used multiple times)
  --tool-name <name>           name of the tool to call (defaults to 'list_tools' to see available tools)
  --tool-arguments <json>      JSON string of arguments to pass to the tool
  --timeout <seconds>          timeout in seconds for MCP server communication (default: 30)
  --explanation <text>         explanation for the operation (defaults to appropriate message for list_tools)

Examples:
  llm-tools mcp_client --server-command "mcp-batch-read-file"
  llm-tools mcp_client --server-command "./bin/mcp-server" --tool-name "list_tools" --explanation "List available tools"
  llm-tools mcp_client --server-command "python" --server-args "my_server.py" --tool-name "my_tool" --tool-arguments '{"param": "value"}' --explanation "Using custom tool"
  llm-tools mcp_client --server-command "node" --server-args "server.js" --tool-name "batch_read_file" --tool-arguments '{"files": [{"target_file": "test.txt", "should_read_entire_file": true}]}' --explanation "Reading files via MCP"
`

func HandleCli(args []string) error {
	var serverCommand string
	var serverArgs []string
	var toolName string
	var toolArgumentsJSON string
	var timeoutStr string
	var explanation string

	args, err := flags.String("--server-command", &serverCommand).
		StringSlice("--server-args", &serverArgs).
		String("--tool-name", &toolName).
		String("--tool-arguments", &toolArgumentsJSON).
		String("--timeout", &timeoutStr).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	if serverCommand == "" {
		return fmt.Errorf("--server-command is required")
	}

	// Default to list_tools if no tool name is provided
	if toolName == "" {
		toolName = "list_tools"
	}

	// Default explanation for list_tools
	if explanation == "" {
		if toolName == "list_tools" {
			explanation = "List available tools from MCP server"
		} else {
			return fmt.Errorf("--explanation is required when using custom tools")
		}
	}

	// Parse timeout
	timeout := 30
	if timeoutStr != "" {
		timeout, err = strconv.Atoi(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout value: %s", timeoutStr)
		}
	}

	// Parse tool arguments
	var toolArguments map[string]interface{}
	if toolArgumentsJSON != "" {
		if err := json.Unmarshal([]byte(toolArgumentsJSON), &toolArguments); err != nil {
			return fmt.Errorf("invalid tool arguments JSON: %w", err)
		}
	}

	req := MCPClientRequest{
		ServerCommand:  serverCommand,
		ServerArgs:     serverArgs,
		ToolName:       toolName,
		ToolArguments:  toolArguments,
		Explanation:    explanation,
		TimeoutSeconds: timeout,
	}

	response, err := MCPClient(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)

	if response.Error != "" {
		fmt.Printf("Error: %s\n", response.Error)
	}

	if len(response.AvailableTools) > 0 {
		fmt.Printf("Available Tools (%d):\n", len(response.AvailableTools))
		for _, tool := range response.AvailableTools {
			fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			if tool.InputSchema != nil {
				schemaJSON, _ := json.MarshalIndent(tool.InputSchema, "    ", "  ")
				fmt.Printf("    Schema: %s\n", string(schemaJSON))
			}
		}
	}

	if response.ToolResult != nil {
		fmt.Printf("Tool Result:\n")
		resultJSON, err := json.MarshalIndent(response.ToolResult, "", "  ")
		if err != nil {
			fmt.Printf("  (failed to format result: %v)\n", err)
		} else {
			fmt.Printf("%s\n", string(resultJSON))
		}
	}

	return nil
}
