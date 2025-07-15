package create_file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// CreateFileRequest represents the input parameters for the create_file tool
type CreateFileRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	TargetFile    string `json:"target_file"`
	Content       string `json:"content"`
	Explanation   string `json:"explanation"`
}

// CreateFileResponse represents the output of the create_file tool
type CreateFileResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	FilePath     string `json:"file_path"`
	BytesWritten int    `json:"bytes_written"`
}

// GetToolDefinition returns the JSON schema definition for the create_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Create a new file with the specified content. If the file already exists, it will be overwritten. The tool will create any necessary parent directories if they don't exist.`,
		Name:        "create_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to create. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"content": {
					Type:        jsonschema.ParamTypeString,
					Description: "The content to write to the file.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"workspace_root", "target_file", "content"},
		},
	}
}

// CreateFile executes the create_file tool with the given parameters
func CreateFile(req CreateFileRequest) (*CreateFileResponse, error) {
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", true)
	if err != nil {
		return nil, err
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write content to file
	err = os.WriteFile(filePath, []byte(req.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &CreateFileResponse{
		Success:      true,
		Message:      fmt.Sprintf("File created successfully: %s", req.TargetFile),
		FilePath:     filePath,
		BytesWritten: len(req.Content),
	}, nil
}

func ParseJSONRequest(jsonInput string) (CreateFileRequest, error) {
	var req CreateFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return CreateFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the create_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req CreateFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := CreateFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
