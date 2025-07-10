package tree

import (
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestCollapseNoCollapse(t *testing.T) {
	input := Item{
		Name:  "root",
		Index: 0,
		Children: []Item{
			{Name: "file1", Index: 0, Children: nil},
			{Name: "file2", Index: 0, Children: nil},
		},
	}
	opts := CollapseOptions{}
	expected := Item{
		Name:  "root",
		Index: 0,
		Children: []Item{
			{Name: "file1", Index: 0, Children: nil},
			{Name: "file2", Index: 0, Children: nil},
		},
	}

	result := Collapse(input, opts)
	if diff := assert.Diff(expected, result); diff != "" {
		t.Errorf("Collapse() result mismatch:\n%s", diff)
	}
}

func TestCollapseRepeated(t *testing.T) {
	tests := []struct {
		name     string
		input    Item
		expected Item
	}{
		{
			name: "consecutive_repeated",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallB", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, Children: nil},
					{Name: "CallA", Index: 4, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallB", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, SubsequentRepeated: 1, Children: nil},
				},
			},
		},
		{
			name: "nested_repeated",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 0, Children: []Item{
						{Name: "Child", Index: 1, Children: nil},
						{Name: "Child", Index: 2, Children: nil},
						{Name: "Child", Index: 3, Children: nil},
					}},
					{Name: "Parent2", Index: 0, Children: []Item{
						{Name: "Other", Index: 0, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 0, Children: []Item{
						{Name: "Child", Index: 1, SubsequentRepeated: 2, Children: nil},
					}},
					{Name: "Parent2", Index: 0, Children: []Item{
						{Name: "Other", Index: 0, Children: nil},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CollapseOptions{CollapseRepeated: true}
			result := Collapse(tt.input, opts)
			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Collapse() result mismatch:\n%s", diff)
			}
		})
	}
}

func TestCollapsePattern(t *testing.T) {
	input := Item{
		Name:  "root",
		Index: 0,
		Children: []Item{
			{Name: "CallA", Index: 1, Children: []Item{
				{Name: "SubCall1", Index: 0, Children: nil},
				{Name: "SubCall2", Index: 0, Children: nil},
			}},
			{Name: "CallB", Index: 2, Children: []Item{
				{Name: "SubCall3", Index: 0, Children: nil},
			}},
			{Name: "CallA", Index: 3, Children: []Item{
				{Name: "SubCall4", Index: 0, Children: nil},
				{Name: "SubCall5", Index: 0, Children: nil},
			}},
			{Name: "CallA", Index: 4, Children: []Item{
				{Name: "SubCall1", Index: 0, Children: nil},
				{Name: "SubCall2", Index: 0, Children: nil},
			}},
		},
	}
	opts := CollapseOptions{CollapsePattern: true}
	expected := Item{
		Name:  "root",
		Index: 0,
		Children: []Item{
			{Name: "CallA", Index: 1, Children: []Item{
				{Name: "SubCall1", Index: 0, Children: nil},
				{Name: "SubCall2", Index: 0, Children: nil},
			}},
			{Name: "CallB", Index: 2, Children: []Item{
				{Name: "SubCall3", Index: 0, Children: nil},
			}},
			{Name: "CallA", Index: 3, Children: []Item{
				{Name: "SubCall4", Index: 0, Children: nil},
				{Name: "SubCall5", Index: 0, Children: nil},
			}},
			{Name: "CallA", Index: 4, CollapsedPatternChildren: 2, Children: nil},
		},
	}

	result := Collapse(input, opts)
	if diff := assert.Diff(expected, result); diff != "" {
		t.Errorf("Collapse() result mismatch:\n%s", diff)
	}
}

func TestCollapseRepeatedAndPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    Item
		expected Item
	}{
		{
			name: "simple_all_same_pattern",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallA", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, Children: nil},
					{Name: "CallB", Index: 4, Children: nil},
					{Name: "CallA", Index: 5, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, SubsequentRepeated: 2, Children: nil},
					{Name: "CallB", Index: 4, Children: nil},
					{Name: "CallA", Index: 5, Children: nil},
				},
			},
		},
		{
			name: "mixed_patterns_with_children",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 2, Children: []Item{
						{Name: "SubCall2", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 3, Children: nil},
					{Name: "CallA", Index: 4, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 5, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 2, Children: []Item{
						{Name: "SubCall2", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 3, Children: nil},
					{Name: "CallA", Index: 4, SubsequentRepeated: 1, CollapsedPatternChildren: 1},
				},
			},
		},
		{
			name: "complex_nested_with_multiple_patterns",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
						{Name: "SubCall2", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 2, Children: []Item{
						{Name: "SubCall3", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 3, Children: []Item{
						{Name: "SubCall4", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 4, Children: []Item{
						{Name: "SubCall3", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 5, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
						{Name: "SubCall2", Index: 0, Children: nil},
					}},
					{Name: "CallC", Index: 6, Children: nil},
					{Name: "CallA", Index: 7, Children: []Item{
						{Name: "SubCall4", Index: 0, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, SubsequentRepeated: 0, Children: []Item{
						{Name: "SubCall1", Index: 0, Children: nil},
						{Name: "SubCall2", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 2, SubsequentRepeated: 0, Children: []Item{
						{Name: "SubCall3", Index: 0, Children: nil},
					}},
					{Name: "CallA", Index: 3, Children: []Item{
						{Name: "SubCall4", Index: 0, Children: nil},
					}},
					{Name: "CallB", Index: 4, CollapsedPatternChildren: 1},
					{Name: "CallA", Index: 5, CollapsedPatternChildren: 2},
					{Name: "CallC", Index: 6, Children: nil},
					{Name: "CallA", Index: 7, CollapsedPatternChildren: 1},
				},
			},
		},
		{
			name: "interleaved_patterns",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallB", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, Children: nil},
					{Name: "CallC", Index: 4, Children: nil},
					{Name: "CallB", Index: 5, Children: nil},
					{Name: "CallA", Index: 6, Children: nil},
					{Name: "CallC", Index: 7, Children: nil},
					{Name: "CallD", Index: 8, Children: nil},
					{Name: "CallC", Index: 9, Children: nil},
					{Name: "CallD", Index: 10, Children: nil},
					{Name: "CallC", Index: 11, Children: nil},
					{Name: "CallD", Index: 12, Children: nil},
					{Name: "CallA", Index: 13, Children: nil},
					{Name: "CallB", Index: 14, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "CallA", Index: 1, Children: nil},
					{Name: "CallB", Index: 2, Children: nil},
					{Name: "CallA", Index: 3, Children: nil},
					{Name: "CallC", Index: 4, Children: nil},
					{Name: "CallB", Index: 5, Children: nil},
					{Name: "CallA", Index: 6, Children: nil},
					{Name: "CallC", Index: 7, SubsequentRepeated: 2, Children: nil},
					{Name: "CallD", Index: 8, SubsequentRepeated: 2, Children: nil},
					{Name: "CallA", Index: 13, Children: nil},
					{Name: "CallB", Index: 14, Children: nil},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CollapseOptions{CollapseRepeated: true, CollapsePattern: true}
			result := Collapse(tt.input, opts)
			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Collapse() result mismatch:\n%s", diff)
			}
		})
	}
}

func TestAdvancedPatternCollapsing(t *testing.T) {
	tests := []struct {
		name        string
		input       Item
		expected    Item
		description string
	}{
		{
			name: "head_with_suffix_repeat",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Setup", Index: 1, Children: nil},
					{Name: "action", Index: 2, Children: nil},
					{Name: "action", Index: 3, Children: nil},
					{Name: "action", Index: 4, Children: nil},
					{Name: "action", Index: 5, Children: nil},
					{Name: "action", Index: 6, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Setup", Index: 1, Children: nil},
					{Name: "action", Index: 2, SubsequentRepeated: 4, Children: nil},
				},
			},
		},
		{
			name: "nested_head_with_suffix_repeat",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{
						Name: "sub",
						Children: []Item{
							{Name: "Setup", Index: 1, Children: nil},
							{Name: "action", Index: 2, Children: nil},
							{Name: "action", Index: 3, Children: nil},
							{Name: "action", Index: 4, Children: nil},
							{Name: "action", Index: 5, Children: nil},
							{Name: "action", Index: 6, Children: nil},
						},
					},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "sub", Index: 0, Children: []Item{
						{Name: "Setup", Index: 1, Children: nil},
						{Name: "action", Index: 2, SubsequentRepeated: 4, Children: nil},
					},
					},
				},
			},
		},
		{
			name: "prefer_higher_repeat_count",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Setup", Index: 1, Children: nil},
					{Name: "ProcessA", Index: 2, Children: nil},
					{Name: "ProcessB", Index: 3, Children: nil},
					{Name: "ProcessA", Index: 4, Children: nil},
					{Name: "ProcessB", Index: 5, Children: nil},
					{Name: "TaskX", Index: 6, Children: nil},
					{Name: "TaskY", Index: 7, Children: nil},
					{Name: "TaskX", Index: 8, Children: nil},
					{Name: "TaskY", Index: 9, Children: nil},
					{Name: "TaskX", Index: 10, Children: nil},
					{Name: "TaskY", Index: 11, Children: nil},
					{Name: "Cleanup", Index: 12, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Setup", Index: 1, Children: nil},
					{Name: "ProcessA", Index: 2, Children: nil},
					{Name: "ProcessB", Index: 3, Children: nil},
					{Name: "ProcessA", Index: 4, Children: nil},
					{Name: "ProcessB", Index: 5, Children: nil},
					{Name: "TaskX", Index: 6, SubsequentRepeated: 2, Children: nil},
					{Name: "TaskY", Index: 7, SubsequentRepeated: 2, Children: nil},
					{Name: "Cleanup", Index: 12, Children: nil},
				},
			},
			description: "Should prefer TaskX-TaskY pattern (3 repeats) over ProcessA-ProcessB pattern (2 repeats)",
		},
		{
			name: "prefer_longer_pattern_same_repeat_count",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Init", Index: 1, Children: nil},
					{Name: "StepA", Index: 2, Children: nil},
					{Name: "StepB", Index: 3, Children: nil},
					{Name: "StepC", Index: 4, Children: nil},
					{Name: "StepA", Index: 5, Children: nil},
					{Name: "StepB", Index: 6, Children: nil},
					{Name: "StepC", Index: 7, Children: nil},
					{Name: "TaskX", Index: 8, Children: nil},
					{Name: "TaskY", Index: 9, Children: nil},
					{Name: "TaskX", Index: 10, Children: nil},
					{Name: "TaskY", Index: 11, Children: nil},
					{Name: "Final", Index: 12, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Init", Index: 1, Children: nil},
					{Name: "StepA", Index: 2, SubsequentRepeated: 1, Children: nil},
					{Name: "StepB", Index: 3, SubsequentRepeated: 1, Children: nil},
					{Name: "StepC", Index: 4, SubsequentRepeated: 1, Children: nil},
					{Name: "TaskX", Index: 8, Children: nil},
					{Name: "TaskY", Index: 9, Children: nil},
					{Name: "TaskX", Index: 10, Children: nil},
					{Name: "TaskY", Index: 11, Children: nil},
					{Name: "Final", Index: 12, Children: nil},
				},
			},
			description: "Should prefer StepA-StepB-StepC pattern (length 3) over TaskX-TaskY pattern (length 2) when both repeat twice",
		},
		{
			name: "prefer_later_start_position",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Begin", Index: 1, Children: nil},
					{Name: "OpA", Index: 2, Children: nil},
					{Name: "OpB", Index: 3, Children: nil},
					{Name: "OpA", Index: 4, Children: nil},
					{Name: "OpB", Index: 5, Children: nil},
					{Name: "Middle", Index: 6, Children: nil},
					{Name: "OpX", Index: 7, Children: nil},
					{Name: "OpY", Index: 8, Children: nil},
					{Name: "OpX", Index: 9, Children: nil},
					{Name: "OpY", Index: 10, Children: nil},
					{Name: "End", Index: 11, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Begin", Index: 1, Children: nil},
					{Name: "OpA", Index: 2, Children: nil},
					{Name: "OpB", Index: 3, Children: nil},
					{Name: "OpA", Index: 4, Children: nil},
					{Name: "OpB", Index: 5, Children: nil},
					{Name: "Middle", Index: 6, Children: nil},
					{Name: "OpX", Index: 7, SubsequentRepeated: 1, Children: nil},
					{Name: "OpY", Index: 8, SubsequentRepeated: 1, Children: nil},
					{Name: "End", Index: 11, Children: nil},
				},
			},
			description: "Should prefer OpX-OpY pattern (starts at 7) over OpA-OpB pattern (starts at 2) when both have same length and repeat count",
		},
		{
			name: "complex_nested_patterns",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Init", Index: 1, Children: nil},
					{Name: "BatchA", Index: 2, Children: []Item{
						{Name: "SubTask1", Index: 0, Children: nil},
						{Name: "SubTask2", Index: 0, Children: nil},
					}},
					{Name: "BatchB", Index: 3, Children: []Item{
						{Name: "SubTask3", Index: 0, Children: nil},
					}},
					{Name: "BatchA", Index: 4, Children: []Item{
						{Name: "SubTask1", Index: 0, Children: nil},
						{Name: "SubTask2", Index: 0, Children: nil},
					}},
					{Name: "BatchB", Index: 5, Children: []Item{
						{Name: "SubTask3", Index: 0, Children: nil},
					}},
					{Name: "BatchA", Index: 6, Children: []Item{
						{Name: "SubTask1", Index: 0, Children: nil},
						{Name: "SubTask2", Index: 0, Children: nil},
					}},
					{Name: "BatchB", Index: 7, Children: []Item{
						{Name: "SubTask3", Index: 0, Children: nil},
					}},
					{Name: "Final", Index: 8, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Init", Index: 1, Children: nil},
					{Name: "BatchA", Index: 2, SubsequentRepeated: 2, Children: []Item{
						{Name: "SubTask1", Index: 0, Children: nil},
						{Name: "SubTask2", Index: 0, Children: nil},
					}},
					{Name: "BatchB", Index: 3, SubsequentRepeated: 2, Children: []Item{
						{Name: "SubTask3", Index: 0, Children: nil},
					}},
					{Name: "Final", Index: 8, Children: nil},
				},
			},
			description: "Should handle complex nested patterns with different child structures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CollapseOptions{CollapseRepeated: true, CollapsePattern: true}
			result := Collapse(tt.input, opts)
			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Collapse() for %s result mismatch:\n%s\nDescription: %s", tt.name, diff, tt.description)
			}
		})
	}
}

func TestCollapseLeaf(t *testing.T) {
	tests := []struct {
		name     string
		input    Item
		expected Item
	}{
		{
			name: "simple_leaf_collapse",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "LeafB", Index: 2, Children: nil},
					{Name: "LeafA", Index: 3, Children: nil},
					{Name: "LeafC", Index: 4, Children: nil},
					{Name: "LeafA", Index: 5, Children: nil},
				},
			},
			expected: Item{
				Name:                  "root",
				Index:                 0,
				CollapsedLeafChildren: 2,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "LeafB", Index: 2, Children: nil},
					{Name: "LeafC", Index: 4, Children: nil},
				},
			},
		},
		{
			name: "nested_leaf_collapse",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
						{Name: "Child2", Index: 2, Children: nil},
						{Name: "Child1", Index: 3, Children: nil},
					}},
					{Name: "Parent2", Index: 2, Children: []Item{
						{Name: "Child3", Index: 1, Children: nil},
						{Name: "Child4", Index: 2, Children: nil},
						{Name: "Child3", Index: 3, Children: nil},
						{Name: "Child4", Index: 4, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 1, CollapsedLeafChildren: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
						{Name: "Child2", Index: 2, Children: nil},
					}},
					{Name: "Parent2", Index: 2, CollapsedLeafChildren: 2, Children: []Item{
						{Name: "Child3", Index: 1, Children: nil},
						{Name: "Child4", Index: 2, Children: nil},
					}},
				},
			},
		},
		{
			name: "mixed_leaf_and_non_leaf",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "Parent1", Index: 2, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
					}},
					{Name: "LeafA", Index: 3, Children: nil},
					{Name: "LeafB", Index: 4, Children: nil},
					{Name: "Parent2", Index: 5, Children: []Item{
						{Name: "Child2", Index: 1, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:                  "root",
				Index:                 0,
				CollapsedLeafChildren: 1,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "Parent1", Index: 2, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
					}},
					{Name: "LeafB", Index: 4, Children: nil},
					{Name: "Parent2", Index: 5, Children: []Item{
						{Name: "Child2", Index: 1, Children: nil},
					}},
				},
			},
		},
		{
			name: "all_children_collapsed_treated_as_leaf",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
						{Name: "Child1", Index: 2, Children: nil},
					}},
					{Name: "Parent2", Index: 2, Children: []Item{
						{Name: "Child2", Index: 1, Children: nil},
					}},
					{Name: "Parent1", Index: 3, Children: []Item{
						{Name: "Child1", Index: 3, Children: nil},
						{Name: "Child1", Index: 4, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:                  "root",
				Index:                 0,
				CollapsedLeafChildren: 1,
				Children: []Item{
					{Name: "Parent1", Index: 1, CollapsedLeafChildren: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
					}},
					{Name: "Parent2", Index: 2, Children: []Item{
						{Name: "Child2", Index: 1, Children: nil},
					}},
				},
			},
		},
		{
			name: "deep_nested_leaf_collapse",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Level1", Index: 1, Children: []Item{
						{Name: "Level2", Index: 1, Children: []Item{
							{Name: "DeepLeaf1", Index: 1, Children: nil},
							{Name: "DeepLeaf2", Index: 2, Children: nil},
							{Name: "DeepLeaf1", Index: 3, Children: nil},
						}},
					}},
					{Name: "Level1Other", Index: 2, Children: []Item{
						{Name: "OtherLeaf", Index: 1, Children: nil},
					}},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Level1", Index: 1, Children: []Item{
						{Name: "Level2", Index: 1, CollapsedLeafChildren: 1, Children: []Item{
							{Name: "DeepLeaf1", Index: 1, Children: nil},
							{Name: "DeepLeaf2", Index: 2, Children: nil},
						}},
					}},
					{Name: "Level1Other", Index: 2, Children: []Item{
						{Name: "OtherLeaf", Index: 1, Children: nil},
					}},
				},
			},
		},
		{
			name: "no_duplicates",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "LeafB", Index: 2, Children: nil},
					{Name: "LeafC", Index: 3, Children: nil},
				},
			},
			expected: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "LeafB", Index: 2, Children: nil},
					{Name: "LeafC", Index: 3, Children: nil},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CollapseOptions{CollapseLeaf: true}
			result := Collapse(tt.input, opts)
			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Collapse() result mismatch:\n%s", diff)
			}
		})
	}
}

func TestCollapseLeafWithOtherOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    Item
		opts     CollapseOptions
		expected Item
	}{
		{
			name: "leaf_with_repeated",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "LeafA", Index: 1, Children: nil},
					{Name: "LeafA", Index: 2, Children: nil},
					{Name: "LeafA", Index: 3, Children: nil},
					{Name: "LeafB", Index: 4, Children: nil},
					{Name: "LeafA", Index: 5, Children: nil},
				},
			},
			opts: CollapseOptions{CollapseLeaf: true, CollapseRepeated: true},
			expected: Item{
				Name:                  "root",
				Index:                 0,
				CollapsedLeafChildren: 1,
				Children: []Item{
					{Name: "LeafA", Index: 1, SubsequentRepeated: 2, Children: nil},
					{Name: "LeafB", Index: 4, Children: nil},
				},
			},
		},
		{
			name: "leaf_with_pattern",
			input: Item{
				Name:  "root",
				Index: 0,
				Children: []Item{
					{Name: "Parent1", Index: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
						{Name: "Child2", Index: 2, Children: nil},
					}},
					{Name: "Parent2", Index: 2, Children: []Item{
						{Name: "Child3", Index: 1, Children: nil},
					}},
					{Name: "Parent1", Index: 3, Children: []Item{
						{Name: "Child1", Index: 3, Children: nil},
						{Name: "Child2", Index: 4, Children: nil},
					}},
					{Name: "LeafA", Index: 4, Children: nil},
					{Name: "LeafA", Index: 5, Children: nil},
				},
			},
			opts: CollapseOptions{CollapseLeaf: true, CollapsePattern: true},
			expected: Item{
				Name:                  "root",
				Index:                 0,
				CollapsedLeafChildren: 1,
				Children: []Item{
					{Name: "Parent1", Index: 1, Children: []Item{
						{Name: "Child1", Index: 1, Children: nil},
						{Name: "Child2", Index: 2, Children: nil},
					}},
					{Name: "Parent2", Index: 2, Children: []Item{
						{Name: "Child3", Index: 1, Children: nil},
					}},
					{Name: "Parent1", Index: 3, CollapsedPatternChildren: 2, Children: nil},
					{Name: "LeafA", Index: 4, Children: nil},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Collapse(tt.input, tt.opts)
			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Errorf("Collapse() result mismatch:\n%s", diff)
			}
		})
	}
}
