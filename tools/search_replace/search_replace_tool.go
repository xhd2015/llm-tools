package search_replace

import (
	"fmt"

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
		String("--old-string", &oldString).
		String("--new-string", &newString).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("target file is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	if oldString == "" {
		return fmt.Errorf("--old-string is required")
	}

	if newString == "" {
		return fmt.Errorf("--new-string is required")
	}

	targetFile := args[0]

	req := SearchReplaceRequest{
		WorkspaceRoot: workspaceRoot,
		FilePath:      targetFile,
		OldString:     oldString,
		NewString:     newString,
	}

	response, err := SearchReplace(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("File Path: %s\n", response.FilePath)
	fmt.Printf("Changes Count: %d\n", response.ChangesCount)
	fmt.Printf("Lines Changed: %d\n", response.LinesChanged)

	return nil
}
