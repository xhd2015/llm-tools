package read_file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// ReadFileRequest represents the input parameters for the read_file tool
type ReadFileRequest struct {
	WorkspaceRoot              string `json:"workspace_root"`
	TargetFile                 string `json:"target_file"`
	ShouldReadEntireFile       bool   `json:"should_read_entire_file"`
	StartLineOneIndexed        int    `json:"start_line_one_indexed"`
	EndLineOneIndexedInclusive int    `json:"end_line_one_indexed_inclusive"`
	Explanation                string `json:"explanation"`
}

// ReadFileResponse represents the output of the read_file tool
type ReadFileResponse struct {
	Contents   string `json:"contents"`
	TotalLines int    `json:"total_lines"`
	LinesShown string `json:"lines_shown"`
	Outline    string `json:"outline,omitempty"`
}

// GetToolDefinition returns the JSON schema definition for the read_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Read the contents of a file. the output of this tool call will be the 1-indexed file contents from start_line_one_indexed to end_line_one_indexed_inclusive, together with a summary of the lines outside start_line_one_indexed and end_line_one_indexed_inclusive.
Note that this call can view at most 250 lines at a time and 200 lines minimum.

When using this tool to gather information, it's your responsibility to ensure you have the COMPLETE context. Specifically, each time you call this command you should:
1) Assess if the contents you viewed are sufficient to proceed with your task.
2) Take note of where there are lines not shown.
3) If the file contents you have viewed are insufficient, and you suspect they may be in lines not shown, proactively call the tool again to view those lines.
4) When in doubt, call this tool again to gather more information. Remember that partial file views may miss critical dependencies, imports, or functionality.

In some cases, if reading a range of lines is not enough, you may choose to read the entire file.
Reading entire files is often wasteful and slow, especially for large files (i.e. more than a few hundred lines). So you should use this option sparingly.
Reading the entire file is not allowed in most cases. You are only allowed to read the entire file if it has been edited or manually attached to the conversation by the user.`,
		Name: "read_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to read. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
				},
				"should_read_entire_file": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether to read the entire file. Defaults to false.",
				},
				"start_line_one_indexed": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "The one-indexed line number to start reading from (inclusive).",
				},
				"end_line_one_indexed_inclusive": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "The one-indexed line number to end reading at (inclusive).",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"target_file", "should_read_entire_file", "start_line_one_indexed", "end_line_one_indexed_inclusive"},
		},
	}
}

// ReadFile executes the read_file tool with the given parameters
func ReadFile(req ReadFileRequest) (*ReadFileResponse, error) {
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", false)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", req.TargetFile)
	}

	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all lines from the file
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	totalLines := len(lines)

	// Handle different reading modes
	var contents string
	var linesShown string
	var outline string

	if req.ShouldReadEntireFile {
		// Read entire file
		contents = strings.Join(lines, "\n")
		linesShown = fmt.Sprintf("1-%d (entire file)", totalLines)
		outline = generateOutline(lines, req.TargetFile)
	} else {
		// Read specific range
		startLine := req.StartLineOneIndexed
		endLine := req.EndLineOneIndexedInclusive

		// Validate line numbers
		if startLine < 1 {
			startLine = 1
		}
		if endLine > totalLines {
			endLine = totalLines
		}
		if startLine > endLine {
			return nil, fmt.Errorf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)
		}

		// Check line limits (max 250 lines, min 200 lines for partial reads)
		requestedLines := endLine - startLine + 1
		if requestedLines > 250 {
			// Adjust end line to stay within 250 line limit
			endLine = startLine + 249
			if endLine > totalLines {
				endLine = totalLines
			}
		}

		// Apply minimum 200 lines rule for partial reads (unless file is smaller)
		if requestedLines < 200 && totalLines > 200 {
			// Expand range to meet minimum, but respect file boundaries
			additionalLines := 200 - requestedLines

			// Try to expand both directions equally
			expandBefore := additionalLines / 2
			expandAfter := additionalLines - expandBefore

			newStartLine := startLine - expandBefore
			newEndLine := endLine + expandAfter

			// Adjust if we go beyond file boundaries
			if newStartLine < 1 {
				newEndLine += (1 - newStartLine)
				newStartLine = 1
			}
			if newEndLine > totalLines {
				newStartLine -= (newEndLine - totalLines)
				newEndLine = totalLines
				if newStartLine < 1 {
					newStartLine = 1
				}
			}

			startLine = newStartLine
			endLine = newEndLine
		}

		// Extract the requested lines (convert to 0-indexed)
		startIdx := startLine - 1
		endIdx := endLine - 1

		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx >= totalLines {
			endIdx = totalLines - 1
		}

		selectedLines := lines[startIdx : endIdx+1]
		contents = strings.Join(selectedLines, "\n")

		// Generate lines shown description
		if startLine == 1 && endLine == totalLines {
			linesShown = fmt.Sprintf("1-%d (entire file)", totalLines)
		} else {
			linesShown = fmt.Sprintf("%d-%d", startLine, endLine)

			// Add context information about lines not shown
			if startLine > 1 || endLine < totalLines {
				contextInfo := fmt.Sprintf("Requested to read lines %d-%d, but returning lines %d-%d to give more context.",
					req.StartLineOneIndexed, req.EndLineOneIndexedInclusive, startLine, endLine)
				contents = contextInfo + "\n" + contents
			}
		}

		outline = generateOutline(selectedLines, req.TargetFile)
	}

	return &ReadFileResponse{
		Contents:   contents,
		TotalLines: totalLines,
		LinesShown: linesShown,
		Outline:    outline,
	}, nil
}

// generateOutline creates a brief outline of the file contents
func generateOutline(lines []string, filename string) string {
	if len(lines) == 0 {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filename))

	// Generate outline based on file type
	switch ext {
	case ".go":
		return generateGoOutline(lines)
	case ".js", ".ts", ".jsx", ".tsx":
		return generateJSOutline(lines)
	case ".py":
		return generatePyOutline(lines)
	case ".java":
		return generateJavaOutline(lines)
	case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp":
		return generateCppOutline(lines)
	default:
		return generateGenericOutline(lines)
	}
}

// generateGoOutline creates an outline for Go files
func generateGoOutline(lines []string) string {
	var outline []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") ||
			strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "var ") ||
			strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "package ") {
			outline = append(outline, trimmed)
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

// generateJSOutline creates an outline for JavaScript/TypeScript files
func generateJSOutline(lines []string) string {
	var outline []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "function ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "export ") ||
			strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "let ") ||
			strings.HasPrefix(trimmed, "var ") {
			outline = append(outline, trimmed)
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

// generatePyOutline creates an outline for Python files
func generatePyOutline(lines []string) string {
	var outline []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "def ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "from ") {
			outline = append(outline, trimmed)
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

// generateJavaOutline creates an outline for Java files
func generateJavaOutline(lines []string) string {
	var outline []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "public class ") ||
			strings.Contains(trimmed, "private class ") ||
			strings.Contains(trimmed, "public interface ") ||
			strings.Contains(trimmed, "public ") && strings.Contains(trimmed, "(") ||
			strings.Contains(trimmed, "private ") && strings.Contains(trimmed, "(") {
			outline = append(outline, trimmed)
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

// generateCppOutline creates an outline for C/C++ files
func generateCppOutline(lines []string) string {
	var outline []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#include ") ||
			strings.HasPrefix(trimmed, "#define ") ||
			strings.Contains(trimmed, "class ") ||
			strings.Contains(trimmed, "struct ") ||
			(strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") && !strings.HasPrefix(trimmed, "//")) {
			outline = append(outline, trimmed)
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

// generateGenericOutline creates a generic outline for other file types
func generateGenericOutline(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// For generic files, just return the first few non-empty lines
	var outline []string
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && count < 3 {
			outline = append(outline, trimmed)
			count++
		}
	}

	if len(outline) == 0 {
		return ""
	}

	return strings.Join(outline, "; ")
}

func ParseJSONRequest(jsonInput string) (ReadFileRequest, error) {
	var req ReadFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return ReadFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the read_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req ReadFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := ReadFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
