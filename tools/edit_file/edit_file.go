package edit_file

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// EditFileRequest represents the input parameters for the edit_file tool
type EditFileRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	TargetFile    string `json:"target_file"`
	OldString     string `json:"old_string"`
	NewString     string `json:"new_string"`
	Explanation   string `json:"explanation"`
}

// EditFileResponse represents the output of the edit_file tool
type EditFileResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	FilePath     string `json:"file_path"`
	ChangesCount int    `json:"changes_count"`
	LinesChanged int    `json:"lines_changed"`
}

// GetToolDefinition returns the JSON schema definition for the edit_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Edit an existing file by replacing occurrences of old_string with new_string. The tool will replace ALL occurrences of the old_string in the file. The file must exist before editing.`,
		Name:        "edit_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to edit. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"old_string": {
					Type:        jsonschema.ParamTypeString,
					Description: "The string to be replaced in the file.",
				},
				"new_string": {
					Type:        jsonschema.ParamTypeString,
					Description: "The string to replace the old_string with.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"workspace_root", "target_file", "old_string", "new_string"},
		},
	}
}

// EditFile executes the edit_file tool with the given parameters
func EditFile(req EditFileRequest) (*EditFileResponse, error) {
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", false)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", req.TargetFile)
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Count occurrences before replacement
	oldCount := strings.Count(originalContent, req.OldString)
	if oldCount == 0 {
		return &EditFileResponse{
			Success:      true,
			Message:      fmt.Sprintf("No occurrences of old_string found in file: %s", req.TargetFile),
			FilePath:     filePath,
			ChangesCount: 0,
			LinesChanged: 0,
		}, nil
	}

	// Replace all occurrences
	newContent := strings.ReplaceAll(originalContent, req.OldString, req.NewString)

	// Count lines changed
	linesChanged := countLinesChanged(originalContent, newContent)

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &EditFileResponse{
		Success:      true,
		Message:      fmt.Sprintf("File edited successfully: %s (%d occurrences replaced)", req.TargetFile, oldCount),
		FilePath:     filePath,
		ChangesCount: oldCount,
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

func ParseJSONRequest(jsonInput string) (EditFileRequest, error) {
	var req EditFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return EditFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the edit_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req EditFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := EditFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
