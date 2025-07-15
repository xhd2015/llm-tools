package search_replace

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// SearchReplaceRequest represents the input parameters for the search_replace tool
type SearchReplaceRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	FilePath      string `json:"file_path"`
	OldString     string `json:"old_string"`
	NewString     string `json:"new_string"`
}

// SearchReplaceResponse represents the output of the search_replace tool
type SearchReplaceResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	FilePath     string `json:"file_path"`
	ChangesCount int    `json:"changes_count"`
	LinesChanged int    `json:"lines_changed"`
}

// GetToolDefinition returns the JSON schema definition for the search_replace tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Use this tool to propose a search and replace operation on an existing file.

The tool will replace ONE occurrence of old_string with new_string in the specified file.

CRITICAL REQUIREMENTS FOR USING THIS TOOL:

1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:
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
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.WORKSPACE_ROOT,
				},
				"file_path": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path to the file you want to search and replace in." + defs.FILE_PATH,
				},
				"old_string": {
					Type:        jsonschema.ParamTypeString,
					Description: "The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)",
				},
				"new_string": {
					Type:        jsonschema.ParamTypeString,
					Description: "The edited text to replace the old_string (must be different from the old_string)",
				},
			},
			Required: []string{"file_path", "old_string", "new_string"},
		},
	}
}

// SearchReplace executes the search_replace tool with the given parameters
func SearchReplace(req SearchReplaceRequest) (*SearchReplaceResponse, error) {
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.FilePath, "file_path", false)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", req.FilePath)
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Validate that old_string and new_string are different
	if req.OldString == req.NewString {
		return nil, fmt.Errorf("old_string and new_string must be different")
	}

	// Count occurrences of old_string
	occurrences := strings.Count(originalContent, req.OldString)
	if occurrences == 0 {
		return nil, fmt.Errorf("old_string not found in file: %s", req.FilePath)
	}

	if occurrences > 1 {
		return nil, fmt.Errorf("old_string appears %d times in file. It must be unique within the file. Please include more context to make it unique", occurrences)
	}

	// Replace the single occurrence
	newContent := strings.Replace(originalContent, req.OldString, req.NewString, 1)

	// Count lines changed
	linesChanged := countLinesChanged(originalContent, newContent)

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &SearchReplaceResponse{
		Success:      true,
		Message:      fmt.Sprintf("File edited successfully: %s (1 occurrence replaced)", req.FilePath),
		FilePath:     filePath,
		ChangesCount: 1,
		LinesChanged: linesChanged,
	}, nil
}

// countLinesChanged counts the number of lines that were modified
func countLinesChanged(original, modified string) int {
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	changed := 0
	maxLines := len(originalLines)
	if len(modifiedLines) > maxLines {
		maxLines = len(modifiedLines)
	}

	for i := 0; i < maxLines; i++ {
		var origLine, modLine string
		if i < len(originalLines) {
			origLine = originalLines[i]
		}
		if i < len(modifiedLines) {
			modLine = modifiedLines[i]
		}
		if origLine != modLine {
			changed++
		}
	}

	return changed
}

func ParseJSONRequest(jsonInput string) (SearchReplaceRequest, error) {
	var req SearchReplaceRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return SearchReplaceRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the search_replace tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req SearchReplaceRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := SearchReplace(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
