package list_dir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// ListDirRequest represents the input parameters for the list_dir tool
type ListDirRequest struct {
	WorkspaceRoot         string `json:"workspace_root"`
	RelativeWorkspacePath string `json:"relative_workspace_path"`
	Explanation           string `json:"explanation"`
}

// ListDirResponse represents the output of the list_dir tool
type ListDirResponse struct {
	Contents []string `json:"contents"`
	Path     string   `json:"path"`
	Count    int      `json:"count"`
}

// GetToolDefinition returns the JSON schema definition for the list_dir tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: "List the contents of a directory. The quick tool to use for discovery, before using more targeted tools like semantic search or file reading. Useful to try to understand the file structure before diving deeper into specific files. Can be used to explore the codebase.",
		Name:        "list_dir",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"relative_workspace_path": {
					Type:        jsonschema.ParamTypeString,
					Description: "Path to list contents of, relative to the workspace root.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"workspace_root", "relative_workspace_path"},
		},
	}
}

// ListDir executes the list_dir tool with the given parameters
func ListDir(req ListDirRequest) (*ListDirResponse, error) {
	// Validate input parameters
	if req.RelativeWorkspacePath == "" {
		return nil, fmt.Errorf("required relative_workspace_path")
	}

	opDir := req.RelativeWorkspacePath
	if !filepath.IsAbs(req.RelativeWorkspacePath) {
		if req.WorkspaceRoot == "" {
			return nil, fmt.Errorf("requires workspace_root when relative_workspace_path is a relative path")
		}
		if !filepath.IsAbs(req.WorkspaceRoot) {
			return nil, fmt.Errorf("workspace_root must be absolute path")
		}

		opDir = filepath.Join(req.WorkspaceRoot, req.RelativeWorkspacePath)
	}

	// Check if it's actually a directory
	info, err := os.Stat(opDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not exist: %s", req.RelativeWorkspacePath)
		}
		return nil, fmt.Errorf("stat path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", req.RelativeWorkspacePath)
	}

	// Read directory contents
	entries, err := os.ReadDir(opDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Process entries
	var contents []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		contents = append(contents, name)
	}

	// Sort contents for consistent output
	sort.Strings(contents)

	// Clean up the path for display
	displayPath := req.RelativeWorkspacePath
	if displayPath == "" || displayPath == "." {
		displayPath = "."
	}

	return &ListDirResponse{
		Contents: contents,
		Path:     displayPath,
		Count:    len(contents),
	}, nil
}

func ParseJSONRequest(jsonInput string) (ListDirRequest, error) {
	var req ListDirRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return ListDirRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the list_dir tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req ListDirRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := ListDir(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
