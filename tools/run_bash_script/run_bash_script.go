package run_bash_script

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// RunBashScriptRequest represents the input parameters for the run_bash_script tool
type RunBashScriptRequest struct {
	Cwd         string `json:"cwd"`
	Script      string `json:"script"`
	Explanation string `json:"explanation,omitempty"`
}

// RunBashScriptResponse represents the output of the run_bash_script tool
type RunBashScriptResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Hint     string `json:"hint,omitempty"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// GetToolDefinition returns the JSON schema definition for the run_bash_script tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `PROPOSE a bash script to run on behalf of the user, in foreground.

Adhere to the following guidelines:
1. For ANY commands that would require user interaction, ASSUME THE USER IS NOT AVAILABLE TO INTERACT and PASS THE NON-INTERACTIVE FLAGS (e.g. --yes for npx).
2. If the command would use a pager, append | cat to the command.
`,
		Name: "run_bash_script",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"script": {
					Type:        jsonschema.ParamTypeString,
					Description: "bash script",
				},
				"cwd": {
					Type:        jsonschema.ParamTypeString,
					Description: "working directory of this script",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "simple explanation of the purpose of this script",
				},
			},
			Required: []string{"script"},
		},
	}
}

// RunBashScript executes the run_bash_script tool with the given parameters
func RunBashScript(req RunBashScriptRequest) (*RunBashScriptResponse, error) {
	log.Printf("Running bash script: %s", req.Script)

	// Validate input parameters
	if req.Script == "" {
		return nil, fmt.Errorf("command is required")
	}

	startTime := time.Now()

	// Prepare response
	response := &RunBashScriptResponse{}

	// Prepare command for execution
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", req.Script)
	} else {
		cmd = exec.Command("bash", "-c", req.Script)
	}

	// Set working directory
	cmd.Dir = req.Cwd

	// Run command in foreground
	err := runBash(cmd, response)
	if err != nil {
		response.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitError.ExitCode()
		} else {
			response.ExitCode = 1
		}
	}

	// Calculate duration
	duration := time.Since(startTime)
	if duration > 1*time.Second {
		response.Duration = duration.String()
	}

	log.Printf("script completed, final response length: %d", len(response.Output))

	return response, nil
}

// runBash executes a command in the foreground and captures output
func runBash(cmd *exec.Cmd, response *RunBashScriptResponse) error {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read output
	var outputBuilder strings.Builder

	// Read stdout
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			outputBuilder.WriteString(line + "\n")
		}
	}()

	// Read stderr
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			outputBuilder.WriteString("STDERR: " + line + "\n")
		}
	}()

	done := make(chan error)
	go func() {
		// Wait for command to complete
		done <- cmd.Wait()
	}()

	var timedout bool
	var cmdErr error
	select {
	case <-time.After(30 * time.Second):
		log.Printf("script timed out after 30 seconds")
		response.Hint = appendHint(response.Hint, "script timed out after 30 seconds, try to adjust the script with smaller scope")
		timedout = true
		response.ExitCode = 1
	case cmdErr = <-done:
		// pass
	}

	// Get the output
	output := outputBuilder.String()

	origLen := len(output)
	log.Printf("script completed, output len: %d", origLen)

	contentAfterEllipse, truncated := ellipse(output, 3612)
	if truncated {
		response.Hint = appendHint(response.Hint, fmt.Sprintf("output is truncated to %d, original len %d is too large, use proper tool to iteratively inspect the content", len(contentAfterEllipse), origLen))
		log.Printf("Output is truncated to %d, original len %d is too large, use proper tool to iteratively inspect the content", len(contentAfterEllipse), origLen)
	}
	response.Output = contentAfterEllipse

	if cmdErr != nil {
		return cmdErr
	}

	if !timedout {
		response.ExitCode = 0
	}
	return nil
}

func appendHint(hint string, s string) string {
	if hint == "" {
		return s
	}
	if s == "" {
		return hint
	}
	return hint + "\n" + s
}

// getShell determines the appropriate shell to use
func getShell() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}

	// Check for SHELL environment variable
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	// Default shells by OS
	switch runtime.GOOS {
	case "darwin":
		return "/bin/bash"
	case "linux":
		return "/bin/bash"
	default:
		return "/bin/sh"
	}
}

// ValidateCommand performs basic validation on the command
func ValidateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for potentially dangerous commands
	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){ :|:& };:", // Fork bomb
		"mkfs",
		"dd if=/dev/zero",
	}

	lowerCmd := strings.ToLower(strings.TrimSpace(command))
	for _, dangerous := range dangerousCommands {
		if strings.Contains(lowerCmd, strings.ToLower(dangerous)) {
			return fmt.Errorf("potentially dangerous command detected: %s", dangerous)
		}
	}

	return nil
}

func ParseJSONRequest(jsonInput string) (RunBashScriptRequest, error) {
	var req RunBashScriptRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return RunBashScriptRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the run_bash_script tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req RunBashScriptRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Validate command
	if err := ValidateCommand(req.Script); err != nil {
		return "", err
	}

	response, err := RunBashScript(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}

// GetProcessInfo returns information about running processes
func GetProcessInfo() ([]string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "aux")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process info: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	return lines, nil
}

// KillProcess attempts to kill a process by PID
func KillProcess(pid int) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		return cmd.Run()
	} else {
		return syscall.Kill(pid, syscall.SIGTERM)
	}
}

func ellipse(msg string, maxLen int) (string, bool) {
	if len(msg) <= maxLen+3 {
		return msg, false
	}
	runes := []rune(msg)
	if len(runes) <= maxLen+3 {
		return msg, false
	}
	return string(runes[:maxLen]) + "...", true
}
