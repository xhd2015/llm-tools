package whats_next

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// WhatsNextRequest represents the input parameters for the whats_next tool
type WhatsNextRequest struct {
	Explanation string `json:"explanation"`
}

// WhatsNextResponse represents the output of the whats_next tool
type WhatsNextResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	UserInput     string `json:"user_input"`
	CommandOutput string `json:"command_output"`
}

// GetToolDefinition returns the JSON schema definition for the whats_next tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Run an interactive CLI that executes 'whats_next' command and waits for user's follow-up questions. This tool will run the whats_next command and capture the user's input for further processing.`,
		Name:        "whats_next",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.EXPLANATION,
				},
			},
			Required: []string{},
		},
	}
}

// WhatsNext executes the whats_next tool with the given parameters
func WhatsNext(req WhatsNextRequest) (*WhatsNextResponse, error) {
	// Execute the whats_next command
	cmd := exec.Command("whats_next")

	// Set up pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start whats_next command: %w", err)
	}

	// Read the command output
	var commandOutput strings.Builder

	// Read from stdout
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			commandOutput.WriteString(line + "\n")
			fmt.Println(line) // Print to console for user to see
		}
	}()

	// Read from stderr
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			commandOutput.WriteString("ERROR: " + line + "\n")
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", line)
		}
	}()

	// Wait for user input
	fmt.Print("user> ")
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read user input: %w", err)
	}

	// Clean up the input
	userInput = strings.TrimSpace(userInput)

	// Send the user input to the command
	if _, err := stdin.Write([]byte(userInput + "\n")); err != nil {
		return nil, fmt.Errorf("failed to write to command stdin: %w", err)
	}

	// Close stdin to signal end of input
	stdin.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		// Command might exit with non-zero status, but that's okay for this tool
		// We'll still return the output
	}

	return &WhatsNextResponse{
		Success:       true,
		Message:       "Successfully executed whats_next and captured user input",
		UserInput:     userInput,
		CommandOutput: commandOutput.String(),
	}, nil
}

func ParseJSONRequest(jsonInput string) (WhatsNextRequest, error) {
	var req WhatsNextRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return WhatsNextRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the whats_next tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req WhatsNextRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := WhatsNext(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
