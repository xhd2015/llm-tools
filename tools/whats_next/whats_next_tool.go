package whats_next

import (
	"fmt"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools whats_next runs an interactive CLI and waits for user follow-up questions

Usage: llm-tools whats_next [OPTIONS]

Options:
  --explanation <text>         explanation for the operation

Examples:
  llm-tools whats_next
  llm-tools whats_next --explanation "Waiting for user follow-up"
`

func HandleCli(args []string) error {
	var explanation string

	args, err := flags.String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	req := WhatsNextRequest{
		Explanation: explanation,
	}

	response, err := WhatsNext(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("User Input: %s\n", response.UserInput)

	if response.CommandOutput != "" {
		fmt.Println()
		fmt.Println("Command Output:")
		fmt.Print(response.CommandOutput)
	}

	return nil
}
