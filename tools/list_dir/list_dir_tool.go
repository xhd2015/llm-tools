package list_dir

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools list_dir lists the contents of a directory

Usage: llm-tools list_dir <path> [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools list_dir .
  llm-tools list_dir src --workspace-root /path/to/workspace
  llm-tools list_dir tools --explanation "Exploring tool structure"
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
		return fmt.Errorf("directory path is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	relativePath := args[0]

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	req := ListDirRequest{
		WorkspaceRoot:         workspaceRoot,
		RelativeWorkspacePath: relativePath,
		Explanation:           explanation,
	}

	response, err := ListDir(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Directory: %s\n", response.Path)
	fmt.Printf("Items: %d\n", response.Count)
	fmt.Println()

	for _, item := range response.Contents {
		fmt.Println(item)
	}

	return nil
}
