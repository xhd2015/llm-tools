package delete_file

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools delete_file safely deletes a file with security checks

Usage: llm-tools delete_file <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools delete_file temp.txt
  llm-tools delete_file logs/old.log --workspace-root /path/to/workspace
  llm-tools delete_file backup.bak --explanation "Cleaning up old backup files"
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
		return fmt.Errorf("target file is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args[1:], ","))
	}

	targetFile := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := DeleteFileRequest{
		WorkspaceRoot: workspaceRoot,
		TargetFile:    targetFile,
		Explanation:   explanation,
	}

	response, err := DeleteFile(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("File: %s\n", response.DeletedFile)
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)

	if !response.Success {
		os.Exit(1)
	}

	return nil
}
