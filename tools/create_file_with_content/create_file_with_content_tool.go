package create_file_with_content

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools create_file_with_content creates a file with content

Usage: llm-tools create_file_with_content <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --content <text>             content to write to the file (required)
  --dangerous-override         allow overwriting existing files (default: false)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools create_file_with_content newfile.txt --content "Hello, world!"
  llm-tools create_file_with_content src/main.go --content "package main\n\nfunc main() {}" --workspace-root /path/to/workspace
  llm-tools create_file_with_content existing.txt --content "New content" --dangerous-override
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var content string
	var dangerousOverride bool
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		String("--content", &content).
		Bool("--dangerous-override", &dangerousOverride).
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

	req := CreateFileWithContentRequest{
		WorkspaceRoot:     workspaceRoot,
		TargetFile:        targetFile,
		NonEmptyContent:   content,
		DangerousOverride: dangerousOverride,
		Explanation:       explanation,
	}

	response, err := CreateFileWithContent(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Bytes Written: %d\n", response.BytesWritten)
	fmt.Printf("Overwritten: %v\n", response.Overwritten)

	return nil
}
