package grep_search

import (
	"fmt"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools grep_search performs fast regex searches using ripgrep

Usage: llm-tools grep_search <pattern> [OPTIONS]

Options:
  --case-sensitive             enable case sensitive search
  --exclude <pattern>          exclude files/directories matching pattern
  --include <pattern>          include files/directories matching pattern
  --explanation <text>         explanation for the operation

Examples:
  llm-tools grep_search "function.*main"
  llm-tools grep_search "HandleCli" --include "*.go"
  llm-tools grep_search "TODO" --exclude "*.log" --case-sensitive
`

func HandleCli(args []string) error {
	var caseSensitive bool
	var excludePattern string
	var includePattern string
	var explanation string

	args, err := flags.Bool("--case-sensitive", &caseSensitive).
		String("--exclude", &excludePattern).
		String("--include", &includePattern).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("search pattern is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	query := args[0]

	req := GrepSearchRequest{
		Query:          query,
		CaseSensitive:  caseSensitive,
		ExcludePattern: excludePattern,
		IncludePattern: includePattern,
		Explanation:    explanation,
	}

	response, err := GrepSearch(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Search query: %s\n", response.SearchQuery)
	fmt.Printf("Total matches: %d", response.TotalMatches)
	if response.Truncated {
		fmt.Printf(" (truncated)")
	}
	fmt.Println()
	fmt.Println()

	if len(response.Matches) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	for _, match := range response.Matches {
		fmt.Printf("%s:%d", match.File, match.Line)
		if match.Column > 0 {
			fmt.Printf(":%d", match.Column)
		}
		fmt.Printf(": %s\n", match.Content)
	}

	return nil
}
