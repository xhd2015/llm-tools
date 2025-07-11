package run_terminal_cmd

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools run_terminal_cmd executes terminal commands

Usage: llm-tools run_terminal_cmd <command> [OPTIONS]

Options:
  --background                 run command in background
  --explanation <text>         explanation for the operation

Examples:
  llm-tools run_terminal_cmd "ls -la"
  llm-tools run_terminal_cmd "go build" --explanation "Building the project"
  llm-tools run_terminal_cmd "python server.py" --background
`

func HandleCli(args []string) error {
	var isBackground bool
	var explanation string

	args, err := flags.Bool("--background", &isBackground).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("command is required")
	}

	// Join all remaining args as the command
	command := strings.Join(args, " ")

	req := RunTerminalCmdRequest{
		Command:      command,
		IsBackground: isBackground,
		Explanation:  explanation,
	}

	response, err := RunTerminalCmd(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Command: %s\n", response.Command)
	fmt.Printf("Exit code: %d\n", response.ExitCode)
	fmt.Printf("Working directory: %s\n", response.WorkingDir)
	if response.Duration != "" {
		fmt.Printf("Duration: %s\n", response.Duration)
	}
	fmt.Printf("Shell info: %s\n", response.ShellInfo)

	if response.Error != "" {
		fmt.Printf("Error: %s\n", response.Error)
	}

	if response.CommandOutput != "" {
		fmt.Println()
		fmt.Println("Output:")
		fmt.Println(response.CommandOutput)
	}

	return nil
}
