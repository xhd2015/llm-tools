package run_terminal_cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// RunTerminalCmdRequest represents the input parameters for the run_terminal_cmd tool
type RunTerminalCmdRequest struct {
	Command      string `json:"command"`
	IsBackground bool   `json:"is_background"`
	Explanation  string `json:"explanation,omitempty"`
}

// RunTerminalCmdResponse represents the output of the run_terminal_cmd tool
type RunTerminalCmdResponse struct {
	ExitCode      int    `json:"exit_code"`
	CommandOutput string `json:"command_output"`
	ShellInfo     string `json:"shell_info"`
	Command       string `json:"command"`
	IsBackground  bool   `json:"is_background"`
	WorkingDir    string `json:"working_dir"`
	Duration      string `json:"duration,omitempty"`
	Error         string `json:"error,omitempty"`
}

// GetToolDefinition returns the JSON schema definition for the run_terminal_cmd tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `PROPOSE a command to run on behalf of the user.
If you have this tool, note that you DO have the ability to run commands directly on the USER's system.
Note that the user will have to approve the command before it is executed.
The user may reject it if it is not to their liking, or may modify the command before approving it.  If they do change it, take those changes into account.
The actual command will NOT execute until the user approves it. The user may not approve it immediately. Do NOT assume the command has started running.
If the step is WAITING for user approval, it has NOT started running.
In using these tools, adhere to the following guidelines:
1. Based on the contents of the conversation, you will be told if you are in the same shell as a previous step or a different shell.
2. If in a new shell, you should cd to the appropriate directory and do necessary setup in addition to running the command.
3. If in the same shell, LOOK IN CHAT HISTORY for your current working directory.
4. For ANY commands that would require user interaction, ASSUME THE USER IS NOT AVAILABLE TO INTERACT and PASS THE NON-INTERACTIVE FLAGS (e.g. --yes for npx).
5. If the command would use a pager, append | cat to the command.
6. For commands that are long running/expected to run indefinitely until interruption, please run them in the background. To run jobs in the background, set is_background to true rather than changing the details of the command.
7. Dont include any newlines in the command.`,
		Name: "run_terminal_cmd",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"command": {
					Type:        jsonschema.ParamTypeString,
					Description: "The terminal command to execute",
				},
				"is_background": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether the command should be run in the background",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this command needs to be run and how it contributes to the goal.",
				},
			},
			Required: []string{"command", "is_background"},
		},
	}
}

// RunTerminalCmd executes the run_terminal_cmd tool with the given parameters
func RunTerminalCmd(req RunTerminalCmdRequest) (*RunTerminalCmdResponse, error) {
	// Validate input parameters
	if req.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "unknown"
	}

	startTime := time.Now()

	// Prepare response
	response := &RunTerminalCmdResponse{
		Command:      req.Command,
		IsBackground: req.IsBackground,
		WorkingDir:   workingDir,
	}

	// Determine shell to use
	shell := getShell()

	// Prepare command for execution
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", req.Command)
	} else {
		cmd = exec.Command(shell, "-c", req.Command)
	}

	// Set working directory
	cmd.Dir = workingDir

	if req.IsBackground {
		// Run command in background
		err := runBackgroundCommand(cmd, response)
		if err != nil {
			response.Error = err.Error()
			response.ExitCode = 1
		}
	} else {
		// Run command in foreground
		err := runForegroundCommand(cmd, response)
		if err != nil {
			response.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				response.ExitCode = exitError.ExitCode()
			} else {
				response.ExitCode = 1
			}
		}
	}

	// Calculate duration
	duration := time.Since(startTime)
	response.Duration = duration.String()

	// Generate shell info
	response.ShellInfo = generateShellInfo(workingDir, shell, req.IsBackground)

	return response, nil
}

// runForegroundCommand executes a command in the foreground and captures output
func runForegroundCommand(cmd *exec.Cmd, response *RunTerminalCmdResponse) error {
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

	// Wait for command to complete
	err = cmd.Wait()

	// Get the output
	response.CommandOutput = strings.TrimRight(outputBuilder.String(), "\n")

	if err != nil {
		return err
	}

	response.ExitCode = 0
	return nil
}

// runBackgroundCommand starts a command in the background
func runBackgroundCommand(cmd *exec.Cmd, response *RunTerminalCmdResponse) error {
	// For background commands, we start them and return immediately
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background command: %w", err)
	}

	response.ExitCode = 0
	response.CommandOutput = fmt.Sprintf("Background command started with PID: %d", cmd.Process.Pid)

	// Start a goroutine to wait for the process (but don't block)
	go func() {
		cmd.Wait()
	}()

	return nil
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

// generateShellInfo creates informational text about the shell state
func generateShellInfo(workingDir, shell string, isBackground bool) string {
	var info strings.Builder

	if isBackground {
		info.WriteString("Background command executed")
	} else {
		info.WriteString("Command completed")
	}

	info.WriteString(fmt.Sprintf(", shell: %s", filepath.Base(shell)))
	info.WriteString(fmt.Sprintf(", directory: %s", workingDir))

	return info.String()
}

// RunTerminalCmdWithTimeout executes a command with a timeout
func RunTerminalCmdWithTimeout(req RunTerminalCmdRequest, timeout time.Duration) (*RunTerminalCmdResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "unknown"
	}

	startTime := time.Now()

	// Prepare response
	response := &RunTerminalCmdResponse{
		Command:      req.Command,
		IsBackground: req.IsBackground,
		WorkingDir:   workingDir,
	}

	// Determine shell to use
	shell := getShell()

	// Prepare command for execution with context
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(ctx, shell, "-c", req.Command)
	}

	// Set working directory
	cmd.Dir = workingDir

	// Run command (background not supported with timeout)
	output, err := cmd.CombinedOutput()
	response.CommandOutput = string(output)

	// Calculate duration
	duration := time.Since(startTime)
	response.Duration = duration.String()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			response.Error = fmt.Sprintf("command timed out after %v", timeout)
			response.ExitCode = 124 // Standard timeout exit code
		} else {
			response.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				response.ExitCode = exitError.ExitCode()
			} else {
				response.ExitCode = 1
			}
		}
	} else {
		response.ExitCode = 0
	}

	// Generate shell info
	response.ShellInfo = generateShellInfo(workingDir, shell, req.IsBackground)

	return response, nil
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

func ParseJSONRequest(jsonInput string) (RunTerminalCmdRequest, error) {
	var req RunTerminalCmdRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return RunTerminalCmdRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the run_terminal_cmd tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req RunTerminalCmdRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Validate command
	if err := ValidateCommand(req.Command); err != nil {
		return "", err
	}

	response, err := RunTerminalCmd(req)
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
