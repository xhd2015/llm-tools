package rename_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools rename_file renames or moves a file from one location to another

Usage: llm-tools rename_file <source_file> <target_file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools rename_file oldfile.txt newfile.txt
  llm-tools rename_file src/old.go src/new.go --workspace-root /path/to/workspace
  llm-tools rename_file file.txt backup/file.txt --explanation "Moving file to backup directory"
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

	if len(args) < 2 {
		return fmt.Errorf("both source and target files are required")
	}

	if len(args) > 2 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	sourceFile := args[0]
	targetFile := args[1]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := RenameFileRequest{
		WorkspaceRoot: workspaceRoot,
		SourceFile:    sourceFile,
		TargetFile:    targetFile,
		Explanation:   explanation,
	}

	response, err := RenameFile(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Source File Path: %s\n", response.SourceFilePath)
	fmt.Printf("Target File Path: %s\n", response.TargetFilePath)

	return nil
}
