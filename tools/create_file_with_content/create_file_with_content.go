package create_file_with_content

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// CreateFileWithContentRequest represents the input parameters for the create_file_with_content tool
type CreateFileWithContentRequest struct {
	WorkspaceRoot     string `json:"workspace_root"`
	TargetFile        string `json:"target_file"`
	NonEmptyContent   string `json:"non_empty_content"`
	DangerousOverride bool   `json:"dangerous_override"`
	Explanation       string `json:"explanation"`
}

// CreateFileWithContentResponse represents the output of the create_file_with_content tool
type CreateFileWithContentResponse struct {
	Success      bool `json:"success"`
	BytesWritten int  `json:"bytes_written"`
	Overwritten  bool `json:"overwritten"`
}

// GetToolDefinition returns the JSON schema definition for the create_file_with_content tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Create a file with content. By default, this tool will fail if the file already exists to prevent accidental overwrites. Use the dangerous_override flag to allow overwriting existing files. The tool will create any necessary parent directories if they don't exist.`,
		Name:        "create_file_with_content",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.WORKSPACE_ROOT,
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to create. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"non_empty_content": {
					Type:        jsonschema.ParamTypeString,
					Description: "The content to write to the file. Must be non-empty.",
				},
				"dangerous_override": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "If true, allows overwriting existing files. If false (default), the operation will fail if the file already exists.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.EXPLANATION,
				},
			},
			Required: []string{"target_file", "non_empty_content"},
		},
	}
}

// CreateFileWithContent executes the create_file_with_content tool with the given parameters
func CreateFileWithContent(req CreateFileWithContentRequest) (*CreateFileWithContentResponse, error) {
	if req.NonEmptyContent == "" {
		return nil, fmt.Errorf("requires non-empty content")
	}

	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", true)
	if err != nil {
		return nil, err
	}

	// Check if file exists and handle override logic
	var overwritten bool
	if _, err := os.Stat(filePath); err == nil {
		// File exists
		if !req.DangerousOverride {
			return nil, fmt.Errorf("file already exists: %s. Use dangerous_override=true to overwrite", req.TargetFile)
		}
		overwritten = true
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write content to file
	err = os.WriteFile(filePath, []byte(req.NonEmptyContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &CreateFileWithContentResponse{
		Success:      true,
		BytesWritten: len(req.NonEmptyContent),
		Overwritten:  overwritten,
	}, nil
}

func ParseJSONRequest(jsonInput string) (CreateFileWithContentRequest, error) {
	var req CreateFileWithContentRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return CreateFileWithContentRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the create_file_with_content tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req CreateFileWithContentRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := CreateFileWithContent(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
