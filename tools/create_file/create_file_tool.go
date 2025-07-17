package create_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools create_file creates a new empty file

Usage: llm-tools create_file <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --mkdirs                     create parent directories if they don't exist (default: false)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools create_file newfile.txt
  llm-tools create_file src/main.go --workspace-root /path/to/workspace
  llm-tools create_file deep/path/file.txt --mkdirs
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var mkdirs bool
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		Bool("--mkdirs", &mkdirs).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("file path is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	filePath := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := CreateFileRequest{
		WorkspaceRoot: workspaceRoot,
		FilePath:      filePath,
		Mkdirs:        mkdirs,
		Explanation:   explanation,
	}

	response, err := CreateFile(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("File Path: %s\n", response.FilePath)
	fmt.Printf("Dirs Created: %v\n", response.DirsCreated)

	return nil
}
