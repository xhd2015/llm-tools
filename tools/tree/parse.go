package tree

import (
	"fmt"
	"regexp"
	"strings"
)

type Item struct {
	Name string
	// Index extracted from numeric prefix (e.g., "1_CallA" -> Index=1, Name="CallA")
	MissingIndex bool
	Index        int
	Dir          bool
	// Number of additional consecutive repetitions (0 = no repetition, 1 = one extra occurrence, etc.)
	SubsequentRepeated int
	// Number of children that were collapsed due to pattern matching (0 = no pattern collapse)
	CollapsedPatternChildren int
	Children                 []Item
}

func Parse(tree string) (Item, error) {
	if tree == "" {
		return Item{}, fmt.Errorf("empty tree string")
	}

	lines := strings.Split(tree, "\n")
	if len(lines) == 0 {
		return Item{}, fmt.Errorf("no lines in tree string")
	}

	// Remove empty lines
	var filteredLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	if len(filteredLines) == 0 {
		return Item{}, fmt.Errorf("no non-empty lines in tree string")
	}

	// First line is the root
	rootIndex, hasIndex, rootName := parseNameWithIndexLocal(strings.TrimSpace(filteredLines[0]))
	root := Item{Name: rootName, Index: rootIndex, MissingIndex: !hasIndex}

	if len(filteredLines) == 1 {
		return root, nil
	}

	// Validate tree structure before parsing
	for i, line := range filteredLines[1:] {
		lineDepth := getLineDepth(line)
		if lineDepth == 0 {
			return Item{}, fmt.Errorf("invalid tree structure: line %d has depth 0 but should be a child", i+1)
		}

		// Check for unexpected depth jumps
		if i == 0 && lineDepth != 1 {
			return Item{}, fmt.Errorf("invalid tree structure: first child line has depth %d, expected 1", lineDepth)
		}

		if i > 0 {
			prevLineDepth := getLineDepth(filteredLines[i])
			if lineDepth > prevLineDepth+1 {
				return Item{}, fmt.Errorf("invalid tree structure: unexpected depth jump from %d to %d at line %q",
					prevLineDepth, lineDepth, line)
			}
		}

		// Check for inconsistent indentation patterns
		if err := validateIndentation(line); err != nil {
			return Item{}, fmt.Errorf("invalid tree structure: %w", err)
		}
	}

	// Parse the tree structure
	var err error
	root.Children, err = parseChildren(filteredLines[1:], 1)
	if err != nil {
		return Item{}, fmt.Errorf("failed to parse children: %w", err)
	}

	return root, nil
}

// parseChildren recursively parses child items from lines starting at the given depth
func parseChildren(lines []string, expectedDepth int) ([]Item, error) {
	var children []Item
	i := 0

	for i < len(lines) {
		line := lines[i]
		lineDepth := getLineDepth(line)

		// If this line is at a lower depth, we're done with this level
		if lineDepth < expectedDepth {
			break
		}

		// Validate that we don't have unexpected depth jumps
		if lineDepth > expectedDepth+1 {
			return nil, fmt.Errorf("unexpected depth jump at line %q: expected depth %d or %d, got %d",
				line, expectedDepth, expectedDepth+1, lineDepth)
		}

		// Skip lines that are deeper than expected - they'll be handled recursively
		if lineDepth > expectedDepth {
			i++
			continue
		}

		// This line is at the expected depth, process it
		rawName := extractItemName(line)
		index, hasIndex, name := parseNameWithIndexLocal(rawName)
		item := Item{Name: name, Index: index, MissingIndex: !hasIndex}

		// Collect all subsequent lines that are children of this item
		var childLines []string
		j := i + 1
		for j < len(lines) {
			childLine := lines[j]
			childDepth := getLineDepth(childLine)

			// If we hit a line at the same or lower depth, we're done with this item's children
			if childDepth <= expectedDepth {
				break
			}

			childLines = append(childLines, childLine)
			j++
		}

		// If there are children, parse them recursively
		if len(childLines) > 0 {
			var err error
			item.Children, err = parseChildren(childLines, expectedDepth+1)
			if err != nil {
				return nil, err
			}
		}

		children = append(children, item)
		i = j // Move to the next item at this level
	}

	return children, nil
}

// getLineDepth calculates the depth of a tree line based on its indentation
func getLineDepth(line string) int {
	// Root level has no tree symbols
	if !strings.Contains(line, "├") && !strings.Contains(line, "└") && !strings.Contains(line, "│") {
		return 0
	}

	depth := 0

	// Count the total leading spaces and pipes to determine depth
	// Tree structure patterns:
	// "├── " or "└── " = depth 1
	// "│   ├── " or "│   └── " = depth 2
	// "│   │   ├── " or "│   │   └── " = depth 3
	// "│       ├── " or "│       └── " = depth 3 (alternative format)

	// First, count all "│   " patterns
	index := 0
	for {
		pipeIndex := strings.Index(line[index:], "│   ")
		if pipeIndex == -1 {
			break
		}
		depth++
		index += pipeIndex + 4
	}

	// Then check if there are additional spaces after the last "│   "
	// This handles cases like "│       ├── " where there are 4 extra spaces
	remainingLine := line[index:]
	spaceCount := 0
	for _, char := range remainingLine {
		if char == ' ' {
			spaceCount++
		} else {
			break
		}
	}

	// Each group of 4 spaces represents one additional level
	depth += spaceCount / 4

	// Add 1 for the final connector (├── or └──)
	if strings.Contains(line, "├") || strings.Contains(line, "└") {
		depth++
	}

	return depth
}

// validateIndentation checks if a line has consistent indentation patterns
func validateIndentation(line string) error {
	// Check for invalid indentation patterns
	// Valid patterns: "├── ", "└── ", "│   ", and proper multiples of 4 spaces

	// Count leading spaces
	spaceCount := 0
	for _, char := range line {
		if char == ' ' {
			spaceCount++
		} else {
			break
		}
	}

	// If we have connectors (├ or └), validate the indentation
	if strings.Contains(line, "├") || strings.Contains(line, "└") {
		// For lines with connectors, spaces should be multiples of 4
		// OR it should be a direct child (no leading spaces, just the connector)
		if spaceCount > 0 && spaceCount%4 != 0 {
			return fmt.Errorf("inconsistent indentation: line %q has %d spaces, expected multiple of 4", line, spaceCount)
		}
	}

	return nil
}

// extractItemName extracts the actual item name from a tree line
func extractItemName(line string) string {
	// Regular expression to match tree symbols and extract the name
	re := regexp.MustCompile(`[│├└─\s]*(.+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(line)
}
