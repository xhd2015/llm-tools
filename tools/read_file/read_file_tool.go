package read_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools read_file reads the contents of a file

Usage: llm-tools read_file <file> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --entire-file                read entire file
  --start-line <num>           start line number (1-indexed)
  --end-line <num>             end line number (1-indexed, inclusive)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools read_file file.go --entire-file
  llm-tools read_file file.go --start-line 1 --end-line 50
  llm-tools read_file file.go --workspace-root /path/to/workspace --start-line 10 --end-line 20
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var entireFile bool
	var startLine int
	var endLine int
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		Bool("--entire-file", &entireFile).
		Int("--start-line", &startLine).
		Int("--end-line", &endLine).
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

	targetFile := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Validate parameters
	if !entireFile && (startLine == 0 || endLine == 0) {
		return fmt.Errorf("when not reading entire file, both --start-line and --end-line are required")
	}

	req := ReadFileRequest{
		WorkspaceRoot:              workspaceRoot,
		TargetFile:                 targetFile,
		ShouldReadEntireFile:       entireFile,
		StartLineOneIndexed:        startLine,
		EndLineOneIndexedInclusive: endLine,
		Explanation:                explanation,
	}

	response, err := ReadFile(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("File: %s\n", targetFile)
	fmt.Printf("Lines: %s (Total: %d)\n", response.LinesShown, response.TotalLines)
	if response.Outline != "" {
		fmt.Printf("Outline: %s\n", response.Outline)
	}
	fmt.Println()
	fmt.Println(response.Contents)

	return nil
}
