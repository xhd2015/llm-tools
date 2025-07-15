package rename_file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// RenameFileRequest represents the input parameters for the rename_file tool
type RenameFileRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	SourceFile    string `json:"source_file"`
	TargetFile    string `json:"target_file"`
	Explanation   string `json:"explanation"`
}

// RenameFileResponse represents the output of the rename_file tool
type RenameFileResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	SourceFilePath string `json:"source_file_path"`
	TargetFilePath string `json:"target_file_path"`
}

// GetToolDefinition returns the JSON schema definition for the rename_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Rename or move a file from one location to another. The tool will create any necessary parent directories for the target file if they don't exist. If the target file already exists, it will be overwritten.`,
		Name:        "rename_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"source_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to rename/move. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The new path/name for the file. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"workspace_root", "source_file", "target_file"},
		},
	}
}

// RenameFile executes the rename_file tool with the given parameters
func RenameFile(req RenameFileRequest) (*RenameFileResponse, error) {
	sourcePath, err := dirs.GetPath(req.WorkspaceRoot, req.SourceFile, "source_file", false)
	if err != nil {
		return nil, err
	}

	targetPath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", true)
	if err != nil {
		return nil, err
	}

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %s", req.SourceFile)
	}

	// Create parent directories for target if they don't exist
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Rename/move the file
	err = os.Rename(sourcePath, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to rename file: %w", err)
	}

	return &RenameFileResponse{
		Success:        true,
		Message:        fmt.Sprintf("File renamed successfully: %s -> %s", req.SourceFile, req.TargetFile),
		SourceFilePath: sourcePath,
		TargetFilePath: targetPath,
	}, nil
}

func ParseJSONRequest(jsonInput string) (RenameFileRequest, error) {
	var req RenameFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return RenameFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the rename_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req RenameFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := RenameFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
