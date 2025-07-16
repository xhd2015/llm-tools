package web_search

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools web_search searches the web for real-time information

Usage: llm-tools web_search <search_term> [OPTIONS]

Options:
  --explanation <text>         explanation for the operation

Examples:
  llm-tools web_search "Go programming language"
  llm-tools web_search "latest news about AI" --explanation "Getting current AI news"
  llm-tools web_search "Docker installation guide"
`

func HandleCli(args []string) error {
	var explanation string

	args, err := flags.String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("search term is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args[1:], ","))
	}

	searchTerm := args[0]

	req := WebSearchRequest{
		SearchTerm:  searchTerm,
		Explanation: explanation,
	}

	response, err := WebSearch(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Search Term: %s\n", response.SearchTerm)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Results (%d):\n", len(response.Results))

	for i, result := range response.Results {
		fmt.Printf("  %d. %s\n", i+1, result.Title)
		fmt.Printf("     URL: %s\n", result.URL)
		fmt.Printf("     Snippet: %s\n", result.Snippet)
		fmt.Printf("\n")
	}

	return nil
}
