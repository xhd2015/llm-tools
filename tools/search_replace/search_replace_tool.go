package search_replace

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools search_replace performs a search and replace operation on a file

Usage: llm-tools search_replace <file> [OPTIONS]

Options:
  --old-string <text>          string to be replaced (required)
  --new-string <text>          string to replace with (required)

Examples:
  llm-tools search_replace main.go --old-string "old_function" --new-string "new_function"
  llm-tools search_replace src/config.json --old-string '"debug": true' --new-string '"debug": false'
  llm-tools search_replace README.md --old-string "Version 1.0" --new-string "Version 2.0"
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var oldString string
	var newString string

	args, err := flags.
		String("--workspace-root", &workspaceRoot).
		String("--old", &oldString).
		String("--new", &newString).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires file")
	}

	file := args[0]
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if oldString == "" {
		return fmt.Errorf("--old is required")
	}

	if newString == "" {
		return fmt.Errorf("--new is required")
	}

	req := SearchReplaceRequest{
		File: file,
		Old:  oldString,
		New:  newString,
	}

	response, err := SearchReplace(req, workspaceRoot)
	if err != nil {
		return err
	}

	// Print results
	fmt.Println(response.Message)

	return nil
}
