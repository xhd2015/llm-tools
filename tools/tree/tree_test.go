package tree

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/expected_outputs/basic_structure_default.txt
var expectedBasicStructureDefault string

//go:embed testdata/expected_outputs/basic_structure_directories_only.txt
var expectedBasicStructureDirectoriesOnly string

//go:embed testdata/expected_outputs/repeated_patterns_collapse.txt
var expectedRepeatedPatternsCollapse string

// generateDiff creates a simple diff output similar to git diff
func generateDiff(expected, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var result strings.Builder
	result.WriteString("--- expected\n")
	result.WriteString("+++ actual\n")

	maxLen := len(expectedLines)
	if len(actualLines) > maxLen {
		maxLen = len(actualLines)
	}

	hasDifferences := false
	for i := 0; i < maxLen; i++ {
		var expectedLine, actualLine string
		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			hasDifferences = true
			if expectedLine != "" {
				result.WriteString(fmt.Sprintf("-%s\n", expectedLine))
			}
			if actualLine != "" {
				result.WriteString(fmt.Sprintf("+%s\n", actualLine))
			}
		}
	}

	if !hasDifferences {
		result.WriteString("(no differences found - strings are identical)\n")
	}

	return result.String()
}

func TestTreeCollapsed(t *testing.T) {
	tests := []struct {
		name                string
		dir                 string
		opts                TreeCollapseOptions
		expectedContains    []string // strings that should be present in the output
		expectedNotContains []string // strings that should NOT be present in the output
		expectMatch         string   // exact expected output - when not empty, compare exactly
		wantErr             bool
	}{
		{
			name:        "basic_structure_default",
			dir:         "testdata/basic_structure",
			opts:        TreeCollapseOptions{},
			expectMatch: expectedBasicStructureDefault,
		},
		{
			name: "basic_structure_directories_only",
			dir:  "testdata/basic_structure",
			opts: TreeCollapseOptions{
				DirectoriesOnly: true,
			},
			expectMatch: expectedBasicStructureDirectoriesOnly,
		},
		{
			name:        "repeated_patterns_collapse",
			dir:         "testdata/repeated_patterns",
			opts:        TreeCollapseOptions{},
			expectMatch: expectedRepeatedPatternsCollapse,
		},
		{
			name: "nested_structure",
			dir:  "testdata/nested_structure",
			opts: TreeCollapseOptions{},
			expectedContains: []string{
				"nested_structure",
				"└── level1",
				"    └── level2",
				"        └── level3",
				"            └── level4",
				"                └── deep_file.txt",
			},
			expectedNotContains: []string{},
		},
		{
			name: "mixed_patterns",
			dir:  "testdata/mixed_patterns",
			opts: TreeCollapseOptions{},
			expectedContains: []string{
				"mixed_patterns",
				"0_batch (3 times)",
				"├── another_dir",
				"├── regular_dir",
				"└── normal_file.txt",
			},
			expectedNotContains: []string{
				"0_mixed_patterns",
				"1_batch",
				"2_batch",
			},
		},
		{
			name: "filter_test_include_go",
			dir:  "testdata/filter_test",
			opts: TreeCollapseOptions{
				IncludePatterns: []string{".*\\.go$", "go_files"},
			},
			expectedContains: []string{
				"filter_test",
				"go_files",
				"main.go",
				"utils.go",
				"test.go",
			},
			expectedNotContains: []string{
				"python_files",
				"data_files",
				"script.py",
				"data.json",
				"readme.md",
			},
		},
		{
			name: "filter_test_exclude_names_index",
			dir:  "testdata/filter_test",
			opts: TreeCollapseOptions{
				ExcludePatterns: []string{"names_index"},
			},
			expectedContains: []string{
				"filter_test",
				"go_files",
				"python_files",
				"data_files",
				"logs",
			},
			expectedNotContains: []string{
				"names_index",
			},
		},
		{
			name: "filter_test_directories_only_exclude_logs",
			dir:  "testdata/filter_test",
			opts: TreeCollapseOptions{
				DirectoriesOnly: true,
				ExcludePatterns: []string{"logs"},
			},
			expectedContains: []string{
				"filter_test",
				"go_files",
				"python_files",
				"data_files",
				"names_index",
			},
			expectedNotContains: []string{
				"logs",
				"main.go",
				"script.py",
				"readme.md",
			},
		},
		{
			name: "collapse_patterns_test",
			dir:  "testdata/collapse_patterns",
			opts: TreeCollapseOptions{},
			expectedContains: []string{
				"collapse_patterns",
				"0_batch (2 times)",
				"0_similar (2 times)",
				"└── unique_dir",
			},
			expectedNotContains: []string{
				"1_batch",
				"1_similar",
			},
		},
		{
			name: "empty_directory",
			dir:  "testdata/empty_dir",
			opts: TreeCollapseOptions{},
			expectedContains: []string{
				"empty_dir",
			},
			expectedNotContains: []string{},
		},
		{
			name: "single_file",
			dir:  "testdata/single_file",
			opts: TreeCollapseOptions{},
			expectedContains: []string{
				"single_file",
				"└── only_file.txt",
			},
			expectedNotContains: []string{},
		},
		{
			name: "single_file_directories_only",
			dir:  "testdata/single_file",
			opts: TreeCollapseOptions{
				DirectoriesOnly: true,
			},
			expectedContains: []string{
				"single_file",
			},
			expectedNotContains: []string{
				"only_file.txt",
			},
		},
		{
			name: "complex_include_exclude",
			dir:  "testdata/filter_test",
			opts: TreeCollapseOptions{
				IncludePatterns: []string{".*files$", ".*\\.go$", ".*\\.py$"},
				ExcludePatterns: []string{"data_files"},
			},
			expectedContains: []string{
				"filter_test",
				"go_files",
				"python_files",
				"main.go",
				"utils.go",
				"test.go",
				"script.py",
				"test.py",
			},
			expectedNotContains: []string{
				"data_files",
				"logs",
				"names_index",
				"readme.md",
				"Makefile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert relative path to absolute path
			absDir := filepath.Join(".", tt.dir)

			result, err := TreeCollapsed(absDir, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("TreeCollapsed() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("TreeCollapsed() unexpected error: %v", err)
				return
			}

			if tt.expectMatch != "" {
				if result != tt.expectMatch {
					t.Errorf("TreeCollapsed() result does not match expected output:\n%s", generateDiff(tt.expectMatch, result))
				}
			} else {
				// Check that expected strings are present
				for _, expected := range tt.expectedContains {
					if !strings.Contains(result, expected) {
						t.Errorf("TreeCollapsed() result should contain %q\nGot:\n%s", expected, result)
					}
				}

				// Check that unwanted strings are not present
				for _, notExpected := range tt.expectedNotContains {
					if strings.Contains(result, notExpected) {
						t.Errorf("TreeCollapsed() result should NOT contain %q\nGot:\n%s", notExpected, result)
					}
				}
			}
		})
	}
}

func TestTreeCollapsedErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		opts    TreeCollapseOptions
		wantErr bool
	}{
		{
			name:    "non_existent_directory",
			dir:     "testdata/does_not_exist",
			opts:    TreeCollapseOptions{},
			wantErr: true,
		},
		{
			name: "invalid_include_pattern",
			dir:  "testdata/basic_structure",
			opts: TreeCollapseOptions{
				IncludePatterns: []string{"[invalid(regex"},
			},
			wantErr: true,
		},
		{
			name: "invalid_exclude_pattern",
			dir:  "testdata/basic_structure",
			opts: TreeCollapseOptions{
				ExcludePatterns: []string{"[invalid(regex"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDir := filepath.Join(".", tt.dir)

			_, err := TreeCollapsed(absDir, tt.opts)

			if tt.wantErr && err == nil {
				t.Errorf("TreeCollapsed() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("TreeCollapsed() unexpected error: %v", err)
			}
		})
	}
}

func TestTreeCollapseOptions(t *testing.T) {
	// Test the individual collapse features more specifically
	t.Run("repeated_collapse_detailed", func(t *testing.T) {
		result, err := TreeCollapsed("testdata/repeated_patterns", TreeCollapseOptions{})
		if err != nil {
			t.Fatalf("TreeCollapsed() error: %v", err)
		}

		// Should collapse directories with pattern "X_test" into "test (3 times)"
		if !strings.Contains(result, "test (3 times)") {
			t.Errorf("Expected 'test (3 times)' in result, got:\n%s", result)
		}

		// Should still contain some individual directory names that weren't collapsed
		expectedPatterns := []string{"3_test", "4_test"}
		for _, pattern := range expectedPatterns {
			if !strings.Contains(result, pattern) {
				t.Errorf("Result should contain '%s', got:\n%s", pattern, result)
			}
		}
	})

	t.Run("mixed_collapse_behavior", func(t *testing.T) {
		result, err := TreeCollapsed("testdata/mixed_patterns", TreeCollapseOptions{})
		if err != nil {
			t.Fatalf("TreeCollapsed() error: %v", err)
		}

		// Should collapse numeric pattern directories
		if !strings.Contains(result, "batch (3 times)") {
			t.Errorf("Expected 'batch (3 times)' in result, got:\n%s", result)
		}

		// Should keep non-pattern directories
		if !strings.Contains(result, "regular_dir") {
			t.Errorf("Expected 'regular_dir' in result, got:\n%s", result)
		}
		if !strings.Contains(result, "another_dir") {
			t.Errorf("Expected 'another_dir' in result, got:\n%s", result)
		}

		// Should keep non-pattern files
		if !strings.Contains(result, "normal_file.txt") {
			t.Errorf("Expected 'normal_file.txt' in result, got:\n%s", result)
		}
	})
}

func TestTreeCollapsePatternExtraction(t *testing.T) {
	// Test the pattern extraction logic
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single_digit", "0_test", "test"},
		{"double_digit", "10_batch", "batch"},
		{"triple_digit", "123_data", "data"},
		{"no_pattern", "regular_name", "regular_name"},
		{"underscore_only", "_test", "_test"},
		{"multiple_underscores", "1_test_data", "test_data"},
		{"no_underscore", "test123", "test123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPattern(tt.input)
			if result != tt.expected {
				t.Errorf("extractPattern(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTreeCollapseBasePattern(t *testing.T) {
	// Test the base pattern extraction logic
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single_digit", "0_test", "test"},
		{"double_digit", "10_batch", "batch"},
		{"triple_digit", "123_data", "data"},
		{"no_pattern", "regular_name", "regular_name"},
		{"underscore_only", "_test", "_test"},
		{"multiple_underscores", "1_test_data", "test_data"},
		{"no_underscore", "test123", "test123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBasePattern(tt.input)
			if result != tt.expected {
				t.Errorf("extractBasePattern(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
