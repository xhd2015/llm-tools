package codebase_search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// CodebaseSearchRequest represents the input parameters for the codebase_search tool
type CodebaseSearchRequest struct {
	WorkspaceRoot     string   `json:"workspace_root"`
	Query             string   `json:"query"`
	SearchOnlyPrs     bool     `json:"search_only_prs"`
	TargetDirectories []string `json:"target_directories"`
	Explanation       string   `json:"explanation"`
}

// CodebaseSearchMatch represents a single search result
type CodebaseSearchMatch struct {
	File    string  `json:"file"`
	Content string  `json:"content"`
	Line    int     `json:"line"`
	Score   float64 `json:"score"`
	Context string  `json:"context"`
}

// CodebaseSearchResponse represents the output of the codebase_search tool
type CodebaseSearchResponse struct {
	Query        string                `json:"query"`
	TotalMatches int                   `json:"total_matches"`
	Matches      []CodebaseSearchMatch `json:"matches"`
	Truncated    bool                  `json:"truncated"`
}

// GetToolDefinition returns the JSON schema definition for the codebase_search tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `semantic search that finds code by meaning, not exact text

### When to Use This Tool

Use codebase_search when you need to:
- Explore unfamiliar codebases
- Ask "how / where / what" questions to understand behavior
- Find code by meaning rather than exact text

### When NOT to Use

Skip codebase_search for:
1. Exact text matches (use grep_search)
2. Reading known files (use read_file)
3. Simple symbol lookups (use grep_search)
4. Find file by name (use file_search)

### Examples

<example>
  Query: "Where is interface MyInterface implemented in the frontend?"

<reasoning>
  Good: Complete question asking about implementation location with specific context (frontend).
</reasoning>
</example>

<example>
  Query: "Where do we encrypt user passwords before saving?"

<reasoning>
  Good: Clear question about a specific process with context about when it happens.
</reasoning>
</example>

<example>
  Query: "MyInterface frontend"

<reasoning>
  BAD: Too vague; use a specific question instead. This would be better as "Where is MyInterface used in the frontend?"
</reasoning>
</example>

<example>
  Query: "AuthService"

<reasoning>
  BAD: Single word searches should use grep_search for exact text matching instead.
</reasoning>
</example>

<example>
  Query: "What is AuthService? How does AuthService work?"

<reasoning>
  BAD: Combines two separate queries together. Semantic search is not good at looking for multiple things in parallel. Split into separate searches: first "What is AuthService?" then "How does AuthService work?"
</reasoning>
</example>

### Target Directories

- Provide ONE directory or file path; [] searches the whole repo. No globs or wildcards.
  Good:
  - ["backend/api/"]   - focus directory
  - ["src/components/Button.tsx"] - single file
  - [] - search everywhere when unsure
  BAD:
  - ["frontend/", "backend/"] - multiple paths
  - ["src/**/utils/**"] - globs
  - ["*.ts"] or ["**/*"] - wildcard paths

### Search Strategy

1. Start with exploratory queries - semantic search is powerful and often finds relevant context in one go. Begin broad with [].
2. Review results; if a directory or file stands out, rerun with that as the target.
3. Break large questions into smaller ones (e.g. auth roles vs session storage).
4. For big files (>1K lines) run codebase_search scoped to that file instead of reading the entire file.

<example>
  Step 1: { "query": "How does user authentication work?", "target_directories": [], "explanation": "Find auth flow" }
  Step 2: Suppose results point to backend/auth/ → rerun:
          { "query": "Where are user roles checked?", "target_directories": ["backend/auth/"], "explanation": "Find role logic" }

<reasoning>
  Good strategy: Start broad to understand overall system, then narrow down to specific areas based on initial results.
</reasoning>
</example>

<example>
  Query: "How are websocket connections handled?"
  Target: ["backend/services/realtime.ts"]

<reasoning>
  Good: We know the answer is in this specific file, but the file is too large to read entirely, so we use semantic search to find the relevant parts.
</reasoning>
</example>`,
		Name: "codebase_search",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"query": {
					Type:        jsonschema.ParamTypeString,
					Description: "A complete question about what you want to understand. Ask as if talking to a colleague: 'How does X work?', 'What happens when Y?', 'Where is Z handled?'",
				},
				"search_only_prs": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "If true, only search pull requests and return no code results.",
				},
				"target_directories": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Prefix directory paths to limit search scope (single directory only, no glob patterns)",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"explanation", "query", "target_directories"},
		},
	}
}

// CodebaseSearch executes the codebase_search tool with the given parameters
func CodebaseSearch(req CodebaseSearchRequest) (*CodebaseSearchResponse, error) {
	// For now, this is a basic implementation that simulates semantic search
	// In a real implementation, this would use vector embeddings or similar technology
	// For the current implementation, we'll use a simple text-based approach

	var searchPaths []string
	if len(req.TargetDirectories) == 0 {
		searchPathRoot, err := dirs.GetPath(req.WorkspaceRoot, "", "", true)
		if err != nil {
			return nil, err
		}
		searchPaths = []string{searchPathRoot}
	} else {
		for _, dir := range req.TargetDirectories {
			searchPath, err := dirs.GetPath(req.WorkspaceRoot, dir, "target_directory", true)
			if err != nil {
				return nil, fmt.Errorf("invalid target directory %s: %w", dir, err)
			}
			searchPaths = append(searchPaths, searchPath)
		}
	}

	var matches []CodebaseSearchMatch

	// Simple implementation: search for keywords in the query across files
	keywords := extractKeywords(req.Query)

	for _, searchPath := range searchPaths {
		pathMatches, err := searchInPath(searchPath, keywords, req.Query)
		if err != nil {
			return nil, fmt.Errorf("error searching in path %s: %w", searchPath, err)
		}
		matches = append(matches, pathMatches...)
	}

	// Sort by score (descending) and limit results
	matches = sortAndLimitMatches(matches, 50)

	return &CodebaseSearchResponse{
		Query:        req.Query,
		TotalMatches: len(matches),
		Matches:      matches,
		Truncated:    len(matches) >= 50,
	}, nil
}

// extractKeywords extracts relevant keywords from the search query
func extractKeywords(query string) []string {
	// Simple keyword extraction - in a real implementation this would be more sophisticated
	words := strings.Fields(strings.ToLower(query))
	var keywords []string

	// Filter out common words
	stopWords := map[string]bool{
		"the": true, "is": true, "are": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "how": true, "what": true, "where": true, "when": true,
		"why": true, "who": true, "which": true, "does": true, "do": true, "did": true,
		"will": true, "would": true, "could": true, "should": true, "can": true, "may": true,
		"might": true, "must": true, "shall": true, "a": true, "an": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "you": true, "he": true,
		"she": true, "it": true, "we": true, "they": true, "me": true, "him": true,
		"her": true, "us": true, "them": true, "my": true, "your": true, "his": true,
		"its": true, "our": true, "their": true,
	}

	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:")
		if len(cleaned) > 2 && !stopWords[cleaned] {
			keywords = append(keywords, cleaned)
		}
	}

	return keywords
}

// searchInPath searches for matches in a given path
func searchInPath(searchPath string, keywords []string, originalQuery string) ([]CodebaseSearchMatch, error) {
	var matches []CodebaseSearchMatch

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip common directories that are not useful for code search
			dirName := filepath.Base(path)
			if dirName == ".git" || dirName == "node_modules" || dirName == ".vscode" ||
				dirName == "dist" || dirName == "build" || dirName == "target" ||
				dirName == "vendor" || dirName == ".idea" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only search in code files
		if !isCodeFile(path) {
			return nil
		}

		fileMatches, err := searchInFile(path, keywords, originalQuery)
		if err != nil {
			return nil // Skip files that can't be read
		}

		matches = append(matches, fileMatches...)
		return nil
	})

	return matches, err
}

// isCodeFile checks if a file is a code file worth searching
func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExtensions := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".py": true, ".java": true, ".cpp": true, ".c": true, ".h": true,
		".cs": true, ".php": true, ".rb": true, ".rs": true, ".swift": true,
		".kt": true, ".scala": true, ".clj": true, ".sh": true, ".bash": true,
		".zsh": true, ".fish": true, ".ps1": true, ".bat": true, ".cmd": true,
		".sql": true, ".html": true, ".css": true, ".scss": true, ".sass": true,
		".less": true, ".vue": true, ".svelte": true, ".elm": true, ".dart": true,
		".r": true, ".m": true, ".mm": true, ".pl": true, ".pm": true, ".lua": true,
		".vim": true, ".md": true, ".rst": true, ".txt": true, ".json": true,
		".yaml": true, ".yml": true, ".toml": true, ".ini": true, ".cfg": true,
		".conf": true, ".config": true, ".xml": true, ".proto": true, ".thrift": true,
		".graphql": true, ".gql": true, ".dockerfile": true, ".makefile": true,
	}

	return codeExtensions[ext] || filepath.Base(path) == "Makefile" ||
		filepath.Base(path) == "Dockerfile" || filepath.Base(path) == "Rakefile"
}

// searchInFile searches for keywords in a specific file
func searchInFile(filePath string, keywords []string, originalQuery string) ([]CodebaseSearchMatch, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)
	lines := strings.Split(text, "\n")

	var matches []CodebaseSearchMatch

	for i, line := range lines {
		score := calculateScore(line, keywords, originalQuery)
		if score > 0.1 { // Threshold for relevance
			matches = append(matches, CodebaseSearchMatch{
				File:    filePath,
				Content: strings.TrimSpace(line),
				Line:    i + 1,
				Score:   score,
				Context: getContext(lines, i, 2),
			})
		}
	}

	return matches, nil
}

// calculateScore calculates relevance score for a line
func calculateScore(line string, keywords []string, originalQuery string) float64 {
	if len(keywords) == 0 {
		return 0
	}

	lineLower := strings.ToLower(line)
	var score float64

	// Score based on keyword matches
	for _, keyword := range keywords {
		if strings.Contains(lineLower, keyword) {
			score += 1.0 / float64(len(keywords))
		}
	}

	// Bonus for exact phrase matches
	if strings.Contains(lineLower, strings.ToLower(originalQuery)) {
		score += 0.5
	}

	// Bonus for function/class/method definitions
	if strings.Contains(lineLower, "func ") || strings.Contains(lineLower, "function ") ||
		strings.Contains(lineLower, "class ") || strings.Contains(lineLower, "def ") ||
		strings.Contains(lineLower, "interface ") || strings.Contains(lineLower, "type ") {
		score += 0.3
	}

	// Bonus for comments that might explain functionality
	if strings.Contains(lineLower, "//") || strings.Contains(lineLower, "/*") ||
		strings.Contains(lineLower, "#") || strings.Contains(lineLower, "\"\"\"") {
		score += 0.2
	}

	return score
}

// getContext returns surrounding lines for context
func getContext(lines []string, lineIndex int, contextSize int) string {
	start := lineIndex - contextSize
	if start < 0 {
		start = 0
	}

	end := lineIndex + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}

	contextLines := lines[start:end]
	return strings.Join(contextLines, "\n")
}

// sortAndLimitMatches sorts matches by score and limits the number of results
func sortAndLimitMatches(matches []CodebaseSearchMatch, limit int) []CodebaseSearchMatch {
	// Simple bubble sort by score (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := 0; j < len(matches)-i-1; j++ {
			if matches[j].Score < matches[j+1].Score {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}

	if len(matches) > limit {
		return matches[:limit]
	}

	return matches
}

func ParseJSONRequest(jsonInput string) (CodebaseSearchRequest, error) {
	var req CodebaseSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return CodebaseSearchRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the codebase_search tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req CodebaseSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := CodebaseSearch(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
