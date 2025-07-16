package todo_write

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools todo_write creates and manages structured task lists for coding sessions

Usage: llm-tools todo_write [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --merge                      merge with existing todos (default: false)
  --todo-file-path <path>      path to the todo file (default: .llm-tools-todos.json)
  --todo <id:content:status>   add a todo item (can be used multiple times)
  --explanation <text>         explanation for the operation

Examples:
  llm-tools todo_write --todo "task1:Implement user auth:pending" --todo "task2:Write tests:pending"
  llm-tools todo_write --merge --todo "task1:Implement user auth:completed"
  llm-tools todo_write --workspace-root /path/to/workspace --todo "task1:Setup database:in_progress"
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var todoFilePath string
	var merge bool
	var todoStrings []string
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		Bool("--merge", &merge).
		String("--todo-file-path", &todoFilePath).
		StringSlice("--todo", &todoStrings).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args, ","))
	}

	if len(todoStrings) == 0 {
		return fmt.Errorf("at least one --todo is required")
	}

	if len(todoStrings) < 2 {
		return fmt.Errorf("at least 2 TODO items are required")
	}

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Parse todo strings
	var todos []TodoItem
	for i, todoStr := range todoStrings {
		parts := strings.SplitN(todoStr, ":", 3)
		if len(parts) != 3 {
			return fmt.Errorf("invalid todo format for item %d: %s (expected id:content:status)", i+1, todoStr)
		}

		id := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])
		status := strings.TrimSpace(parts[2])

		if id == "" || content == "" || status == "" {
			return fmt.Errorf("invalid todo format for item %d: id, content, and status cannot be empty", i+1)
		}

		// Validate status
		validStatuses := map[string]bool{
			"pending":     true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid status for item %d: %s (must be one of: pending, in_progress, completed, cancelled)", i+1, status)
		}

		todos = append(todos, TodoItem{
			ID:           id,
			Content:      content,
			Status:       status,
			Dependencies: []string{}, // CLI doesn't support dependencies for simplicity
		})
	}

	req := TodoWriteRequest{
		WorkspaceRoot: workspaceRoot,
		TodoFilePath:  todoFilePath,
		Merge:         merge,
		Todos:         todos,
		Explanation:   explanation,
	}

	response, err := TodoWrite(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("TODOs written: %d\n", response.TodosWritten)
	fmt.Println()

	if response.Success && len(response.Todos) > 0 {
		fmt.Println("Current TODO list:")
		for _, todo := range response.Todos {
			fmt.Printf("  [%s] %s (%s)\n", todo.Status, todo.Content, todo.ID)
			if len(todo.Dependencies) > 0 {
				fmt.Printf("    Dependencies: %s\n", strings.Join(todo.Dependencies, ", "))
			}
		}
	}

	if !response.Success {
		os.Exit(1)
	}

	return nil
}
