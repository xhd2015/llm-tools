package search_replace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// SearchReplaceRequest represents the input parameters for the search_replace tool
type SearchReplaceRequest struct {
	File string `json:"file"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

// SearchReplaceResponse represents the output of the search_replace tool
type SearchReplaceResponse struct {
	Message string `json:"message"`
}

// GetToolDefinition returns the JSON schema definition for the search_replace tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Use this tool to propose a search and replace operation on an existing file.

The tool will replace ONE occurrence of old with new in the specified file.

CRITICAL REQUIREMENTS FOR USING THIS TOOL:

1. UNIQUENESS: The old MUST uniquely identify the specific instance you want to change. This means:
   - Include AT LEAST 3-5 lines of context BEFORE the change point
   - Include AT LEAST 3-5 lines of context AFTER the change point
   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file

2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:
   - Make separate calls to this tool for each instance
   - Each call must uniquely identify its specific instance using extensive context

3. VERIFICATION: Before using this tool:
   - If multiple instances exist, gather enough context to uniquely identify each one
   - Plan separate tool calls for each instance
`,
		Name: "search_replace",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"file": {
					Type: jsonschema.ParamTypeString,
				},
				"old": {
					Type:        jsonschema.ParamTypeString,
					Description: "The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)",
				},
				"new": {
					Type:        jsonschema.ParamTypeString,
					Description: "The edited text to replace the old (must be different from the old). If empty, the old will be deleted.",
				},
			},
			Required: []string{"file", "old", "new"},
		},
	}
}

// SearchReplace executes the search_replace tool with the given parameters
func SearchReplace(req SearchReplaceRequest, workspaceRoot string) (*SearchReplaceResponse, error) {
	if req.File == "" {
		return nil, fmt.Errorf("requires file")
	}
	if req.Old == "" {
		return nil, fmt.Errorf("requires old")
	}
	// Validate that old and new must be different
	if req.Old == req.New {
		return nil, fmt.Errorf("old and new must be different")
	}
	file := req.File
	actualFilePath := file
	if !filepath.IsAbs(file) && workspaceRoot != "" {
		actualFilePath = filepath.Join(workspaceRoot, file)
	}

	// Read the file
	content, err := os.ReadFile(actualFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	idx := strings.Index(originalContent, req.Old)
	if idx < 0 {
		return nil, fmt.Errorf("old not found in file: %s", file)
	}

	subSequentOccurrences := strings.Count(originalContent[idx+len(req.Old):], req.Old)
	if subSequentOccurrences > 0 {
		return nil, fmt.Errorf("old appears %d times in file. It must be unique within the file. Please include more context to make it unique", subSequentOccurrences+1)
	}

	// Replace the single occurrence
	newContent := originalContent[:idx] + req.New + originalContent[idx+len(req.Old):]

	// Write the modified content back to the file
	err = os.WriteFile(actualFilePath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	suffix := "replaced"
	if req.New == "" {
		suffix = "deleted"
	}

	return &SearchReplaceResponse{
		Message: fmt.Sprintf("edited %s: 1 occurrence %s", file, suffix),
	}, nil
}

func ParseJSONRequest(jsonInput string) (SearchReplaceRequest, error) {
	var req SearchReplaceRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return SearchReplaceRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the search_replace tool from JSON input
func ExecuteFromJSON(jsonInput string, workspaceRoot string) (string, error) {
	var req SearchReplaceRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := SearchReplace(req, workspaceRoot)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
