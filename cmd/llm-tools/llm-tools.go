package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
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
	case "batch_read_file":
		return handle(args)
	case "read_file":
		return handle(args)
	case "tree":
		return tree.HandleCli(args)
	case "grep_search":
		return handle(args)
	case "list_dir":
		return handle(args)
	case "run_terminal_cmd":
		return handle(args)
	default:
		return fmt.Errorf("unrecognized: %s", cmd)
	}
}

func handle(args []string) error {
	var verbose bool
	var dir string
	args, err := flags.String("--dir", &dir).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	_ = dir
	_ = verbose
	return nil
}
