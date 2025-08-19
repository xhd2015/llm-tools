package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/llm-tools/tools/batch_read_file"
	"github.com/xhd2015/llm-tools/tools/create_file"
	"github.com/xhd2015/llm-tools/tools/create_file_with_content"
	"github.com/xhd2015/llm-tools/tools/delete_file"
	"github.com/xhd2015/llm-tools/tools/edit_file"
	"github.com/xhd2015/llm-tools/tools/file_search"
	"github.com/xhd2015/llm-tools/tools/get_workspace_root"
	"github.com/xhd2015/llm-tools/tools/grep_search"
	"github.com/xhd2015/llm-tools/tools/list_dir"
	"github.com/xhd2015/llm-tools/tools/mcp_client"
	"github.com/xhd2015/llm-tools/tools/read_file"
	"github.com/xhd2015/llm-tools/tools/rename_file"
	"github.com/xhd2015/llm-tools/tools/run_terminal_cmd"
	"github.com/xhd2015/llm-tools/tools/search_replace"
	"github.com/xhd2015/llm-tools/tools/send_answer"
	"github.com/xhd2015/llm-tools/tools/todo_write"
	"github.com/xhd2015/llm-tools/tools/tree"
	"github.com/xhd2015/llm-tools/tools/web_search"
	"github.com/xhd2015/llm-tools/tools/whats_next"
	"github.com/xhd2015/llm-tools/tools/write_file"
)

const help = `
llm-tools - A collection of tools for LLM development

Usage: llm-tools <cmd> [OPTIONS]

Available commands:
  get_workspace_root               get the workspace root directory
  batch_read_file                  read multiple files in a single batch operation
  read_file                        read the contents of a file
  tree                             display directory tree structure
  grep_search                      search for text patterns using regex
  list_dir                         list the contents of a directory
  run_terminal_cmd                 execute terminal commands
  create_file                      create a new empty file with optional directory creation
  create_file_with_content         create a file with content and optional override protection
  write_file                       create a file with content
  rename_file                      rename or move a file
  edit_file                        edit a file by replacing all occurrences of a string
  search_replace                   search and replace a single occurrence in a file
  send_answer                      send a structured answer to another tool
  mcp_client                       communicate with external MCP servers
  file_search                      search for files by name pattern
  delete_file                      delete a file safely
  web_search                       search the web for real-time information
  todo_write                       create and manage structured task lists
  whats_next                       run interactive CLI and wait for user follow-up questions
  help                             show this help message

Options:
  -h, --help                       show help message for specific command

Examples:
  llm-tools help                         show this help message
  llm-tools read_file --help             show help for read_file command
  llm-tools grep_search "pattern" .      search for pattern in current directory
  llm-tools create_file new.txt          create a new empty file
  llm-tools create_file deep/path/file.txt --mkdirs  create file with parent directories
  llm-tools create_file_with_content new.txt --content "content"  create file with content
  llm-tools edit_file file.txt --old-string "old" --new-string "new"
  llm-tools search_replace file.txt --old-string "unique_old" --new-string "new"
  llm-tools whats_next                   run interactive CLI for follow-up questions
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
		return fmt.Errorf("requires sub command, use 'llm-tools help' to see available commands")
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
	case "create_file":
		return create_file.HandleCli(args)
	case "create_file_with_content":
		return create_file_with_content.HandleCli(args)
	case "write_file":
		return write_file.HandleCli(args)
	case "rename_file":
		return rename_file.HandleCli(args)
	case "edit_file":
		return edit_file.HandleCli(args)
	case "search_replace":
		return search_replace.HandleCli(args)
	case "send_answer":
		return send_answer.HandleCli(args)
	case "mcp_client":
		return mcp_client.HandleCli(args)
	case "file_search":
		return file_search.HandleCli(args)
	case "delete_file":
		return delete_file.HandleCli(args)
	case "web_search":
		return web_search.HandleCli(args)
	case "todo_write":
		return todo_write.HandleCli(args)
	case "whats_next":
		return whats_next.HandleCli(args)
	default:
		return fmt.Errorf("unrecognized: %s", cmd)
	}
}
