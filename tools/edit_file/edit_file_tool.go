package edit_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools edit_file edits an existing file by replacing occurrences of old_string with new_string

Usage: llm-tools edit_file <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --old-string <text>          string to be replaced (required)
  --new-string <text>          string to replace with (required)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools edit_file main.go --old-string "old_function" --new-string "new_function"
  llm-tools edit_file src/config.json --old-string '"debug": true' --new-string '"debug": false'
  llm-tools edit_file README.md --old-string "Version 1.0" --new-string "Version 2.0" --workspace-root /path/to/workspace
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var oldString string
	var newString string
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		String("--old-string", &oldString).
		String("--new-string", &newString).
		String("--explanation", &explanation).
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

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := EditFileRequest{
		WorkspaceRoot: workspaceRoot,
		TargetFile:    targetFile,
		OldString:     oldString,
		NewString:     newString,
		Explanation:   explanation,
	}

	response, err := EditFile(req)
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
