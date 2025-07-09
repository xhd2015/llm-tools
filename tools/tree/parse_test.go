package tree

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Item
		wantErr  bool
	}{
		{
			name:  "empty_root",
			input: "root",
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children:     nil,
			},
		},
		{
			name: "single_child",
			input: `root
└── child1`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "child1", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "multiple_children",
			input: `root
├── child1
├── child2
└── child3`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "child1", Index: 0, MissingIndex: true, Children: nil},
					{Name: "child2", Index: 0, MissingIndex: true, Children: nil},
					{Name: "child3", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "nested_structure",
			input: `root
├── dir1
│   ├── file1.txt
│   └── file2.go
└── dir2
    └── file3.md`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{
						Name:         "dir1",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{Name: "file1.txt", Index: 0, MissingIndex: true, Children: nil},
							{Name: "file2.go", Index: 0, MissingIndex: true, Children: nil},
						},
					},
					{
						Name:         "dir2",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{Name: "file3.md", Index: 0, MissingIndex: true, Children: nil},
						},
					},
				},
			},
		},
		{
			name: "deep_nesting",
			input: `root
└── level1
    └── level2
        └── level3
            └── deep_file.txt`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{
						Name:         "level1",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{
								Name:         "level2",
								Index:        0,
								MissingIndex: true,
								Children: []Item{
									{
										Name:         "level3",
										Index:        0,
										MissingIndex: true,
										Children: []Item{
											{Name: "deep_file.txt", Index: 0, MissingIndex: true, Children: nil},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "collapsed_entries",
			input: `root
└── test (5 times)`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "test (5 times)", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "mixed_collapsed_and_regular",
			input: `root
├── batch (3 times)
├── regular_dir
└── file.txt (2 times)`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "batch (3 times)", Index: 0, MissingIndex: true, Children: nil},
					{Name: "regular_dir", Index: 0, MissingIndex: true, Children: nil},
					{Name: "file.txt (2 times)", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "complex_nested_with_mixed_types",
			input: `project
├── src
│   ├── main.go
│   └── utils
│       ├── helper.go
│       └── test (3 times)
├── docs
│   └── readme.md
└── config.json`,
			expected: Item{
				Name:         "project",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{
						Name:         "src",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{Name: "main.go", Index: 0, MissingIndex: true, Children: nil},
							{
								Name:         "utils",
								Index:        0,
								MissingIndex: true,
								Children: []Item{
									{Name: "helper.go", Index: 0, MissingIndex: true, Children: nil},
									{Name: "test (3 times)", Index: 0, MissingIndex: true, Children: nil},
								},
							},
						},
					},
					{
						Name:         "docs",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{Name: "readme.md", Index: 0, MissingIndex: true, Children: nil},
						},
					},
					{Name: "config.json", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name:    "empty_string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only_whitespace",
			input:   "   \n  \t  \n   ",
			wantErr: true,
		},
		{
			name: "whitespace_in_names",
			input: `root
├── file with spaces.txt
└── another file`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "file with spaces.txt", Index: 0, MissingIndex: true, Children: nil},
					{Name: "another file", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "special_characters",
			input: `root
├── file@#$%.txt
├── [brackets]
└── (parentheses)`,
			expected: Item{
				Name:         "root",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{Name: "file@#$%.txt", Index: 0, MissingIndex: true, Children: nil},
					{Name: "[brackets]", Index: 0, MissingIndex: true, Children: nil},
					{Name: "(parentheses)", Index: 0, MissingIndex: true, Children: nil},
				},
			},
		},
		{
			name: "numeric_prefixes",
			input: `1_root
├── 1_CallA
├── 2_CallB
└── 3_CallA`,
			expected: Item{
				Name:  "root",
				Index: 1,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallB", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, Children: nil},
				},
			},
		},
		{
			name: "mixed_numeric_and_regular",
			input: `project
├── 1_src
│   ├── main.go
│   └── 2_utils
│       ├── helper.go
│       └── 10_test
├── docs
└── 5_config`,
			expected: Item{
				Name:         "project",
				Index:        0,
				MissingIndex: true,
				Children: []Item{
					{
						Name:  "src",
						Index: 1,
						Children: []Item{
							{Name: "main.go", Index: 0, MissingIndex: true, Children: nil},
							{
								Name:  "utils",
								Index: 2,
								Children: []Item{
									{Name: "helper.go", Index: 0, MissingIndex: true, Children: nil},
									{Name: "test", Index: 10, Children: nil},
								},
							},
						},
					},
					{Name: "docs", Index: 0, MissingIndex: true, Children: nil},
					{Name: "config", Index: 5, Children: nil},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Parse(): %s", diff)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	// Test round-trip compatibility: TreeCollapsed -> Parse -> TreeCollapsed
	tests := []struct {
		name string
		dir  string
		opts TreeCollapseOptions
	}{
		{
			name: "basic_structure",
			dir:  "testdata/basic_structure",
			opts: TreeCollapseOptions{},
		},
		{
			name: "directories_only",
			dir:  "testdata/basic_structure",
			opts: TreeCollapseOptions{DirectoriesOnly: true},
		},
		{
			name: "nested_structure",
			dir:  "testdata/nested_structure",
			opts: TreeCollapseOptions{},
		},
		{
			name: "empty_dir",
			dir:  "testdata/empty_dir",
			opts: TreeCollapseOptions{},
		},
		{
			name: "single_file",
			dir:  "testdata/single_file",
			opts: TreeCollapseOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate tree string
			original, err := TreeCollapsed(tt.dir, tt.opts)
			if err != nil {
				t.Fatalf("TreeCollapsed() failed: %v", err)
			}

			// Parse it back
			parsed, err := Parse(original)
			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			// Convert parsed back to string representation
			regenerated := itemToString(parsed)

			// The regenerated tree should be functionally equivalent
			// (may not be character-for-character identical due to formatting differences)
			if !treeStructureEqual(original, regenerated) {
				t.Errorf("Round-trip failed:\nOriginal:\n%s\nRegenerated:\n%s", original, regenerated)
			}
		})
	}
}

func TestParseHelperFunctions(t *testing.T) {
	t.Run("getLineDepth", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int
		}{
			{"root", 0},
			{"├── child", 1},
			{"│   ├── grandchild", 2},
			{"│   │   └── great_grandchild", 3},
			{"└── last_child", 1},
			{"    └── indented", 2}, // 4 spaces = depth 1, + connector = depth 2
		}

		for _, tt := range tests {
			result := getLineDepth(tt.input)
			if result != tt.expected {
				t.Errorf("getLineDepth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("extractItemName", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"root", "root"},
			{"├── child", "child"},
			{"│   ├── grandchild", "grandchild"},
			{"│   │   └── great_grandchild", "great_grandchild"},
			{"└── last_child", "last_child"},
			{"    └── indented", "indented"},
			{"├── file with spaces.txt", "file with spaces.txt"},
			{"└── test (5 times)", "test (5 times)"},
		}

		for _, tt := range tests {
			result := extractItemName(tt.input)
			if result != tt.expected {
				t.Errorf("extractItemName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})
}

func TestParseNameWithIndex(t *testing.T) {
	tests := []struct {
		input         string
		expectedIndex int
		expectedName  string
	}{
		{"CallA", 0, "CallA"},
		{"1_CallA", 1, "CallA"},
		{"2_CallB", 2, "CallB"},
		{"10_test", 10, "test"},
		{"123_longname", 123, "longname"},
		{"0_zero", 0, "zero"},
		{"_invalid", 0, "_invalid"}, // No number before underscore
		{"invalid_", 0, "invalid_"}, // No number before underscore
		{"abc_def", 0, "abc_def"},   // No number before underscore
		{"1_", 1, ""},               // Number but empty name
		{"file.txt", 0, "file.txt"}, // No underscore
		{"1_2_3", 1, "2_3"},         // Multiple underscores, only first is prefix
		{"999_very_long_name", 999, "very_long_name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			index, hasIndex, name := parseNameWithIndexLocal(tt.input)
			_ = hasIndex
			if index != tt.expectedIndex {
				t.Errorf("parseNameWithIndex(%q) index = %d, want %d", tt.input, index, tt.expectedIndex)
			}
			if name != tt.expectedName {
				t.Errorf("parseNameWithIndex(%q) name = %q, want %q", tt.input, name, tt.expectedName)
			}
		})
	}
}

func TestParseEdgeCases(t *testing.T) {
	t.Run("malformed_input", func(t *testing.T) {
		malformedInputs := []string{
			"root\n      └── too_deep",             // Unexpected depth jump
			"root\n├── child1\n  └── wrong_indent", // Inconsistent indentation
		}

		for _, input := range malformedInputs {
			_, err := Parse(input)
			if err == nil {
				t.Errorf("Parse() should have failed on malformed input: %q", input)
			}
		}
	})

	t.Run("single_line_variations", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"root", "root"},
			{"  root  ", "root"},
			{"\troot\t", "root"},
			{"root\n", "root"},
		}

		for _, tt := range tests {
			result, err := Parse(tt.input)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", tt.input, err)
				continue
			}

			if result.Name != tt.expected {
				t.Errorf("Parse(%q).Name = %q, want %q", tt.input, result.Name, tt.expected)
			}
		}
	})
}

// Helper functions for testing

// itemToString converts an Item back to a tree string representation
func itemToString(item Item) string {
	var result strings.Builder
	result.WriteString(item.Name)
	if len(item.Children) > 0 {
		result.WriteString("\n")
		writeItemChildren(&result, item.Children, "")
	}
	return result.String()
}

// writeItemChildren recursively writes children with proper indentation
func writeItemChildren(result *strings.Builder, children []Item, prefix string) {
	for i, child := range children {
		isLast := i == len(children)-1

		var connector string
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}

		result.WriteString(prefix + connector + child.Name)

		if len(child.Children) > 0 {
			result.WriteString("\n")
			var newPrefix string
			if isLast {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}
			writeItemChildren(result, child.Children, newPrefix)
		}

		if i < len(children)-1 {
			result.WriteString("\n")
		}
	}
}

// treeStructureEqual checks if two tree strings represent the same structure
func treeStructureEqual(tree1, tree2 string) bool {
	// Normalize whitespace and compare line by line
	lines1 := strings.Split(strings.TrimSpace(tree1), "\n")
	lines2 := strings.Split(strings.TrimSpace(tree2), "\n")

	if len(lines1) != len(lines2) {
		return false
	}

	for i := range lines1 {
		name1 := extractItemName(lines1[i])
		name2 := extractItemName(lines2[i])
		depth1 := getLineDepth(lines1[i])
		depth2 := getLineDepth(lines2[i])

		if name1 != name2 || depth1 != depth2 {
			return false
		}
	}

	return true
}
