package grep_search

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/grep_search/model"
	"github.com/xhd2015/llm-tools/tools/grep_search/pure_go_search"
	"github.com/xhd2015/llm-tools/tools/grep_search/rg_search"
)

// Re-export types for backward compatibility
type GrepSearchRequest = model.GrepSearchRequest
type GrepSearchMatch = model.GrepSearchMatch
type GrepSearchResponse = model.GrepSearchResponse

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
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The root directory of the workspace",
				},
				"relative_path_to_search": {
					Type:        jsonschema.ParamTypeString,
					Description: "The relative path to the workspace root to search in",
				},
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

// createSearcher creates the appropriate searcher based on availability
func createSearcher() model.GrepSearcher {
	// Try ripgrep first
	ripgrepSearcher := rg_search.NewRipgrepSearcher()
	if ripgrepSearcher.IsAvailable() {
		return ripgrepSearcher
	}

	// Fall back to pure Go implementation
	fmt.Fprintf(os.Stderr, "Warning: grep_search will be slower due to missing ripgrep (rg). Consider installing ripgrep for better performance.\n")
	return pure_go_search.NewPureGoSearcher()
}

// GrepSearch executes the grep_search tool with the given parameters
func GrepSearch(req GrepSearchRequest) (*GrepSearchResponse, error) {
	searcher := createSearcher()
	return searcher.Search(req)
}

func GoGrepSearch(req GrepSearchRequest) (*GrepSearchResponse, error) {
	searcher := pure_go_search.NewPureGoSearcher()
	return searcher.Search(req)
}

// GrepSearchSimple provides a simpler interface for backward compatibility
// This is used as a fallback when JSON parsing fails
func GrepSearchSimple(req GrepSearchRequest) (*GrepSearchResponse, error) {
	// For backward compatibility, we'll use the main GrepSearch function
	return GrepSearch(req)
}

func ParseJSONRequest(jsonInput string) (GrepSearchRequest, error) {
	var req GrepSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return GrepSearchRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the grep_search tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req GrepSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := GrepSearch(req)
	if err != nil {
		return "", err
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
