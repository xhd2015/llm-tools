package run_bash_script

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools run_bash_script executes terminal commands

Usage: llm-tools run_bash_script <command> [OPTIONS]

Options:
  --background                 run command in background
  --explanation <text>         explanation for the operation

Examples:
  llm-tools run_bash_script "ls -la"
  llm-tools run_bash_script "go build" --explanation "Building the project"
  llm-tools run_bash_script "python server.py" --background
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

	req := RunBashScriptRequest{
		Script:      command,
		Explanation: explanation,
	}

	response, err := RunBashScript(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Script: %s\n", req.Script)
	fmt.Printf("Exit code: %d\n", response.ExitCode)
	if response.Duration != "" {
		fmt.Printf("Duration: %s\n", response.Duration)
	}

	if response.Error != "" {
		fmt.Printf("Error: %s\n", response.Error)
	}

	if response.Output != "" {
		fmt.Println()
		fmt.Println("Output:")
		fmt.Println(response.Output)
	}

	return nil
}
