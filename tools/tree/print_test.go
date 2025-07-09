package tree

import (
	"strings"
	"testing"
)

func TestPrintItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []Item
		expected string
	}{
		{
			name:     "empty_items",
			items:    []Item{},
			expected: "",
		},
		{
			name: "single_item",
			items: []Item{
				{Name: "CallA", Index: 1},
			},
			expected: `1_CallA
`,
		},
		{
			name: "multiple_items",
			items: []Item{
				{Name: "CallA", Index: 1},
				{Name: "CallB", Index: 2},
				{Name: "CallC", Index: 3},
			},
			expected: `1_CallA
2_CallB
3_CallC
`,
		},
		{
			name: "items_with_children",
			items: []Item{
				{
					Name:  "Parent",
					Index: 1,
					Children: []Item{
						{Name: "Child1", Index: 0},
						{Name: "Child2", Index: 0},
					},
				},
				{Name: "Sibling", Index: 2},
			},
			expected: `1_Parent
├── 0_Child1
└── 0_Child2
2_Sibling
`,
		},
		{
			name: "nested_children",
			items: []Item{
				{
					Name:  "Root",
					Index: 1,
					Children: []Item{
						{
							Name:         "Level1",
							Index:        0,
							MissingIndex: true,
							Children: []Item{
								{Name: "Level2", Index: 0, MissingIndex: true},
							},
						},
					},
				},
			},
			expected: `1_Root
└── Level1
    └── Level2
`,
		},
		{
			name: "items_with_repetition",
			items: []Item{
				{Name: "CallA", Index: 1, SubsequentRepeated: 2},
				{Name: "CallB", Index: 2},
			},
			expected: `1_CallA (3 times)
2_CallB
`,
		},
		{
			name: "items_with_collapsed_pattern",
			items: []Item{
				{Name: "CallA", Index: 1, CollapsedPatternChildren: 5},
				{Name: "CallB", Index: 2},
			},
			expected: `1_CallA (...5 collapsed)
2_CallB
`,
		},
		{
			name: "complex_structure",
			items: []Item{
				{
					Name:  "Main",
					Index: 1,
					Children: []Item{
						{Name: "SubA", Index: 0, SubsequentRepeated: 1},
						{
							Name:  "SubB",
							Index: 0,
							Children: []Item{
								{Name: "DeepCall", Index: 0, CollapsedPatternChildren: 3},
							},
						},
					},
				},
			},
			expected: `1_Main
├── 0_SubA (2 times)
└── 0_SubB
    └── 0_DeepCall (...3 collapsed)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrintItems(tt.items)
			if result != tt.expected {
				t.Errorf("PrintItems() = \n%s\n, want \n%s\n", result, tt.expected)
			}
		})
	}
}

func TestPrintItem(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name: "simple_item",
			item: Item{Name: "CallA", Index: 1},
			expected: `1_CallA
`,
		},
		{
			name: "item_with_children",
			item: Item{
				Name:  "Parent",
				Index: 1,
				Children: []Item{
					{Name: "Child1", Index: 0, MissingIndex: true},
					{Name: "Child2", Index: 1},
				},
			},
			expected: `1_Parent
├── Child1
└── 1_Child2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrintItem(tt.item)
			if result != tt.expected {
				t.Errorf("PrintItem() = \n%s\n, want \n%s\n", result, tt.expected)
			}
		})
	}
}

func TestPrintItemsWithRoot(t *testing.T) {
	items := []Item{
		{Name: "CallA", Index: 1},
		{Name: "CallB", Index: 2},
	}
	expected := `root
1_CallA
2_CallB
`

	result := PrintItemsWithRoot("root", items)
	if result != expected {
		t.Errorf("PrintItemsWithRoot() = %q, want %q", result, expected)
	}
}

func TestPrintItemCompact(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "simple_item",
			item:     Item{Name: "CallA", Index: 1},
			expected: "1_CallA",
		},
		{
			name:     "item_with_repetition",
			item:     Item{Name: "CallA", Index: 1, SubsequentRepeated: 2},
			expected: "1_CallA (3 times)",
		},
		{
			name:     "item_with_collapsed_pattern",
			item:     Item{Name: "CallA", Index: 1, CollapsedPatternChildren: 5},
			expected: "1_CallA(...5)",
		},
		{
			name: "item_with_children",
			item: Item{
				Name:  "Parent",
				Index: 1,
				Children: []Item{
					{Name: "Child1", Index: 0, MissingIndex: true},
					{Name: "Child2", Index: 1},
				},
			},
			expected: "1_Parent[Child1, 1_Child2]",
		},
		{
			name: "nested_children",
			item: Item{
				Name:  "Root",
				Index: 1,
				Children: []Item{
					{
						Name:         "Level1",
						Index:        0,
						MissingIndex: true,
						Children: []Item{
							{Name: "Level2", Index: 0, MissingIndex: true},
						},
					},
				},
			},
			expected: "1_Root[Level1[Level2]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrintItemCompact(tt.item)
			if result != tt.expected {
				t.Errorf("PrintItemCompact() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrintItemsIntegration(t *testing.T) {
	// Test with a real example from the collapse tests
	items := []Item{
		{Name: "CallA", Index: 1},
		{Name: "CallB", Index: 2},
		{Name: "CallA", Index: 3},
		{Name: "CallC", Index: 4},
		{Name: "CallB", Index: 5},
		{Name: "CallA", Index: 6},
		{Name: "CallC", Index: 7, SubsequentRepeated: 2},
		{Name: "CallD", Index: 8, SubsequentRepeated: 2},
		{Name: "CallA", Index: 13},
		{Name: "CallB", Index: 14},
	}

	result := PrintItems(items)

	// Verify the output contains the expected structure
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines, got %d", len(lines))
	}

	// Check specific lines
	if lines[6] != "7_CallC (3 times)" {
		t.Errorf("Expected '7_CallC (3 times)', got %q", lines[6])
	}

	if lines[7] != "8_CallD (3 times)" {
		t.Errorf("Expected '8_CallD (3 times)', got %q", lines[7])
	}
}
