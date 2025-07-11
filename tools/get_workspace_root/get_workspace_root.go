package get_workspace_root

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// GetWorkspaceRootRequest represents the input parameters for the batch_read_file tool
type GetWorkspaceRootRequest struct {
	Explanation string `json:"explanation"`
}

// GetWorkspaceRootResponse represents the output of the batch_read_file tool
type GetWorkspaceRootResponse struct {
	WorkspaceRoot string `json:"workspace_root"`
}

// GetToolDefinition returns the JSON schema definition for the batch_read_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Get the workspace root directory.`,
		Name:        "get_workspace_root",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "brief explanation",
				},
			},
		},
	}
}

// GetWorkspaceRoot executes the get_workspace_root tool with the given parameters
func GetWorkspaceRoot(req GetWorkspaceRootRequest, defaultWorkspaceRoot string) (*GetWorkspaceRootResponse, error) {
	// Set default values
	if defaultWorkspaceRoot != "" {
		absDir, _ := filepath.Abs(defaultWorkspaceRoot)
		if absDir != "" {
			return &GetWorkspaceRootResponse{
				WorkspaceRoot: absDir,
			}, nil
		}
		return &GetWorkspaceRootResponse{
			WorkspaceRoot: defaultWorkspaceRoot,
		}, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	return &GetWorkspaceRootResponse{
		WorkspaceRoot: wd,
	}, nil
}

func ParseJSONRequest(jsonInput string) (GetWorkspaceRootRequest, error) {
	var req GetWorkspaceRootRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return GetWorkspaceRootRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the batch_read_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req GetWorkspaceRootRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := GetWorkspaceRoot(req, "")
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
