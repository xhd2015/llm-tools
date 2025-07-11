package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/llm-tools/tools/batch_read_file"
	"github.com/xhd2015/llm-tools/tools/get_workspace_root"
	"github.com/xhd2015/llm-tools/tools/grep_search"
	"github.com/xhd2015/llm-tools/tools/list_dir"
	"github.com/xhd2015/llm-tools/tools/read_file"
	"github.com/xhd2015/llm-tools/tools/run_terminal_cmd"
	"github.com/xhd2015/llm-tools/tools/tree"
)

const help = `
llm-tools help to parse flags

Usage: llm-tools <cmd> [OPTIONS]

Available commands:
  create <name>                    create a new project
  help                             show help message

Options:
  --dir <dir>                      set the output directory
  -v,--verbose                     show verbose info  

Examples:
  llm-tools help                         show help message
  llm-tools create my_project            create a new project named my_project
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires sub command: create")
	}
	cmd := args[0]
	args = args[1:]
	if cmd == "help" || cmd == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	switch cmd {
	case "get_workspace_root":
		return get_workspace_root.HandleCli(args)
	case "batch_read_file":
		return batch_read_file.HandleCli(args)
	case "read_file":
		return read_file.HandleCli(args)
	case "tree":
		return tree.HandleCli(args)
	case "grep_search":
		return grep_search.HandleCli(args)
	case "list_dir":
		return list_dir.HandleCli(args)
	case "run_terminal_cmd":
		return run_terminal_cmd.HandleCli(args)
	default:
		return fmt.Errorf("unrecognized: %s", cmd)
	}
}
