package send_answer

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools send_answer sends a structured answer to ~/llm-answer.json

Usage: llm-tools send_answer [OPTIONS]

Options:
  --answer <text>              answer string (can be used multiple times for array)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools send_answer --answer "First answer" --answer "Second answer"
  llm-tools send_answer --answer "Single answer" --explanation "Providing structured response"
`

func HandleCli(args []string) error {
	var answers []string
	var explanation string

	args, err := flags.StringSlice("--answer", &answers).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	if len(answers) == 0 {
		return fmt.Errorf("at least one --answer is required")
	}

	req := SendAnswerRequest{
		Answer:      answers,
		Explanation: explanation,
	}

	response, err := SendAnswer(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Answer: [%s]\n", strings.Join(req.Answer, ", "))

	return nil
}
