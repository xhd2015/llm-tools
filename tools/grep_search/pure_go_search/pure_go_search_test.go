package pure_go_search

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/xhd2015/llm-tools/tools/grep_search/rg_search"
)

// createTestFiles creates temporary test files for testing
func createTestFiles(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create test files with known content
	files := map[string]string{
		"file1.txt": `line one
line two contains pattern
line three
line four with another pattern`,
		"file2.go": `package main

import "fmt"

func main() {
	fmt.Println("pattern in go file")
	pattern := "variable"
}`,
		"file3.json": `{
  "name": "test",
  "value": "pattern in json",
  "items": [
    {
      "id": 123,
      "pattern": "nested pattern"
    }
  ]
}`,
		"binary.bin": "\x00\x01\x02\x03pattern\x00\x01\x02", // Binary file
		"empty.txt":  "",                                    // Empty file
		"subdir/nested.txt": `nested file
with pattern in subdirectory
and multiple lines`,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	return tmpDir
}

func TestPureGoSearcher_IsAvailable(t *testing.T) {
	searcher := NewPureGoSearcher()
	if !searcher.IsAvailable() {
		t.Error("PureGoSearcher should always be available")
	}
}

func TestPureGoSearcher_Search_BasicFunctionality(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if response.SearchQuery != "pattern" {
		t.Errorf("Expected search query 'pattern', got '%s'", response.SearchQuery)
	}

	if response.TotalMatches == 0 {
		t.Error("Expected to find matches, but got 0")
	}

	// Should find matches in file1.txt, file2.go, file3.json, and subdir/nested.txt
	expectedFiles := []string{"file1.txt", "file2.go", "file3.json", "subdir/nested.txt"}
	foundFiles := make(map[string]bool)
	for _, match := range response.Matches {
		foundFiles[match.File] = true
	}

	for _, expectedFile := range expectedFiles {
		if !foundFiles[expectedFile] {
			t.Errorf("Expected to find matches in %s, but didn't", expectedFile)
		}
	}
}

func TestPureGoSearcher_Search_CaseSensitive(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	// Create a file with mixed case
	testFile := filepath.Join(tmpDir, "case_test.txt")
	if err := os.WriteFile(testFile, []byte("Pattern\npattern\nPATTERN"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		query         string
		caseSensitive bool
		expectedCount int
	}{
		{
			name:          "case insensitive",
			query:         "pattern",
			caseSensitive: false,
			expectedCount: 3, // Should match all three variants
		},
		{
			name:          "case sensitive",
			query:         "pattern",
			caseSensitive: true,
			expectedCount: 1, // Should match only lowercase
		},
		{
			name:          "case sensitive uppercase",
			query:         "PATTERN",
			caseSensitive: true,
			expectedCount: 1, // Should match only uppercase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := rg_search.GrepSearchRequest{
				WorkspaceRoot:        tmpDir,
				RelativePathToSearch: "",
				Query:                tt.query,
				CaseSensitive:        tt.caseSensitive,
			}

			response, err := searcher.Search(req)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			matchCount := 0
			for _, match := range response.Matches {
				if match.File == "case_test.txt" {
					matchCount++
				}
			}

			if matchCount != tt.expectedCount {
				t.Errorf("Expected %d matches, got %d", tt.expectedCount, matchCount)
			}
		})
	}
}

func TestPureGoSearcher_Search_IncludePattern(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
		IncludePattern:       "*.go",
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should only find matches in Go files
	for _, match := range response.Matches {
		matched, err := filepath.Match("*.go", match.File)
		if err != nil {
			t.Fatalf("Error matching pattern: %v", err)
		}
		if !matched {
			t.Errorf("Expected only Go files, but found match in %s", match.File)
		}
	}
}

func TestPureGoSearcher_Search_ExcludePattern(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
		ExcludePattern:       "*.json",
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should not find matches in JSON files
	for _, match := range response.Matches {
		matched, err := filepath.Match("*.json", match.File)
		if err != nil {
			t.Fatalf("Error matching pattern: %v", err)
		}
		if matched {
			t.Errorf("Expected to exclude JSON files, but found match in %s", match.File)
		}
	}
}

func TestPureGoSearcher_Search_RegexPattern(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	// Create a file with specific patterns
	testFile := filepath.Join(tmpDir, "regex_test.txt")
	content := `email@example.com
user@domain.org
not-an-email
another.email@test.co.uk`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                `\w+@\w+\.\w+`,
		CaseSensitive:        false,
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find email-like patterns
	emailCount := 0
	for _, match := range response.Matches {
		if match.File == "regex_test.txt" {
			emailCount++
		}
	}

	if emailCount == 0 {
		t.Error("Expected to find email patterns, but got 0")
	}
}

func TestPureGoSearcher_Search_EmptyQuery(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "",
		CaseSensitive:        false,
	}

	_, err := searcher.Search(req)
	if err == nil {
		t.Error("Expected error for empty query, but got none")
	}
}

func TestPureGoSearcher_Search_InvalidRegex(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "[invalid regex",
		CaseSensitive:        false,
	}

	_, err := searcher.Search(req)
	if err == nil {
		t.Error("Expected error for invalid regex, but got none")
	}
}

func TestPureGoSearcher_Search_NonexistentDirectory(t *testing.T) {
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        "/nonexistent/directory",
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
	}

	_, err := searcher.Search(req)
	if err == nil {
		t.Error("Expected error for nonexistent directory, but got none")
	}
}

func TestPureGoSearcher_Search_MatchDetails(t *testing.T) {
	tmpDir := createTestFiles(t)
	searcher := NewPureGoSearcher()

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify match details
	for _, match := range response.Matches {
		if match.Line <= 0 {
			t.Errorf("Expected positive line number, got %d", match.Line)
		}
		if match.Column <= 0 {
			t.Errorf("Expected positive column number, got %d", match.Column)
		}
		if match.Content == "" {
			t.Error("Expected non-empty content")
		}
		if match.File == "" {
			t.Error("Expected non-empty file path")
		}
	}
}

func TestPureGoSearcher_shouldSkipFile(t *testing.T) {
	searcher := NewPureGoSearcher()
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		fileName    string
		content     []byte
		shouldSkip  bool
		description string
	}{
		{
			name:        "text file",
			fileName:    "test.txt",
			content:     []byte("hello world"),
			shouldSkip:  false,
			description: "regular text file should not be skipped",
		},
		{
			name:        "binary file",
			fileName:    "test.bin",
			content:     []byte{0x00, 0x01, 0x02, 0x03},
			shouldSkip:  true,
			description: "binary file should be skipped",
		},
		{
			name:        "executable",
			fileName:    "test.exe",
			content:     []byte("hello world"),
			shouldSkip:  true,
			description: "executable file should be skipped by extension",
		},
		{
			name:        "go file",
			fileName:    "test.go",
			content:     []byte("package main"),
			shouldSkip:  false,
			description: "go file should not be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(filePath, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := searcher.shouldSkipFile(filePath)
			if result != tt.shouldSkip {
				t.Errorf("shouldSkipFile(%s) = %v, expected %v (%s)",
					tt.fileName, result, tt.shouldSkip, tt.description)
			}
		})
	}
}

func TestPureGoSearcher_globToRegex(t *testing.T) {
	searcher := NewPureGoSearcher()

	tests := []struct {
		glob     string
		expected string
	}{
		{"*.go", "^.*\\.go$"},
		{"test.txt", "^test\\.txt$"},
		{"dir/*", "^dir/.*$"},
		{"file?.txt", "^file.\\.txt$"},
		{"test[abc]", "^test\\[abc\\]$"},
		{"file{1,2}", "^file\\{1,2\\}$"},
	}

	for _, tt := range tests {
		t.Run(tt.glob, func(t *testing.T) {
			result := searcher.globToRegex(tt.glob)
			if result != tt.expected {
				t.Errorf("globToRegex(%s) = %s, expected %s", tt.glob, result, tt.expected)
			}

			// Also test that the regex compiles
			if _, err := regexp.Compile(result); err != nil {
				t.Errorf("Generated regex %s does not compile: %v", result, err)
			}
		})
	}
}

func TestPureGoSearcher_Search_Truncation(t *testing.T) {
	tmpDir := t.TempDir()
	searcher := NewPureGoSearcher()

	// Create a file with many matches
	content := ""
	for i := 0; i < 100; i++ {
		content += "pattern line " + string(rune(i)) + "\n"
	}

	testFile := filepath.Join(tmpDir, "many_matches.txt")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := rg_search.GrepSearchRequest{
		WorkspaceRoot:        tmpDir,
		RelativePathToSearch: "",
		Query:                "pattern",
		CaseSensitive:        false,
	}

	response, err := searcher.Search(req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should be truncated to 50 matches
	if len(response.Matches) > 50 {
		t.Errorf("Expected at most 50 matches, got %d", len(response.Matches))
	}

	if len(response.Matches) == 50 && !response.Truncated {
		t.Error("Expected truncated flag to be true when matches reach limit")
	}
}
