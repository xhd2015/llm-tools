package pure_go_search

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/llm-tools/tools/dirs"
	"github.com/xhd2015/llm-tools/tools/grep_search/rg_search"
)

// PureGoSearcher implements GrepSearcher using pure Go without external dependencies
type PureGoSearcher struct{}

// NewPureGoSearcher creates a new pure Go searcher
func NewPureGoSearcher() PureGoSearcher {
	return PureGoSearcher{}
}

// IsAvailable always returns true since pure Go implementation doesn't require external dependencies
func (p PureGoSearcher) IsAvailable() bool {
	return true
}

// Search executes the grep_search tool using pure Go implementation
func (p PureGoSearcher) Search(req rg_search.GrepSearchRequest) (*rg_search.GrepSearchResponse, error) {
	// Validate input parameters
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.RelativePathToSearch, "relative_path_to_search", true)
	if err != nil {
		return nil, err
	}

	// Compile regex pattern
	regExpr := req.Query
	if !req.CaseSensitive {
		regExpr = "(?i)" + regExpr
	}
	regex, err := regexp.Compile(regExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Search for matches
	matches, err := p.searchFiles(filePath, regex, req)
	if err != nil {
		return nil, err
	}

	// Check if results were truncated
	truncated := len(matches) >= 50

	return &rg_search.GrepSearchResponse{
		Matches:      matches,
		TotalMatches: len(matches),
		SearchQuery:  req.Query,
		Truncated:    truncated,
	}, nil
}

// searchFiles recursively searches files in the directory
func (p PureGoSearcher) searchFiles(dir string, regex *regexp.Regexp, req rg_search.GrepSearchRequest) ([]rg_search.GrepSearchMatch, error) {
	var matches []rg_search.GrepSearchMatch
	var includePattern, excludePattern *regexp.Regexp

	// Compile include pattern if specified
	if req.IncludePattern != "" {
		includePattern, _ = regexp.Compile(p.globToRegex(req.IncludePattern))
	}

	// Compile exclude pattern if specified
	if req.ExcludePattern != "" {
		excludePattern, _ = regexp.Compile(p.globToRegex(req.ExcludePattern))
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Skip files that can't be read
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			relPath = path
		}

		// Check include pattern
		if includePattern != nil && !includePattern.MatchString(relPath) {
			return nil
		}

		// Check exclude pattern
		if excludePattern != nil && excludePattern.MatchString(relPath) {
			return nil
		}

		// Skip binary files and large files
		if p.shouldSkipFile(path) {
			return nil
		}

		// Search in file
		fileMatches, err := p.searchInFile(path, relPath, regex)
		if err != nil {
			return nil // Skip files that can't be read
		}

		matches = append(matches, fileMatches...)

		// Stop if we've reached the limit
		if len(matches) >= 50 {
			return fmt.Errorf("reached match limit")
		}

		return nil
	})

	// If we stopped due to reaching limit, ignore the error
	if err != nil && err.Error() == "reached match limit" {
		err = nil
	}

	// Limit to 50 matches
	if len(matches) > 50 {
		matches = matches[:50]
	}

	return matches, err
}

// searchInFile searches for matches in a single file
func (p PureGoSearcher) searchInFile(fullPath string, relPath string, regex *regexp.Regexp) ([]rg_search.GrepSearchMatch, error) {
	var matches []rg_search.GrepSearchMatch

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Find all matches in this line
		indices := regex.FindAllStringIndex(line, -1)
		for _, match := range indices {
			result := rg_search.GrepSearchMatch{
				File:       relPath,
				Line:       lineNumber,
				Column:     match[0] + 1, // Convert to 1-indexed
				Content:    line,
				MatchStart: match[0],
				MatchEnd:   match[1],
			}
			matches = append(matches, result)

			// Stop if we've reached the limit
			if len(matches) >= 50 {
				return matches, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// shouldSkipFile determines if a file should be skipped based on its properties
func (p PureGoSearcher) shouldSkipFile(filePath string) bool {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	// Skip large files (> 10MB)
	if info.Size() > 10*1024*1024 {
		return true
	}

	// Skip common binary file extensions
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExtensions := []string{
		".exe", ".dll", ".so", ".dylib", ".a", ".lib", ".o", ".obj",
		".bin", ".dat", ".db", ".sqlite", ".sqlite3",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".ico",
		".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
		".class", ".jar", ".war", ".ear",
		".pyc", ".pyo", ".pyd",
		".node", ".wasm",
	}

	for _, binaryExt := range binaryExtensions {
		if ext == binaryExt {
			return true
		}
	}

	// Check if file appears to be binary by reading first few bytes
	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return true
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

// globToRegex converts a glob pattern to a regular expression
func (p PureGoSearcher) globToRegex(glob string) string {
	// Simple glob to regex conversion
	// * -> .*
	// ? -> .
	// Escape other regex special characters
	result := strings.ReplaceAll(glob, ".", "\\.")
	result = strings.ReplaceAll(result, "*", ".*")
	result = strings.ReplaceAll(result, "?", ".")
	result = strings.ReplaceAll(result, "+", "\\+")
	result = strings.ReplaceAll(result, "^", "\\^")
	result = strings.ReplaceAll(result, "$", "\\$")
	result = strings.ReplaceAll(result, "(", "\\(")
	result = strings.ReplaceAll(result, ")", "\\)")
	result = strings.ReplaceAll(result, "[", "\\[")
	result = strings.ReplaceAll(result, "]", "\\]")
	result = strings.ReplaceAll(result, "{", "\\{")
	result = strings.ReplaceAll(result, "}", "\\}")
	result = strings.ReplaceAll(result, "|", "\\|")

	return "^" + result + "$"
}
