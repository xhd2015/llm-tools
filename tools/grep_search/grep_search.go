package grep_search

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// GrepSearchRequest represents the input parameters for the grep_search tool
type GrepSearchRequest struct {
	Query          string `json:"query"`
	CaseSensitive  bool   `json:"case_sensitive,omitempty"`
	ExcludePattern string `json:"exclude_pattern,omitempty"`
	IncludePattern string `json:"include_pattern,omitempty"`
	Explanation    string `json:"explanation"`
}

// GrepSearchMatch represents a single search match result
type GrepSearchMatch struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column,omitempty"`
	Content    string `json:"content"`
	MatchStart int    `json:"match_start,omitempty"`
	MatchEnd   int    `json:"match_end,omitempty"`
}

// GrepSearchResponse represents the output of the grep_search tool
type GrepSearchResponse struct {
	Matches      []GrepSearchMatch `json:"matches"`
	TotalMatches int               `json:"total_matches"`
	SearchQuery  string            `json:"search_query"`
	Truncated    bool              `json:"truncated"`
}

// GetToolDefinition returns the JSON schema definition for the grep_search tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `### Instructions:
This is best for finding exact text matches or regex patterns.
This is preferred over semantic search when we know the exact symbol/function name/etc. to search in some set of directories/file types.

Use this tool to run fast, exact regex searches over text files using the ripgrep engine.
To avoid overwhelming output, the results are capped at 50 matches.
Use the include or exclude patterns to filter the search scope by file type or specific paths.

- Always escape special regex characters: ( ) [ ] { } + * ? ^ $ | . \
- Use \ to escape any of these characters when they appear in your search string.
- Do NOT perform fuzzy or semantic matches.
- Return only a valid regex pattern string.

### Examples:
| Literal               | Regex Pattern            |
|-----------------------|--------------------------|
| function(             | function\(              |
| value[index]          | value\[index\]         |
| file.txt               | file\.txt                |
| user|admin            | user\|admin             |
| path\to\file         | path\\to\\file        |
| hello world           | hello world              |
| foo\(bar\)          | foo\\(bar\\)         |`,
		Name: "grep_search",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"query": {
					Type:        jsonschema.ParamTypeString,
					Description: "The regex pattern to search for",
				},
				"case_sensitive": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether the search should be case sensitive",
				},
				"exclude_pattern": {
					Type:        jsonschema.ParamTypeString,
					Description: "Glob pattern for files to exclude",
				},
				"include_pattern": {
					Type:        jsonschema.ParamTypeString,
					Description: "Glob pattern for files to include (e.g. '*.ts' for TypeScript files)",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"query"},
		},
	}
}

// GrepSearch executes the grep_search tool with the given parameters
func GrepSearch(req GrepSearchRequest) (*GrepSearchResponse, error) {
	// Validate input parameters
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Check if ripgrep is available
	if _, err := exec.LookPath("rg"); err != nil {
		return nil, fmt.Errorf("ripgrep (rg) is not installed or not in PATH")
	}

	// Build ripgrep command
	args := []string{
		"--json",            // Output in JSON format for easier parsing
		"--line-number",     // Include line numbers
		"--column",          // Include column numbers
		"--no-heading",      // Don't group by file
		"--max-count", "50", // Limit to 50 matches total
	}

	// Add case sensitivity option
	if !req.CaseSensitive {
		args = append(args, "--ignore-case")
	}

	// Add include pattern if specified
	if req.IncludePattern != "" {
		args = append(args, "--glob", req.IncludePattern)
	}

	// Add exclude pattern if specified
	if req.ExcludePattern != "" {
		args = append(args, "--glob", "!"+req.ExcludePattern)
	}

	// Add the search pattern
	args = append(args, req.Query)

	// Add search path (current directory)
	args = append(args, ".")

	// Execute ripgrep command
	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()

	// ripgrep returns exit code 1 when no matches are found, which is not an error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				// No matches found
				return &GrepSearchResponse{
					Matches:      []GrepSearchMatch{},
					TotalMatches: 0,
					SearchQuery:  req.Query,
					Truncated:    false,
				}, nil
			}
		}
		return nil, fmt.Errorf("ripgrep command failed: %w", err)
	}

	// Parse ripgrep JSON output
	matches, err := parseRipgrepOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ripgrep output: %w", err)
	}

	// Check if results were truncated
	truncated := len(matches) >= 50

	return &GrepSearchResponse{
		Matches:      matches,
		TotalMatches: len(matches),
		SearchQuery:  req.Query,
		Truncated:    truncated,
	}, nil
}

// parseRipgrepOutput parses the JSON output from ripgrep
func parseRipgrepOutput(output string) ([]GrepSearchMatch, error) {
	var matches []GrepSearchMatch

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse each JSON line from ripgrep
		var rgResult map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rgResult); err != nil {
			continue // Skip malformed lines
		}

		// Only process match results
		if rgResult["type"] != "match" {
			continue
		}

		data, ok := rgResult["data"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract file path
		path, ok := data["path"].(map[string]interface{})
		if !ok {
			continue
		}
		filePath, ok := path["text"].(string)
		if !ok {
			continue
		}

		// Extract line number
		lineNumber := 0
		if lineData, ok := data["line_number"].(float64); ok {
			lineNumber = int(lineData)
		}

		// Extract line content
		lines, ok := data["lines"].(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := lines["text"].(string)
		if !ok {
			continue
		}

		// Extract match positions
		submatches, ok := data["submatches"].([]interface{})
		if !ok || len(submatches) == 0 {
			continue
		}

		// Get the first submatch for column information
		firstMatch, ok := submatches[0].(map[string]interface{})
		if !ok {
			continue
		}

		column := 0
		matchStart := 0
		matchEnd := 0

		if startData, ok := firstMatch["start"].(float64); ok {
			column = int(startData) + 1 // Convert to 1-indexed
			matchStart = int(startData)
		}

		if endData, ok := firstMatch["end"].(float64); ok {
			matchEnd = int(endData)
		}

		// Create match result
		match := GrepSearchMatch{
			File:       filePath,
			Line:       lineNumber,
			Column:     column,
			Content:    strings.TrimRight(content, "\n\r"),
			MatchStart: matchStart,
			MatchEnd:   matchEnd,
		}

		matches = append(matches, match)
	}

	return matches, nil
}

// GrepSearchSimple provides a simpler interface using basic ripgrep without JSON parsing
// This is used as a fallback when JSON parsing fails
func GrepSearchSimple(req GrepSearchRequest) (*GrepSearchResponse, error) {
	// Validate input parameters
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Check if ripgrep is available
	if _, err := exec.LookPath("rg"); err != nil {
		return nil, fmt.Errorf("ripgrep (rg) is not installed or not in PATH")
	}

	// Build ripgrep command
	args := []string{
		"--line-number",     // Include line numbers
		"--no-heading",      // Don't group by file
		"--max-count", "50", // Limit to 50 matches total
	}

	// Add case sensitivity option
	if !req.CaseSensitive {
		args = append(args, "--ignore-case")
	}

	// Add include pattern if specified
	if req.IncludePattern != "" {
		args = append(args, "--glob", req.IncludePattern)
	}

	// Add exclude pattern if specified
	if req.ExcludePattern != "" {
		args = append(args, "--glob", "!"+req.ExcludePattern)
	}

	// Add the search pattern
	args = append(args, req.Query)

	// Add search path (current directory)
	args = append(args, ".")

	// Execute ripgrep command
	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()

	// ripgrep returns exit code 1 when no matches are found, which is not an error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				// No matches found
				return &GrepSearchResponse{
					Matches:      []GrepSearchMatch{},
					TotalMatches: 0,
					SearchQuery:  req.Query,
					Truncated:    false,
				}, nil
			}
		}
		return nil, fmt.Errorf("ripgrep command failed: %w", err)
	}

	// Parse simple ripgrep output
	matches, err := parseSimpleRipgrepOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ripgrep output: %w", err)
	}

	// Check if results were truncated
	truncated := len(matches) >= 50

	return &GrepSearchResponse{
		Matches:      matches,
		TotalMatches: len(matches),
		SearchQuery:  req.Query,
		Truncated:    truncated,
	}, nil
}

// parseSimpleRipgrepOutput parses the simple text output from ripgrep
func parseSimpleRipgrepOutput(output string) ([]GrepSearchMatch, error) {
	var matches []GrepSearchMatch

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse format: filename:line_number:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		lineNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		content := parts[2]

		// Create match result
		match := GrepSearchMatch{
			File:    filePath,
			Line:    lineNumber,
			Content: content,
		}

		matches = append(matches, match)
	}

	return matches, nil
}

// ExecuteFromJSON executes the grep_search tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req GrepSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Try the full JSON-based approach first
	response, err := GrepSearch(req)
	if err != nil {
		// Fall back to simple approach if JSON parsing fails
		response, err = GrepSearchSimple(req)
		if err != nil {
			return "", err
		}
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}

// ValidateRegexPattern validates and suggests corrections for regex patterns
func ValidateRegexPattern(pattern string) error {
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	return nil
}

// EscapeRegexSpecialChars escapes special regex characters in a literal string
func EscapeRegexSpecialChars(literal string) string {
	// Characters that need escaping in regex: ( ) [ ] { } + * ? ^ $ | . \
	specialChars := []string{
		"\\", "(", ")", "[", "]", "{", "}", "+", "*", "?", "^", "$", "|", ".",
	}

	result := literal
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}

// GetWorkingDirectory returns the current working directory for search context
func GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

// Main function for standalone execution
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: grep_search <json_input>")
		fmt.Println("Example: grep_search '{\"query\":\"function\",\"case_sensitive\":false}'")
		os.Exit(1)
	}

	jsonInput := os.Args[1]

	// If it's a file path, read the JSON from file
	if strings.HasSuffix(jsonInput, ".json") {
		file, err := os.Open(jsonInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening JSON file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		jsonBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
			os.Exit(1)
		}
		jsonInput = string(jsonBytes)
	}

	output, err := ExecuteFromJSON(jsonInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}
