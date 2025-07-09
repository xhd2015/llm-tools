package tree

import (
	"crypto/md5"
	"encoding/hex"
)

type CollapseOptions struct {
	// CollapseRepeated collapses repeated entries into a single entry with count
	CollapseRepeated bool
	// CollapsePattern collapses duplicate patterns, showing full structure only for first appearance
	CollapsePattern bool
	// CollapsedDirs are directories that should be collapsed
	CollapsedDirs []string
}

type internalItem struct {
	item          Item
	children      []internalItem
	uniquePattern string
}

func toInternal(items []Item) []internalItem {
	if items == nil {
		return nil
	}
	result := make([]internalItem, 0, len(items))
	for _, item := range items {
		result = append(result, internalItem{
			item:     item,
			children: toInternal(item.Children),
		})
	}
	return result
}

func toExternal(internalItems []internalItem) []Item {
	if internalItems == nil {
		return nil
	}
	result := make([]Item, 0, len(internalItems))
	for _, internalItem := range internalItems {
		item := internalItem.item
		item.Children = toExternal(internalItem.children)
		result = append(result, item)
	}
	return result
}

func recomputePatterns(items []internalItem) {
	var compute func(item *internalItem) string
	compute = func(item *internalItem) string {
		h := md5.New()
		h.Write([]byte(item.item.Name))
		for i := 0; i < len(item.children); i++ {
			sub := compute(&item.children[i])
			h.Write([]byte(sub))
		}
		sum := h.Sum(nil)
		uniq := hex.EncodeToString(sum)
		item.uniquePattern = uniq
		return uniq
	}
	for i := 0; i < len(items); i++ {
		compute(&items[i])
	}
}

func Collapse(item Item, opts CollapseOptions) Item {
	collapsePattern := opts.CollapsePattern
	collapseRepeated := opts.CollapseRepeated

	if !collapsePattern && !collapseRepeated {
		return item
	}
	if len(item.Children) == 0 {
		return item
	}
	// Create a copy of the item to avoid modifying the original
	result := item
	if len(item.Children) == 0 {
		return result
	}
	children := toInternal(item.Children)
	recomputePatterns(children)

	if collapseRepeated {
		children = collapseRepeatedItems(children)
	}

	if collapsePattern {
		collapsePatternItems(children)
	}
	if len(opts.CollapsedDirs) > 0 {
		collapseByNames(children, opts.CollapsedDirs)
	}

	result.Children = toExternal(children)
	return result
}

func collapseRepeatedItems(items []internalItem) []internalItem {
	n := len(items)

	// First recursively process children
	for i := 0; i < n; i++ {
		if len(items[i].children) > 0 {
			items[i].children = collapseRepeatedItems(items[i].children)
		}
	}

	// Find all possible consecutive repeating patterns
	type pattern struct {
		start       int
		length      int
		repeatCount int
		totalLength int
	}

	var patterns []pattern

	// Look for patterns starting at each position
	for i := 0; i < n; i++ {
		// Try different pattern lengths
		for patternLen := 1; patternLen <= (n-i)/2; patternLen++ {
			if i+patternLen >= n {
				break
			}

			// Check if we have consecutive repeating pattern of length patternLen
			repeatCount := 1
			pos := i + patternLen

			// Count consecutive repeats only
			for pos+patternLen <= n {
				matches := true
				for j := 0; j < patternLen; j++ {
					if items[i+j].uniquePattern != items[pos+j].uniquePattern {
						matches = false
						break
					}
				}
				if !matches {
					break
				}
				repeatCount++
				pos += patternLen
			}

			// Only consider patterns that repeat consecutively at least twice
			if repeatCount >= 2 {
				totalLen := patternLen * repeatCount
				patterns = append(patterns, pattern{
					start:       i,
					length:      patternLen,
					repeatCount: repeatCount,
					totalLength: totalLen,
				})
			}
		}
	}

	// Choose the best pattern to collapse
	// Prefer patterns with higher repeat counts, then longer patterns, then later starting position
	var bestPattern *pattern
	for i := range patterns {
		p := &patterns[i]
		if bestPattern == nil {
			bestPattern = p
		} else {
			// Prefer patterns with higher repeat counts first
			if p.repeatCount > bestPattern.repeatCount {
				bestPattern = p
			} else if p.repeatCount == bestPattern.repeatCount {
				// If same repeat count, prefer longer patterns
				if p.length > bestPattern.length {
					bestPattern = p
				} else if p.length == bestPattern.length {
					// If same length, prefer later starting position
					if p.start > bestPattern.start {
						bestPattern = p
					}
				}
			}
		}
	}

	if bestPattern != nil {
		// Build result with the selected pattern collapsed
		result := make([]internalItem, 0, n)

		// Add items before the pattern
		for i := 0; i < bestPattern.start; i++ {
			result = append(result, items[i])
		}

		// Add the pattern items with repeat counts
		for j := 0; j < bestPattern.length; j++ {
			item := items[bestPattern.start+j]
			item.item.SubsequentRepeated = bestPattern.repeatCount - 1
			result = append(result, item)
		}

		// Add items after the pattern
		afterPatternStart := bestPattern.start + bestPattern.totalLength
		for i := afterPatternStart; i < n; i++ {
			result = append(result, items[i])
		}

		return result
	}

	// No patterns found, return original items
	return items
}

// collapsePatternItems collapses duplicate patterns, showing full structure only for first appearance
func collapsePatternItems(items []internalItem) {
	// name mapping
	nameMapping := make(map[string]map[string]bool)
	checkIsNew := func(name string, uniqPattern string) bool {
		patternMapping, ok := nameMapping[name]
		if !ok {
			patternMapping = make(map[string]bool)
			nameMapping[name] = patternMapping
		}
		ok = patternMapping[uniqPattern]
		if ok {
			return false
		}
		patternMapping[uniqPattern] = true
		return true
	}

	var collapse func(item *internalItem)
	collapse = func(item *internalItem) {
		if len(item.children) > 0 && !checkIsNew(item.item.Name, item.uniquePattern) {
			item.item.CollapsedPatternChildren += len(item.children)
			item.children = nil
		}
		for i := 0; i < len(item.children); i++ {
			collapse(&item.children[i])
		}
	}

	for i := 0; i < len(items); i++ {
		collapse(&items[i])
	}
}

func collapseByNames(items []internalItem, collapsedDirs []string) {
	var collapse func(item *internalItem)
	collapse = func(item *internalItem) {
		if len(item.children) > 0 {
			itemName := getName(item.item)
			var found bool
			for _, dir := range collapsedDirs {
				if dir == itemName {
					found = true
					break
				}
			}

			if found {
				item.item.CollapsedPatternChildren += len(item.children)
				item.children = nil
			}
		}
		for i := 0; i < len(item.children); i++ {
			collapse(&item.children[i])
		}
	}

	for i := 0; i < len(items); i++ {
		collapse(&items[i])
	}
}
