package codebase_search

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools codebase_search performs semantic search to find code by meaning

Usage: llm-tools codebase_search <query> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --target-directories <dirs>  comma-separated list of target directories to search in
  --search-only-prs            only search pull requests and return no code results
  --explanation <text>         explanation for the operation

Examples:
  llm-tools codebase_search "How does user authentication work?"
  llm-tools codebase_search "Where are user roles checked?" --target-directories backend/auth/
  llm-tools codebase_search "What is the login flow?" --workspace-root /path/to/workspace
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var targetDirectories string
	var searchOnlyPrs bool
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		String("--target-directories", &targetDirectories).
		Bool("--search-only-prs", &searchOnlyPrs).
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

	// Parse target directories
	var targetDirs []string
	if targetDirectories != "" {
		targetDirs = strings.Split(targetDirectories, ",")
		for i, dir := range targetDirs {
			targetDirs[i] = strings.TrimSpace(dir)
		}
	}

	req := CodebaseSearchRequest{
		WorkspaceRoot:     workspaceRoot,
		Query:             query,
		SearchOnlyPrs:     searchOnlyPrs,
		TargetDirectories: targetDirs,
		Explanation:       explanation,
	}

	response, err := CodebaseSearch(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Search query: %s\n", response.Query)
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
		fmt.Printf("%s:%d (score: %.2f)\n", match.File, match.Line, match.Score)
		fmt.Printf("  %s\n", match.Content)
		if match.Context != "" {
			fmt.Printf("  Context:\n%s\n", match.Context)
		}
		fmt.Println()
	}

	return nil
}
