package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type TreeCollapseOptions struct {
	// IncludePatterns are regex patterns to include file/directory names (whitelist)
	IncludePatterns []string
	// ExcludePatterns are regex patterns to exclude file/directory names (blacklist)
	ExcludePatterns []string
	// DirectoriesOnly shows only directories, not files
	DirectoriesOnly bool
	CollapsedDirs   []string
}

// TreeOptions contains options for tree generation
type TreeOptions struct {
	// IncludePatterns are regex patterns to include file/directory names (whitelist)
	IncludePatterns []string
	// ExcludePatterns are regex patterns to exclude file/directory names (blacklist)
	ExcludePatterns []string
	// DirectoriesOnly shows only directories, not files
	DirectoriesOnly bool
	// CollapseRepeated collapses repeated entries into a single entry with count
	CollapseRepeated bool
	// CollapsePattern collapses duplicate patterns, showing full structure only for first appearance
	CollapsePattern bool
	// CollapseLeaf collapses duplicate leaf items by adding to parent's CollapsedPatternChildren
	CollapseLeaf  bool
	CollapsedDirs []string
	// Depth is the maximum depth of the directory tree to traverse (default 8)
	Depth int
	// MaxEntriesPerDir is the maximum number of entries to display in each directory (default 48)
	MaxEntriesPerDir int
	// ExpandDirs are directories that should be expanded with additional depth
	ExpandDirs []string
}

func Tree(dir string, opts TreeOptions) (string, error) {
	return traverseTree(dir, opts)
}

func TreeCollapsed(dir string, opts TreeCollapseOptions) (string, error) {
	return traverseTree(dir, TreeOptions{
		IncludePatterns:  opts.IncludePatterns,
		ExcludePatterns:  opts.ExcludePatterns,
		DirectoriesOnly:  opts.DirectoriesOnly,
		CollapseRepeated: true,
		CollapsePattern:  true,
		CollapseLeaf:     true,
		CollapsedDirs:    opts.CollapsedDirs,
		Depth:            DEFAULT_MAX_DEPTH,
		MaxEntriesPerDir: DEFAULT_MAX_ENTRIES_PER_DIR,
		ExpandDirs:       opts.CollapsedDirs,
	})
}

// traverseTree builds a tree using Item structures and applies collapsing
func traverseTree(dir string, opts TreeOptions) (string, error) {
	rootItem, err := traverseTreeItem(dir, opts)
	if err != nil {
		return "", err
	}
	return PrintItem(rootItem), nil
}

func traverseTreeItem(dir string, opts TreeOptions) (Item, error) {
	// Get the base directory name
	baseName := filepath.Base(dir)

	// Build the tree structure as Items
	rootItem, err := buildTreeAsItem(dir, opts)
	if err != nil {
		return Item{}, err
	}
	index, hasIndex, name := parseNameWithIndexLocal(baseName)
	// Set the root name
	rootItem.Name = name
	rootItem.MissingIndex = !hasIndex
	rootItem.Index = index

	sortItems(rootItem.Children)

	// Apply collapsing if enabled using the modern Collapse function
	if opts.CollapseRepeated || opts.CollapsePattern || opts.CollapseLeaf {
		rootItem = Collapse(rootItem, CollapseOptions{
			CollapseRepeated: opts.CollapseRepeated,
			CollapsePattern:  opts.CollapsePattern,
			CollapseLeaf:     opts.CollapseLeaf,
			CollapsedDirs:    opts.CollapsedDirs,
		})
	}

	// Apply max entries limit after collapse recursively
	if opts.MaxEntriesPerDir > 0 {
		applyMaxEntriesLimit(&rootItem, opts.MaxEntriesPerDir)
	}

	// Convert to string representation using modern PrintItem
	return rootItem, nil
}

func matchSuffix(prefix []string, name string, path []string) bool {
	n := len(path)
	if n == 0 {
		return true
	}
	if path[n-1] != name {
		return false
	}
	m := len(prefix)
	if n-1 > m {
		return false
	}
	for i := n - 2; i >= 0; i-- {
		if path[i] != prefix[m-1] {
			return false
		}
		m--
	}
	return true
}

func findPathInTree(prefix []string, item Item, path []string) [][]string {
	var result [][]string

	itemOrigName := getName(item)
	if matchSuffix(prefix, itemOrigName, path) {
		foundPath := append(cloneList(prefix), itemOrigName)
		result = append(result, foundPath)
	}

	nextPrefix := append(prefix, itemOrigName)
	for _, child := range item.Children {
		foundPath := findPathInTree(nextPrefix, child, path)
		result = append(result, foundPath...)
	}
	return result
}
func cloneList(list []string) []string {
	clone := make([]string, len(list))
	copy(clone, list)
	return clone
}

// applyMaxEntriesLimit recursively applies the max entries limit to all items
func applyMaxEntriesLimit(item *Item, maxEntries int) {
	if len(item.Children) > maxEntries {
		item.Children = item.Children[:maxEntries]
	}

	for i := range item.Children {
		applyMaxEntriesLimit(&item.Children[i], maxEntries)
	}
}

func sortItems(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Dir != items[j].Dir {
			return items[i].Dir
		}
		if items[i].MissingIndex != items[j].MissingIndex {
			return items[j].MissingIndex
		}

		d := items[i].Index - items[j].Index
		if d != 0 {
			return d < 0
		}
		return items[i].Name < items[j].Name
	})
	for _, item := range items {
		if len(item.Children) > 0 {
			sortItems(item.Children)
		}
	}
}

// buildTreeAsItem recursively builds a tree structure as Items
func buildTreeAsItem(dir string, opts TreeOptions) (Item, error) {
	// Compile regex patterns if provided
	var includePatterns []*regexp.Regexp
	var excludePatterns []*regexp.Regexp

	for _, pattern := range opts.IncludePatterns {
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				return Item{}, fmt.Errorf("invalid include pattern '%s': %v", pattern, err)
			}
			includePatterns = append(includePatterns, regex)
		}
	}

	for _, pattern := range opts.ExcludePatterns {
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				return Item{}, fmt.Errorf("invalid exclude pattern '%s': %v", pattern, err)
			}
			excludePatterns = append(excludePatterns, regex)
		}
	}

	// Set default values if not specified
	if opts.Depth == 0 {
		opts.Depth = DEFAULT_MAX_DEPTH
	}
	if opts.MaxEntriesPerDir == 0 {
		opts.MaxEntriesPerDir = DEFAULT_MAX_ENTRIES_PER_DIR
	}

	return buildTreeAsItemRecursive(dir, opts, includePatterns, excludePatterns, 0)
}

// buildTreeAsItemRecursive recursively builds tree structure as Items
func buildTreeAsItemRecursive(dir string, opts TreeOptions, includePatterns []*regexp.Regexp, excludePatterns []*regexp.Regexp, currentDepth int) (Item, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Item{}, err
	}

	// Filter entries based on options
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		// Handle file filtering - use DirectoriesOnly if set, otherwise use IncludeFiles
		if opts.DirectoriesOnly && !entry.IsDir() {
			continue
		}

		// Apply include patterns first (whitelist)
		if len(includePatterns) > 0 {
			matched := false
			for _, pattern := range includePatterns {
				if pattern.MatchString(entry.Name()) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Apply exclude patterns (blacklist)
		if len(excludePatterns) > 0 {
			excluded := false
			for _, pattern := range excludePatterns {
				if pattern.MatchString(entry.Name()) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		filteredEntries = append(filteredEntries, entry)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(filteredEntries, func(i, j int) bool {
		if filteredEntries[i].IsDir() && !filteredEntries[j].IsDir() {
			return true
		}
		if !filteredEntries[i].IsDir() && filteredEntries[j].IsDir() {
			return false
		}
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})

	// Create the root item
	rootItem := Item{
		Name: filepath.Base(dir),
		Dir:  true,
	}

	// Process children
	var children []Item
	for _, entry := range filteredEntries {
		entryName := entry.Name()

		// Parse name with index
		index, hasIndex, name := parseNameWithIndexLocal(entryName)

		isDir := entry.IsDir()
		child := Item{
			Name:         name,
			Index:        index,
			MissingIndex: !hasIndex,
			Dir:          isDir,
		}

		// If it's a directory, recursively process it
		if isDir {
			subDir := filepath.Join(dir, entryName)

			// Check depth limit (but allow expansion for directories in ExpandDirs)
			shouldExpand := false
			if opts.Depth > 0 {
				// Check if this directory should be expanded
				for _, expandDir := range opts.ExpandDirs {
					if strings.Contains(subDir, expandDir) || strings.Contains(expandDir, subDir) {
						shouldExpand = true
						break
					}
				}

				if currentDepth+1 >= opts.Depth && !shouldExpand {
					// Don't recurse further, but mark as directory
					child.Dir = true
				} else {
					subItem, err := buildTreeAsItemRecursive(subDir, opts, includePatterns, excludePatterns, currentDepth+1)
					if err != nil {
						return Item{}, err
					}
					// Use the children from the recursive call
					child.Children = subItem.Children
				}
			} else {
				// No depth limit, always recurse
				subItem, err := buildTreeAsItemRecursive(subDir, opts, includePatterns, excludePatterns, currentDepth+1)
				if err != nil {
					return Item{}, err
				}
				// Use the children from the recursive call
				child.Children = subItem.Children
			}
		}

		children = append(children, child)
	}

	rootItem.Children = children
	return rootItem, nil
}

// parseNameWithIndexLocal parses a name that may have a numeric prefix
// e.g., "1_CallA" -> index=1, name="CallA"
// e.g., "CallA" -> index=0, name="CallA"
func parseNameWithIndexLocal(rawName string) (index int, hasIndex bool, name string) {
	// Check if the name starts with a number followed by underscore
	re := regexp.MustCompile(`^(\d+)_(.*)$`)
	matches := re.FindStringSubmatch(rawName)

	if len(matches) == 3 {
		// Found numeric prefix
		if idx, err := strconv.Atoi(matches[1]); err == nil {
			return idx, true, matches[2]
		}
	}

	// No numeric prefix found
	return 0, false, rawName
}

func doTree(dir string, opts TreeOptions) (string, error) {
	var result strings.Builder

	// Get the base directory name
	baseName := filepath.Base(dir)
	result.WriteString(baseName + "\n")

	// Compile regex patterns if provided
	var includePatterns []*regexp.Regexp
	var excludePatterns []*regexp.Regexp

	for _, pattern := range opts.IncludePatterns {
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid include pattern '%s': %v", pattern, err)
			}
			includePatterns = append(includePatterns, regex)
		}
	}

	for _, pattern := range opts.ExcludePatterns {
		if pattern != "" {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid exclude pattern '%s': %v", pattern, err)
			}
			excludePatterns = append(excludePatterns, regex)
		}
	}

	// Track seen patterns for CollapsePattern option
	seenPatterns := make(map[string]bool)

	// Build the tree structure
	err := buildTreeWithOptions(dir, "", true, &result, opts, includePatterns, excludePatterns, seenPatterns)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func buildTreeWithOptions(dir string, prefix string, isLast bool, result *strings.Builder, opts TreeOptions, includePatterns []*regexp.Regexp, excludePatterns []*regexp.Regexp, seenPatterns map[string]bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter entries based on options
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		// Skip files if DirectoriesOnly is true
		if opts.DirectoriesOnly && !entry.IsDir() {
			continue
		}

		// Apply include patterns first (whitelist)
		if len(includePatterns) > 0 {
			matched := false
			for _, pattern := range includePatterns {
				if pattern.MatchString(entry.Name()) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Apply exclude patterns (blacklist)
		if len(excludePatterns) > 0 {
			excluded := false
			for _, pattern := range excludePatterns {
				if pattern.MatchString(entry.Name()) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		filteredEntries = append(filteredEntries, entry)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(filteredEntries, func(i, j int) bool {
		if filteredEntries[i].IsDir() && !filteredEntries[j].IsDir() {
			return true
		}
		if !filteredEntries[i].IsDir() && filteredEntries[j].IsDir() {
			return false
		}
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})

	// Collapse repeated entries if option is enabled
	if opts.CollapseRepeated {
		filteredEntries = collapseRepeatedEntries(filteredEntries)
	}

	for i, entry := range filteredEntries {
		isLastEntry := i == len(filteredEntries)-1

		// Build the tree symbols
		var connector string
		if isLastEntry {
			connector = "└── "
		} else {
			connector = "├── "
		}

		entryName := entry.Name()

		// Handle pattern collapse
		if opts.CollapsePattern {
			pattern := extractBasePattern(entryName)
			if seenPatterns[pattern] {
				// This pattern has been seen before, show collapsed version
				result.WriteString(prefix + connector + pattern + " (...)\n")
				continue
			} else {
				// First time seeing this pattern, mark it as seen
				seenPatterns[pattern] = true
			}
		}

		// Write the current entry
		result.WriteString(prefix + connector + entryName + "\n")

		// If it's a directory and not a collapsed entry, recursively process it
		if entry.IsDir() && !isCollapsedEntry(entry) {
			var newPrefix string
			if isLastEntry {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}

			subDir := filepath.Join(dir, entryName)
			err := buildTreeWithOptions(subDir, newPrefix, isLastEntry, result, opts, includePatterns, excludePatterns, seenPatterns)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// collapsedEntry represents a collapsed entry with count information
type collapsedEntry struct {
	name  string
	count int
	isDir bool
}

// Name returns the display name for the collapsed entry
func (c *collapsedEntry) Name() string {
	if c.count > 1 {
		return fmt.Sprintf("%s (%d times)", c.name, c.count)
	}
	return c.name
}

// IsDir returns whether this is a directory
func (c *collapsedEntry) IsDir() bool {
	return c.isDir
}

// Type returns the file type info (dummy implementation for interface compatibility)
func (c *collapsedEntry) Type() os.FileMode {
	if c.isDir {
		return os.ModeDir
	}
	return 0
}

// Info returns file info (dummy implementation for interface compatibility)
func (c *collapsedEntry) Info() (os.FileInfo, error) {
	return nil, fmt.Errorf("info not available for collapsed entry")
}

// IsCollapsed returns true if this is a collapsed entry (not a real filesystem entry)
func (c *collapsedEntry) IsCollapsed() bool {
	return true
}

// isCollapsedEntry checks if an entry is a collapsed entry
func isCollapsedEntry(entry os.DirEntry) bool {
	if collapsed, ok := entry.(*collapsedEntry); ok {
		return collapsed.IsCollapsed()
	}
	return false
}

// collapseRepeatedEntries detects and collapses repeated entries
func collapseRepeatedEntries(entries []os.DirEntry) []os.DirEntry {
	if len(entries) <= 1 {
		return entries
	}

	// Group entries by their base pattern (removing numeric prefixes)
	groups := make(map[string][]os.DirEntry)
	for _, entry := range entries {
		pattern := extractPattern(entry.Name())
		groups[pattern] = append(groups[pattern], entry)
	}

	var result []os.DirEntry
	for pattern, groupEntries := range groups {
		if len(groupEntries) > 1 && hasNumericSequence(groupEntries) {
			// Create a collapsed entry
			collapsed := &collapsedEntry{
				name:  pattern,
				count: len(groupEntries),
				isDir: groupEntries[0].IsDir(),
			}
			result = append(result, collapsed)
		} else {
			// Add all entries individually
			result = append(result, groupEntries...)
		}
	}

	// Sort the result again
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir() && !result[j].IsDir() {
			return true
		}
		if !result[i].IsDir() && result[j].IsDir() {
			return false
		}
		return result[i].Name() < result[j].Name()
	})

	return result
}

// extractPattern extracts the base pattern from a name by removing numeric prefixes
func extractPattern(name string) string {
	// Look for pattern like "0_name", "1_name", etc.
	if len(name) > 2 && name[1] == '_' {
		if _, err := strconv.Atoi(string(name[0])); err == nil {
			return name[2:]
		}
	}

	// Look for pattern like "10_name", "11_name", etc.
	for i := 1; i < len(name); i++ {
		if name[i] == '_' {
			if _, err := strconv.Atoi(name[:i]); err == nil {
				return name[i+1:]
			}
			break
		}
	}

	return name
}

// extractBasePattern extracts the base pattern from a name by removing numeric prefixes
func extractBasePattern(name string) string {
	// Look for pattern like "0_name", "1_name", etc.
	if len(name) > 2 && name[1] == '_' {
		if _, err := strconv.Atoi(string(name[0])); err == nil {
			return name[2:]
		}
	}

	// Look for pattern like "10_name", "11_name", etc.
	for i := 1; i < len(name); i++ {
		if name[i] == '_' {
			if _, err := strconv.Atoi(name[:i]); err == nil {
				return name[i+1:]
			}
			break
		}
	}

	return name
}

// hasNumericSequence checks if the entries form a numeric sequence
func hasNumericSequence(entries []os.DirEntry) bool {
	if len(entries) < 2 {
		return false
	}

	// Extract numeric prefixes
	var numbers []int
	for _, entry := range entries {
		name := entry.Name()
		var numStr string

		// Find the numeric prefix
		for i := 0; i < len(name); i++ {
			if name[i] == '_' {
				numStr = name[:i]
				break
			}
		}

		if numStr == "" {
			return false
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return false
		}
		numbers = append(numbers, num)
	}

	// Check if it's a sequence (doesn't need to be consecutive, just ordered)
	sort.Ints(numbers)
	return len(numbers) >= 2
}
