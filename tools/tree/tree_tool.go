package tree

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

const DEFAULT_MAX_DEPTH = 10
const DEFAULT_MAX_ENTRIES_PER_DIR = 40

// TreeRequest represents the input parameters for the tree tool
type TreeRequest struct {
	WorkspaceRoot         string   `json:"workspace_root"`
	RelativeWorkspacePath string   `json:"relative_workspace_path"`
	IncludePatterns       []string `json:"include_patterns,omitempty"`
	ExcludePatterns       []string `json:"exclude_patterns,omitempty"`
	IncludeFiles          bool     `json:"include_files,omitempty"`
	Depth                 int      `json:"depth,omitempty"`
	MaxEntriesPerDir      int      `json:"max_entries_per_dir,omitempty"`
	ExpandDirs            []string `json:"expand_dirs,omitempty"`
	Explanation           string   `json:"explanation"`
}

// TreeResponse represents the output of the tree tool
type TreeResponse struct {
	Tree string `json:"tree"`
}

// GetToolDefinition returns the JSON schema definition for the tree tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: "Display directory tree structure. Useful for understanding project organization and file hierarchy. Supports filtering, collapsing repeated patterns, and various display options.",
		Name:        "tree",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.WORKSPACE_ROOT,
				},
				"relative_workspace_path": {
					Type:        jsonschema.ParamTypeString,
					Description: "Path to display tree for, relative to the workspace root.",
				},
				"include_patterns": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Regex patterns to include file/directory names (whitelist).",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"exclude_patterns": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Regex patterns to exclude file/directory names (blacklist).",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"include_files": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Include files in the tree display.",
				},
				"depth": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "Maximum depth of the directory tree to traverse.",
					Default:     DEFAULT_MAX_DEPTH,
				},
				"max_entries_per_dir": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "Maximum number of entries to display in each directory.",
					Default:     DEFAULT_MAX_ENTRIES_PER_DIR,
				},
				"expand_dirs": {
					Type:        jsonschema.ParamTypeArray,
					Description: "List of directory paths to expand in the tree display.",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.EXPLANATION,
				},
			},
			Required: []string{"relative_workspace_path"},
		},
	}
}

// ExecuteTree executes the tree tool with the given parameters
func ExecuteTree(req TreeRequest) (*TreeResponse, error) {
	targetDir, err := dirs.GetPath(req.WorkspaceRoot, req.RelativeWorkspacePath, "relative_workspace_path", true)
	if err != nil {
		return nil, err
	}

	// Check if the directory exists
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", req.RelativeWorkspacePath)
		}
		return nil, fmt.Errorf("error accessing directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", req.RelativeWorkspacePath)
	}

	// Set defaults if not specified
	depth := req.Depth
	if depth == 0 {
		depth = DEFAULT_MAX_DEPTH
	}

	maxEntriesPerDir := req.MaxEntriesPerDir
	if maxEntriesPerDir == 0 {
		maxEntriesPerDir = DEFAULT_MAX_ENTRIES_PER_DIR
	}

	// Build tree options
	opts := TreeOptions{
		IncludePatterns:  req.IncludePatterns,
		ExcludePatterns:  req.ExcludePatterns,
		DirectoriesOnly:  !req.IncludeFiles,
		Depth:            depth,
		MaxEntriesPerDir: maxEntriesPerDir,
		ExpandDirs:       req.ExpandDirs,
	}

	// Generate tree
	treeOutput, err := Tree(targetDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tree: %w", err)
	}

	// Clean up the path for display
	displayPath := req.RelativeWorkspacePath
	if displayPath == "" || displayPath == "." {
		displayPath = "."
	}

	return &TreeResponse{
		Tree: treeOutput,
	}, nil
}

// ParseJSONRequest parses JSON input into TreeRequest
func ParseJSONRequest(jsonInput string) (TreeRequest, error) {
	var req TreeRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return TreeRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the tree tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req TreeRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := ExecuteTree(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
