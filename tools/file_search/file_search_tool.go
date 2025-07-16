package file_search

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools file_search performs fast fuzzy file search based on file path

Usage: llm-tools file_search <query> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools file_search "main.go"
  llm-tools file_search "user" --workspace-root /path/to/workspace
  llm-tools file_search "test.js"
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("search query is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args[1:], ","))
	}

	query := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := FileSearchRequest{
		WorkspaceRoot: workspaceRoot,
		Query:         query,
		Explanation:   explanation,
	}

	response, err := FileSearch(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Search query: %s\n", req.Query)
	fmt.Printf("Total matches: %d", response.TotalMatches)
	if response.Truncated {
		fmt.Printf(" (truncated to 10)")
	}
	fmt.Println()
	fmt.Println()

	if len(response.Matches) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	for _, match := range response.Matches {
		fmt.Printf("%s (score: %.2f)\n", match.File, match.Score)
	}

	return nil
}
