package file_search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// FileSearchRequest represents the input parameters for the file_search tool
type FileSearchRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	Query         string `json:"query"`
	Explanation   string `json:"explanation"`
}

// FileSearchMatch represents a single file search result
type FileSearchMatch struct {
	File  string  `json:"file"`
	Score float64 `json:"score"`
}

// FileSearchResponse represents the output of the file_search tool
type FileSearchResponse struct {
	TotalMatches int               `json:"total_matches"`
	Matches      []FileSearchMatch `json:"matches"`
	Truncated    bool              `json:"truncated"`
}

// GetToolDefinition returns the JSON schema definition for the file_search tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: "Fast file search based on fuzzy matching against file path. Use if you know part of the file path but don't know where it's located exactly. Response will be capped to 10 results. Make your query more specific if need to filter results further.",
		Name:        "file_search",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.WORKSPACE_ROOT,
				},
				"query": {
					Type:        jsonschema.ParamTypeString,
					Description: "Fuzzy filename to search for",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.EXPLANATION,
				},
			},
			Required: []string{"query"},
		},
	}
}

// FileSearch executes the file_search tool with the given parameters
func FileSearch(req FileSearchRequest) (*FileSearchResponse, error) {
	searchPath, err := dirs.GetPath(req.WorkspaceRoot, "", "", true)
	if err != nil {
		return nil, err
	}
	// Validate that workspace root exists
	if _, err := os.Stat(req.WorkspaceRoot); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace root does not exist: %s", req.WorkspaceRoot)
	}

	var matches []FileSearchMatch

	// Walk through all files in the workspace
	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip common directories that are not useful for file search
			dirName := filepath.Base(path)
			if dirName == ".git" || dirName == "node_modules" || dirName == ".vscode" ||
				dirName == "dist" || dirName == "build" || dirName == "target" ||
				dirName == "vendor" || dirName == ".idea" || dirName == ".DS_Store" {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate fuzzy match score
		score := calculateFuzzyScore(path, req.Query, searchPath)
		if score > 0 {
			// Make path relative to workspace root for display
			relPath, err := filepath.Rel(searchPath, path)
			if err != nil {
				relPath = path
			}

			matches = append(matches, FileSearchMatch{
				File:  relPath,
				Score: score,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Sort by score (descending) and limit to 10 results
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	truncated := len(matches) > 10
	if truncated {
		matches = matches[:10]
	}

	return &FileSearchResponse{
		TotalMatches: len(matches),
		Matches:      matches,
		Truncated:    truncated,
	}, nil
}

// calculateFuzzyScore calculates a fuzzy match score for a file path
func calculateFuzzyScore(filePath, query, workspaceRoot string) float64 {
	// Make path relative to workspace root for scoring
	relPath, err := filepath.Rel(workspaceRoot, filePath)
	if err != nil {
		relPath = filePath
	}

	// Normalize paths for comparison
	relPath = filepath.ToSlash(relPath)
	query = strings.ToLower(query)
	relPathLower := strings.ToLower(relPath)

	// Exact match gets highest score
	if relPathLower == query {
		return 1.0
	}

	// Filename exact match gets very high score
	filename := filepath.Base(relPath)
	filenameLower := strings.ToLower(filename)
	if filenameLower == query {
		return 0.9
	}

	// Filename contains query gets high score
	if strings.Contains(filenameLower, query) {
		return 0.8
	}

	// Path contains query gets good score
	if strings.Contains(relPathLower, query) {
		return 0.7
	}

	// Try fuzzy matching
	fuzzyScore := fuzzyMatch(relPathLower, query)
	if fuzzyScore > 0.3 {
		return fuzzyScore * 0.6 // Scale down fuzzy matches
	}

	return 0
}

// fuzzyMatch performs fuzzy string matching
func fuzzyMatch(text, pattern string) float64 {
	if pattern == "" {
		return 0
	}

	if text == pattern {
		return 1.0
	}

	// Simple fuzzy matching: check if all characters in pattern appear in order in text
	textRunes := []rune(text)
	patternRunes := []rune(pattern)

	if len(patternRunes) > len(textRunes) {
		return 0
	}

	textIndex := 0
	matchedChars := 0

	for _, patternChar := range patternRunes {
		found := false
		for textIndex < len(textRunes) {
			if textRunes[textIndex] == patternChar {
				matchedChars++
				textIndex++
				found = true
				break
			}
			textIndex++
		}
		if !found {
			return 0
		}
	}

	// Score based on how many characters matched and how close they are
	if matchedChars == len(patternRunes) {
		// Calculate score based on match density
		score := float64(matchedChars) / float64(len(textRunes))

		// Bonus for consecutive matches
		consecutiveBonus := calculateConsecutiveBonus(text, pattern)
		score += consecutiveBonus * 0.3

		// Bonus for matches at word boundaries
		wordBoundaryBonus := calculateWordBoundaryBonus(text, pattern)
		score += wordBoundaryBonus * 0.2

		return score
	}

	return 0
}

// calculateConsecutiveBonus calculates bonus for consecutive character matches
func calculateConsecutiveBonus(text, pattern string) float64 {
	if pattern == "" {
		return 0
	}

	maxConsecutive := 0
	currentConsecutive := 0

	textIndex := 0
	for _, patternChar := range pattern {
		found := false
		for textIndex < len(text) {
			if rune(text[textIndex]) == patternChar {
				if currentConsecutive > 0 {
					currentConsecutive++
				} else {
					currentConsecutive = 1
				}
				if currentConsecutive > maxConsecutive {
					maxConsecutive = currentConsecutive
				}
				textIndex++
				found = true
				break
			} else {
				currentConsecutive = 0
				textIndex++
			}
		}
		if !found {
			break
		}
	}

	return float64(maxConsecutive) / float64(len(pattern))
}

// calculateWordBoundaryBonus calculates bonus for matches at word boundaries
func calculateWordBoundaryBonus(text, pattern string) float64 {
	if pattern == "" {
		return 0
	}

	// Simple word boundary detection: spaces, slashes, dots, underscores
	wordBoundaries := []rune{' ', '/', '.', '_', '-'}

	boundaryMatches := 0
	textRunes := []rune(text)

	for _, patternChar := range pattern {
		for i, textChar := range textRunes {
			if textChar == patternChar {
				// Check if this is at a word boundary
				if i == 0 || contains(wordBoundaries, textRunes[i-1]) {
					boundaryMatches++
				}
				break
			}
		}
	}

	return float64(boundaryMatches) / float64(len(pattern))
}

// contains checks if a slice contains a specific rune
func contains(slice []rune, item rune) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ParseJSONRequest(jsonInput string) (FileSearchRequest, error) {
	var req FileSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return FileSearchRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the file_search tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req FileSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := FileSearch(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
