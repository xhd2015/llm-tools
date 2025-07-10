package tree

import (
	"fmt"
	"strings"
)

func StringsToItem(list []string) Item {
	if len(list) == 0 {
		return Item{}
	}
	root := Item{
		Name: list[0],
	}

	if len(list) > 1 {
		child := StringsToItem(list[1:])
		root.Children = append(root.Children, child)
	}
	return root
}

// PrintItems converts a slice of Items to a tree-like string representation
func PrintItems(items []Item) string {
	if len(items) == 0 {
		return ""
	}

	var result strings.Builder
	for _, item := range items {
		// Each item at the root level is treated as its own tree
		printItem(item, "", true, &result)
	}
	return result.String()
}

// PrintItem converts a single Item to a tree-like string representation
func PrintItem(item Item) string {
	var result strings.Builder
	printItem(item, "", true, &result)
	return result.String()
}

func addCollapsedInfo(name string, item Item) string {
	totalCollapsed := item.CollapsedPatternChildren + item.CollapsedLeafChildren
	if totalCollapsed > 0 {
		word := "collapsed"
		if len(item.Children) > 0 {
			word = "omitted"
		}
		name = fmt.Sprintf("%s (...%d %s)", name, totalCollapsed, word)
	}
	return name
}

// printItem recursively prints an item and its children with proper tree formatting
func printItem(item Item, prefix string, isLast bool, result *strings.Builder) {
	// Build the item name with additional info
	name := getName(item)

	if item.Star {
		name = "*" + name
	}

	// Add repetition indicator
	if item.SubsequentRepeated > 0 {
		name = fmt.Sprintf("%s (%d times)", name, item.SubsequentRepeated+1)
	}

	name = addCollapsedInfo(name, item)

	// Determine the connector symbol
	var connector string
	if prefix == "" {
		// Root level item
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// Write the current item
	result.WriteString(prefix + connector + name + "\n")

	// Print children if any
	if len(item.Children) > 0 {
		var childPrefix string
		if prefix == "" {
			// Root level item - children start with no prefix but will get connectors
			childPrefix = ""
		} else if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}

		for i, child := range item.Children {
			isLastChild := i == len(item.Children)-1
			// For children of root level items, we need to force them to have connectors
			if prefix == "" {
				printItemWithConnector(child, childPrefix, isLastChild, result)
			} else {
				printItem(child, childPrefix, isLastChild, result)
			}
		}
	}
}

// printItemWithConnector is like printItem but forces connectors even at root level
func printItemWithConnector(item Item, prefix string, isLast bool, result *strings.Builder) {
	// Build the item name with additional info
	name := getName(item)

	// Add repetition indicator
	if item.SubsequentRepeated > 0 {
		name = fmt.Sprintf("%s (%d times)", name, item.SubsequentRepeated+1)
	}

	name = addCollapsedInfo(name, item)

	// Always use connectors
	var connector string
	if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// Write the current item
	result.WriteString(prefix + connector + name + "\n")

	// Print children if any
	if len(item.Children) > 0 {
		var childPrefix string
		if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}

		for i, child := range item.Children {
			isLastChild := i == len(item.Children)-1
			printItemWithConnector(child, childPrefix, isLastChild, result)
		}
	}
}

// PrintItemsWithRoot converts a slice of Items to a tree-like string representation with a root name
func PrintItemsWithRoot(rootName string, items []Item) string {
	var result strings.Builder
	result.WriteString(rootName + "\n")

	for i, item := range items {
		isLast := i == len(items)-1
		printItem(item, "", isLast, &result)
	}

	return result.String()
}

func getName(item Item) string {
	if item.MissingIndex {
		return item.Name
	}
	return fmt.Sprintf("%d_%s", item.Index, item.Name)
}

// PrintItemCompact converts an Item to a compact string representation (single line)
func PrintItemCompact(item Item) string {
	name := getName(item)

	if item.SubsequentRepeated > 0 {
		name = fmt.Sprintf("%s (%d times)", name, item.SubsequentRepeated+1)
	}

	name = addCollapsedInfo(name, item)

	if len(item.Children) > 0 {
		var childNames []string
		for _, child := range item.Children {
			childNames = append(childNames, PrintItemCompact(child))
		}
		name = fmt.Sprintf("%s[%s]", name, strings.Join(childNames, ", "))
	}

	return name
}
