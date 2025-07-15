package create_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools create_file creates a new file with the specified content

Usage: llm-tools create_file <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --content <text>             content to write to the file (required)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools create_file newfile.txt --content "Hello, world!"
  llm-tools create_file src/main.go --content "package main\n\nfunc main() {}" --workspace-root /path/to/workspace
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var content string
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		String("--content", &content).
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

	if content == "" {
		return fmt.Errorf("--content is required")
	}

	targetFile := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := CreateFileRequest{
		WorkspaceRoot: workspaceRoot,
		TargetFile:    targetFile,
		Content:       content,
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
	fmt.Printf("Bytes Written: %d\n", response.BytesWritten)

	return nil
}
