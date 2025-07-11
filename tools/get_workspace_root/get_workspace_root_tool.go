package get_workspace_root

import (
	"fmt"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools get_workspace_root gets the workspace root directory

Usage: llm-tools get_workspace_root [OPTIONS]

Options:
  --workspace-root <path>  workspace root path
  --explanation <text>     explanation for the operation

Examples:
  llm-tools get_workspace_root                    get current workspace root
  llm-tools get_workspace_root --explanation "..."  get workspace root with explanation
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

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	req := GetWorkspaceRootRequest{
		Explanation: explanation,
	}

	response, err := GetWorkspaceRoot(req, workspaceRoot)
	if err != nil {
		return err
	}

	fmt.Println(response.WorkspaceRoot)
	return nil
}
