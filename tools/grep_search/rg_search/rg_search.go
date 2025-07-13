package rg_search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/llm-tools/tools/dirs"
	"github.com/xhd2015/llm-tools/tools/grep_search/model"
)

// GrepSearchRequest represents the input parameters for the grep_search tool
type GrepSearchRequest = model.GrepSearchRequest

// GrepSearchMatch represents a single search match result
type GrepSearchMatch = model.GrepSearchMatch

// GrepSearchResponse represents the output of the grep_search tool
type GrepSearchResponse = model.GrepSearchResponse

// RipgrepSearcher implements GrepSearcher using ripgrep (rg)
type RipgrepSearcher struct{}

// NewRipgrepSearcher creates a new ripgrep-based searcher
func NewRipgrepSearcher() *RipgrepSearcher {
	return &RipgrepSearcher{}
}

// IsAvailable checks if ripgrep is available in the system
func (r *RipgrepSearcher) IsAvailable() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

// Search executes the grep_search tool using ripgrep
func (r *RipgrepSearcher) Search(req GrepSearchRequest) (*GrepSearchResponse, error) {
	// Validate input parameters
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Check if ripgrep is available
	if !r.IsAvailable() {
		return nil, fmt.Errorf("ripgrep (rg) is not installed or not in PATH")
	}

	// Try JSON-based search first
	response, err := r.searchWithJSON(req)
	if err != nil {
		// Fall back to simple approach if JSON parsing fails
		response, err = r.searchSimple(req)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

// searchWithJSON executes ripgrep with JSON output
func (r *RipgrepSearcher) searchWithJSON(req GrepSearchRequest) (*GrepSearchResponse, error) {
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.RelativePathToSearch, "relative_path_to_search", true)
	if err != nil {
		return nil, err
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
	cmd.Dir = filePath
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
	matches, err := r.ParseRipgrepOutput(string(output))
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

// searchSimple provides a simpler interface using basic ripgrep without JSON parsing
func (r *RipgrepSearcher) searchSimple(req GrepSearchRequest) (*GrepSearchResponse, error) {
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
	matches, err := r.ParseSimpleRipgrepOutput(string(output))
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

// ParseRipgrepOutput parses the JSON output from ripgrep
func (r *RipgrepSearcher) ParseRipgrepOutput(output string) ([]GrepSearchMatch, error) {
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

// ParseSimpleRipgrepOutput parses the simple text output from ripgrep
func (r *RipgrepSearcher) ParseSimpleRipgrepOutput(output string) ([]GrepSearchMatch, error) {
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
