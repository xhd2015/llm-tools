package batch_read_file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// FileReadRequest represents a single file read request within the batch
type FileReadRequest struct {
	TargetFile                 string `json:"target_file"`
	ShouldReadEntireFile       bool   `json:"should_read_entire_file"`
	StartLineOneIndexed        int    `json:"start_line_one_indexed"`
	EndLineOneIndexedInclusive int    `json:"end_line_one_indexed_inclusive"`
	MaxLines                   int    `json:"max_lines,omitempty"` // Optional per-file max lines limit
}

// FileReadResponse represents the response for a single file read
type FileReadResponse struct {
	TargetFile string `json:"target_file"`
	Contents   string `json:"contents"`
	TotalLines int    `json:"total_lines"`
	LinesShown string `json:"lines_shown"`
	Outline    string `json:"outline,omitempty"`
	Error      string `json:"error,omitempty"`
}

// BatchReadFileRequest represents the input parameters for the batch_read_file tool
type BatchReadFileRequest struct {
	WorkspaceRoot   string            `json:"workspace_root"`
	Files           []FileReadRequest `json:"files"`
	GlobalMaxLines  int               `json:"global_max_lines,omitempty"` // Global max lines per file (default: 250)
	GlobalMinLines  int               `json:"global_min_lines,omitempty"` // Global min lines per file (default: 200)
	ContinueOnError bool              `json:"continue_on_error"`          // Whether to continue processing other files if one fails
	IncludeOutline  bool              `json:"include_outline"`            // Whether to include outline in responses
	Explanation     string            `json:"explanation"`
}

// BatchReadFileResponse represents the output of the batch_read_file tool
type BatchReadFileResponse struct {
	Files        []FileReadResponse `json:"files"`
	TotalFiles   int                `json:"total_files"`
	SuccessCount int                `json:"success_count"`
	ErrorCount   int                `json:"error_count"`
}

// GetToolDefinition returns the JSON schema definition for the batch_read_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Read the contents of multiple files in a single batch operation. This tool improves efficiency by reading multiple files at once instead of making separate read_file calls.
Each file in the batch can have individual line range settings, and the tool respects the same line limits as read_file (max 250 lines, min 200 lines for partial reads).

Key features:
- Batch processing of multiple files
- Individual line range control per file
- Global and per-file line limits
- Error handling with continue-on-error option
- Optional outline generation
- Structured response with success/error counts

This tool is particularly useful when you need to read multiple related files (e.g., examining imports, comparing implementations, or gathering context from multiple source files).`,
		Name: "batch_read_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"files": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Array of file read requests, each with its own target file and line range settings",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeObject,
						Properties: map[string]*jsonschema.JsonSchema{
							"target_file": {
								Type:        jsonschema.ParamTypeString,
								Description: "The path of the file to read. You can use either a relative path in the workspace or an absolute path.",
							},
							"should_read_entire_file": {
								Type:        jsonschema.ParamTypeBoolean,
								Description: "Whether to read the entire file. When true, start_line_one_indexed and end_line_one_indexed_inclusive are ignored. When false, line range parameters are required. Defaults to false.",
							},
							"start_line_one_indexed": {
								Type:        jsonschema.ParamTypeNumber,
								Description: "The one-indexed line number to start reading from (inclusive). Required when should_read_entire_file is false. Ignored when should_read_entire_file is true.",
							},
							"end_line_one_indexed_inclusive": {
								Type:        jsonschema.ParamTypeNumber,
								Description: "The one-indexed line number to end reading at (inclusive). Required when should_read_entire_file is false. Ignored when should_read_entire_file is true.",
							},
							"max_lines": {
								Type:        jsonschema.ParamTypeNumber,
								Description: "Optional per-file maximum lines limit. Overrides global_max_lines for this file.",
							},
						},
						Required: []string{"target_file"},
					},
				},
				"global_max_lines": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "Global maximum lines per file (default: 250). Can be overridden per file.",
				},
				"global_min_lines": {
					Type:        jsonschema.ParamTypeNumber,
					Description: "Global minimum lines per file for partial reads (default: 200). Applied when expanding ranges.",
				},
				"continue_on_error": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether to continue processing other files if one fails (default: true).",
				},
				"include_outline": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether to include outline in responses (default: true).",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"workspace_root", "files"},
		},
	}
}

// BatchReadFile executes the batch_read_file tool with the given parameters
func BatchReadFile(req BatchReadFileRequest) (*BatchReadFileResponse, error) {
	// Set default values
	if req.GlobalMaxLines == 0 {
		req.GlobalMaxLines = 250
	}
	if req.GlobalMinLines == 0 {
		req.GlobalMinLines = 200
	}

	// Validate input
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("at least one file must be specified")
	}

	var responses []FileReadResponse
	successCount := 0
	errorCount := 0

	for _, fileReq := range req.Files {
		response := processFileRequest(req.WorkspaceRoot, fileReq, req.GlobalMaxLines, req.GlobalMinLines, req.IncludeOutline)
		responses = append(responses, response)

		if response.Error != "" {
			errorCount++
			if !req.ContinueOnError {
				break
			}
		} else {
			successCount++
		}
	}

	return &BatchReadFileResponse{
		Files:        responses,
		TotalFiles:   len(req.Files),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	}, nil
}

// processFileRequest processes a single file read request
func processFileRequest(workspaceRoot string, fileReq FileReadRequest, globalMaxLines, globalMinLines int, includeOutline bool) FileReadResponse {
	response := FileReadResponse{
		TargetFile: fileReq.TargetFile,
	}

	// Validate target file
	if fileReq.TargetFile == "" {
		response.Error = "target_file is required"
		return response
	}

	// Validate parameters based on should_read_entire_file
	if !fileReq.ShouldReadEntireFile {
		// When not reading entire file, line range parameters are required
		if fileReq.StartLineOneIndexed == 0 && fileReq.EndLineOneIndexedInclusive == 0 {
			response.Error = "start_line_one_indexed and end_line_one_indexed_inclusive are required when should_read_entire_file is false"
			return response
		}
		if fileReq.StartLineOneIndexed <= 0 {
			response.Error = "start_line_one_indexed must be greater than 0"
			return response
		}
		if fileReq.EndLineOneIndexedInclusive <= 0 {
			response.Error = "end_line_one_indexed_inclusive must be greater than 0"
			return response
		}
	}

	filePath := fileReq.TargetFile
	if !filepath.IsAbs(filePath) {
		if workspaceRoot == "" {
			response.Error = "workspace_root is required when target_file is a relative path"
			return response
		}
		filePath = filepath.Join(workspaceRoot, filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error = fmt.Sprintf("file does not exist: %s", filePath)
		return response
	}

	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		response.Error = fmt.Sprintf("failed to open file: %v", err)
		return response
	}
	defer file.Close()

	// Read all lines from the file
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		response.Error = fmt.Sprintf("error reading file: %v", err)
		return response
	}

	totalLines := len(lines)
	response.TotalLines = totalLines

	// Determine max lines limit for this file
	maxLines := globalMaxLines
	if fileReq.MaxLines > 0 {
		maxLines = fileReq.MaxLines
	}

	// Handle different reading modes
	var contents string
	var linesShown string
	var outline string

	if fileReq.ShouldReadEntireFile {
		// Read entire file, but respect max lines limit
		if totalLines > maxLines {
			contents = strings.Join(lines[:maxLines], "\n")
			linesShown = fmt.Sprintf("1-%d (truncated from %d total lines due to max_lines limit)", maxLines, totalLines)
		} else {
			contents = strings.Join(lines, "\n")
			linesShown = fmt.Sprintf("1-%d (entire file)", totalLines)
		}
		if includeOutline {
			if totalLines > maxLines {
				outline = generateOutline(lines[:maxLines], fileReq.TargetFile)
			} else {
				outline = generateOutline(lines, fileReq.TargetFile)
			}
		}
	} else {
		// Read specific range
		startLine := fileReq.StartLineOneIndexed
		endLine := fileReq.EndLineOneIndexedInclusive

		// Validate line numbers
		if startLine < 1 {
			startLine = 1
		}
		if endLine > totalLines {
			endLine = totalLines
		}
		if startLine > endLine {
			response.Error = fmt.Sprintf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)
			return response
		}

		// Check line limits
		requestedLines := endLine - startLine + 1
		if requestedLines > maxLines {
			// Adjust end line to stay within max lines limit
			endLine = startLine + maxLines - 1
			if endLine > totalLines {
				endLine = totalLines
			}
		}

		// Apply minimum lines rule for partial reads (unless file is smaller)
		if requestedLines < globalMinLines && totalLines > globalMinLines {
			// Expand range to meet minimum, but respect file boundaries and max lines
			additionalLines := globalMinLines - requestedLines
			if additionalLines > maxLines-requestedLines {
				additionalLines = maxLines - requestedLines
			}

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

			// Ensure we don't exceed max lines
			if newEndLine-newStartLine+1 > maxLines {
				newEndLine = newStartLine + maxLines - 1
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
					fileReq.StartLineOneIndexed, fileReq.EndLineOneIndexedInclusive, startLine, endLine)
				contents = contextInfo + "\n" + contents
			}
		}

		if includeOutline {
			outline = generateOutline(selectedLines, fileReq.TargetFile)
		}
	}

	response.Contents = contents
	response.LinesShown = linesShown
	response.Outline = outline

	return response
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

func ParseJSONRequest(jsonInput string) (BatchReadFileRequest, error) {
	var req BatchReadFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return BatchReadFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the batch_read_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req BatchReadFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := BatchReadFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
