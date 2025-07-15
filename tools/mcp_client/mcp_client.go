package mcp_client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// MCPClientRequest represents the input parameters for the mcp_client tool
type MCPClientRequest struct {
	ServerCommand  string                 `json:"server_command"`
	ServerArgs     []string               `json:"server_args,omitempty"`
	ToolName       string                 `json:"tool_name"`
	ToolArguments  map[string]interface{} `json:"tool_arguments"`
	Explanation    string                 `json:"explanation"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
}

// MCPClientResponse represents the output of the mcp_client tool
type MCPClientResponse struct {
	Success        bool                   `json:"success"`
	Message        string                 `json:"message"`
	ToolResult     map[string]interface{} `json:"tool_result,omitempty"`
	AvailableTools []MCPTool              `json:"available_tools,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

// MCPTool represents a tool available on the MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// MCPMessage represents a JSON-RPC message
type MCPMessage struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// GetToolDefinition returns the JSON schema definition for the mcp_client tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Communicate with external MCP (Model Context Protocol) servers to use custom tools. This tool allows you to connect to any MCP server and use its exposed tools. The MCP server should be a command that can be executed and communicates via stdin/stdout using JSON-RPC.`,
		Name:        "mcp_client",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"server_command": {
					Type:        jsonschema.ParamTypeString,
					Description: "The command to execute the MCP server (e.g., 'python', 'node', './my-server')",
				},
				"server_args": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Arguments to pass to the MCP server command",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"tool_name": {
					Type:        jsonschema.ParamTypeString,
					Description: "The name of the tool to call on the MCP server. Use 'list_tools' to get available tools.",
				},
				"tool_arguments": {
					Type:        jsonschema.ParamTypeObject,
					Description: "Arguments to pass to the tool",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
				"timeout_seconds": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "Timeout in seconds for the MCP server communication (default: 30)",
				},
			},
			Required: []string{"server_command", "tool_name", "explanation"},
		},
	}
}

// MCPClient executes the mcp_client tool with the given parameters
func MCPClient(req MCPClientRequest) (*MCPClientResponse, error) {
	// Set default timeout
	timeout := 30
	if req.TimeoutSeconds > 0 {
		timeout = req.TimeoutSeconds
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Start the MCP server process
	cmd := exec.CommandContext(ctx, req.ServerCommand, req.ServerArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Ensure cleanup
	defer func() {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Initialize the MCP server
	_, err = initializeMCPServer(stdin, stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP server: %w", err)
	}

	// Handle special case: list_tools
	if req.ToolName == "list_tools" {
		tools, err := listMCPTools(stdin, stdout)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools: %w", err)
		}

		return &MCPClientResponse{
			Success:        true,
			Message:        fmt.Sprintf("Successfully listed %d tools from MCP server", len(tools)),
			AvailableTools: tools,
		}, nil
	}

	// Call the specified tool
	toolResult, err := callMCPTool(stdin, stdout, req.ToolName, req.ToolArguments)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", req.ToolName, err)
	}

	return &MCPClientResponse{
		Success:    true,
		Message:    fmt.Sprintf("Successfully called tool %s", req.ToolName),
		ToolResult: toolResult,
	}, nil
}

// initializeMCPServer sends the initialize message to the MCP server
func initializeMCPServer(stdin io.Writer, stdout io.Reader) (map[string]interface{}, error) {
	initMsg := MCPMessage{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "llm-tools-mcp-client",
				"version": "1.0.0",
			},
		},
	}

	return sendMCPMessage(stdin, stdout, initMsg)
}

// listMCPTools gets the list of available tools from the MCP server
func listMCPTools(stdin io.Writer, stdout io.Reader) ([]MCPTool, error) {
	listMsg := MCPMessage{
		JsonRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	response, err := sendMCPMessage(stdin, stdout, listMsg)
	if err != nil {
		return nil, err
	}

	// Extract tools from response
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format for tools/list")
	}

	toolsData, ok := result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no tools found in response")
	}

	var tools []MCPTool
	for _, toolData := range toolsData {
		toolMap, ok := toolData.(map[string]interface{})
		if !ok {
			continue
		}

		tool := MCPTool{
			Name:        getStringFromMap(toolMap, "name"),
			Description: getStringFromMap(toolMap, "description"),
		}

		if inputSchema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
			tool.InputSchema = inputSchema
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// callMCPTool calls a specific tool on the MCP server
func callMCPTool(stdin io.Writer, stdout io.Reader, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	callMsg := MCPMessage{
		JsonRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	response, err := sendMCPMessage(stdin, stdout, callMsg)
	if err != nil {
		return nil, err
	}

	// Extract result from response
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format for tools/call")
	}

	return result, nil
}

// sendMCPMessage sends a JSON-RPC message to the MCP server and returns the response
func sendMCPMessage(stdin io.Writer, stdout io.Reader, message MCPMessage) (map[string]interface{}, error) {
	// Send the message
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := stdin.Write(msgBytes); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	if _, err := stdin.Write([]byte("\n")); err != nil {
		return nil, fmt.Errorf("failed to write newline: %w", err)
	}

	// Read the response
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read response")
	}

	responseBytes := scanner.Bytes()
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors in the response
	if errorData, ok := response["error"]; ok {
		return nil, fmt.Errorf("MCP server error: %v", errorData)
	}

	return response, nil
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func ParseJSONRequest(jsonInput string) (MCPClientRequest, error) {
	var req MCPClientRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return MCPClientRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the mcp_client tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req MCPClientRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := MCPClient(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
