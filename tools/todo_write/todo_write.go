package todo_write

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// TodoItem represents a single TODO item
type TodoItem struct {
	ID           string   `json:"id"`
	Content      string   `json:"content"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

// TodoWriteRequest represents the input parameters for the todo_write tool
type TodoWriteRequest struct {
	WorkspaceRoot string     `json:"workspace_root"`
	TodoFilePath  string     `json:"todo_file_path"`
	Merge         bool       `json:"merge"`
	Todos         []TodoItem `json:"todos"`
	Explanation   string     `json:"explanation"`
}

// TodoWriteResponse represents the output of the todo_write tool
type TodoWriteResponse struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message"`
	TodosWritten int        `json:"todos_written"`
	Todos        []TodoItem `json:"todos"`
}

// TodoList represents the structure of the todo list file
type TodoList struct {
	Version   string     `json:"version"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Todos     []TodoItem `json:"todos"`
}

// GetToolDefinition returns the JSON schema definition for the todo_write tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Use this tool to create and manage a structured task list for your current coding session. This helps track progress, organize complex tasks, and demonstrate thoroughness.

### When to Use This Tool

Use proactively for:
1. Complex multi-step tasks (3+ distinct steps)
2. Non-trivial tasks requiring careful planning
3. User explicitly requests todo list
4. User provides multiple tasks (numbered/comma-separated)
5. After receiving new instructions - capture requirements as todos (use merge=false to add new ones)
6. After completing tasks - mark complete with merge=true and add follow-ups
7. When starting new tasks - mark as in_progress (ideally only one at a time)

### When NOT to Use

Skip for:
1. Single, straightforward tasks
2. Trivial tasks with no organizational benefit
3. Tasks completable in < 3 trivial steps
4. Purely conversational/informational requests
5. Don't add a task to test the change unless asked, or you'll overfocus on testing

### Task States and Management

1. **Task States:**
  - pending: Not yet started
  - in_progress: Currently working on
  - completed: Finished successfully
  - cancelled: No longer needed

2. **Task Management:**
  - Update status in real-time
  - Mark complete IMMEDIATELY after finishing
  - Only ONE task in_progress at a time
  - Complete current tasks before starting new ones

3. **Task Breakdown:**
  - Create specific, actionable items
  - Break complex tasks into manageable steps
  - Use clear, descriptive names

4. **Task Dependencies:**
  - Use dependencies field for natural prerequisites
  - Avoid circular dependencies
  - Independent tasks can run in parallel

When in doubt, use this tool. Proactive task management demonstrates attentiveness and ensures complete requirements.`,
		Name: "todo_write",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.WORKSPACE_ROOT,
				},
				"merge": {
					Type:        jsonschema.ParamTypeBoolean,
					Description: "Whether to merge the todos with the existing todos. If true, the todos will be merged into the existing todos based on the id field. You can leave unchanged properties undefined. If false, the new todos will replace the existing todos.",
				},
				"todos": {
					Type:        jsonschema.ParamTypeArray,
					Description: "Array of TODO items to write to the workspace",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeObject,
						Properties: map[string]*jsonschema.JsonSchema{
							"id": {
								Type:        jsonschema.ParamTypeString,
								Description: "Unique identifier for the TODO item",
							},
							"content": {
								Type:        jsonschema.ParamTypeString,
								Description: "The description/content of the TODO item",
							},
							"status": {
								Type:        jsonschema.ParamTypeString,
								Description: "The current status of the TODO item",
							},
							"dependencies": {
								Type:        jsonschema.ParamTypeArray,
								Description: "List of other task IDs that are prerequisites for this task, i.e. we cannot complete this task until these tasks are done",
								Items: &jsonschema.JsonSchema{
									Type: jsonschema.ParamTypeString,
								},
							},
						},
						Required: []string{"content", "status", "id", "dependencies"},
					},
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: defs.EXPLANATION,
				},
			},
			Required: []string{"todos", "merge"},
		},
	}
}

// TodoWrite executes the todo_write tool with the given parameters
func TodoWrite(req TodoWriteRequest) (*TodoWriteResponse, error) {
	todoFilePath := req.TodoFilePath
	if req.TodoFilePath == "" {
		todoFilePath = ".llm-tools-todos.json"
	}
	todoFilePath, err := dirs.GetPath(req.WorkspaceRoot, todoFilePath, "todo_file_path", false)
	if err != nil {
		return nil, err
	}

	// Validate todo items
	for i, todo := range req.Todos {
		if todo.ID == "" {
			return &TodoWriteResponse{
				Success: false,
				Message: fmt.Sprintf("TODO item %d is missing ID", i+1),
			}, nil
		}
		if todo.Content == "" {
			return &TodoWriteResponse{
				Success: false,
				Message: fmt.Sprintf("TODO item %d is missing content", i+1),
			}, nil
		}
		if todo.Status == "" {
			return &TodoWriteResponse{
				Success: false,
				Message: fmt.Sprintf("TODO item %d is missing status", i+1),
			}, nil
		}
		// Validate status
		validStatuses := map[string]bool{
			"pending":     true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		}
		if !validStatuses[todo.Status] {
			return &TodoWriteResponse{
				Success: false,
				Message: fmt.Sprintf("TODO item %d has invalid status: %s", i+1, todo.Status),
			}, nil
		}
	}

	// Check for duplicate IDs
	idMap := make(map[string]bool)
	for i, todo := range req.Todos {
		if idMap[todo.ID] {
			return &TodoWriteResponse{
				Success: false,
				Message: fmt.Sprintf("Duplicate TODO ID found: %s (item %d)", todo.ID, i+1),
			}, nil
		}
		idMap[todo.ID] = true
	}

	// Validate dependencies
	for i, todo := range req.Todos {
		for _, depID := range todo.Dependencies {
			if depID == todo.ID {
				return &TodoWriteResponse{
					Success: false,
					Message: fmt.Sprintf("TODO item %d has circular dependency on itself", i+1),
				}, nil
			}
		}
	}

	var todoList TodoList
	now := time.Now().Format(time.RFC3339)

	if req.Merge {
		// Load existing todos if they exist
		if _, err := os.Stat(todoFilePath); err == nil {
			data, err := os.ReadFile(todoFilePath)
			if err != nil {
				return &TodoWriteResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to read existing todo file: %v", err),
				}, nil
			}

			if err := json.Unmarshal(data, &todoList); err != nil {
				return &TodoWriteResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to parse existing todo file: %v", err),
				}, nil
			}
		} else {
			// Initialize new todo list
			todoList = TodoList{
				Version:   "1.0",
				CreatedAt: now,
				Todos:     []TodoItem{},
			}
		}

		// Merge todos
		existingTodos := make(map[string]*TodoItem)
		for i := range todoList.Todos {
			existingTodos[todoList.Todos[i].ID] = &todoList.Todos[i]
		}

		for _, newTodo := range req.Todos {
			// Set timestamps
			if newTodo.CreatedAt == "" {
				newTodo.CreatedAt = now
			}
			newTodo.UpdatedAt = now

			if existing, exists := existingTodos[newTodo.ID]; exists {
				// Update existing todo
				if newTodo.Content != "" {
					existing.Content = newTodo.Content
				}
				if newTodo.Status != "" {
					existing.Status = newTodo.Status
				}
				if newTodo.Dependencies != nil {
					existing.Dependencies = newTodo.Dependencies
				}
				existing.UpdatedAt = now
			} else {
				// Add new todo
				todoList.Todos = append(todoList.Todos, newTodo)
			}
		}
	} else {
		// Replace all todos
		for i := range req.Todos {
			if req.Todos[i].CreatedAt == "" {
				req.Todos[i].CreatedAt = now
			}
			req.Todos[i].UpdatedAt = now
		}

		todoList = TodoList{
			Version:   "1.0",
			CreatedAt: now,
			Todos:     req.Todos,
		}
	}

	todoList.UpdatedAt = now

	// Write to file
	data, err := json.MarshalIndent(todoList, "", "  ")
	if err != nil {
		return &TodoWriteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to marshal todo list: %v", err),
		}, nil
	}

	dir := filepath.Dir(todoFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &TodoWriteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create directory: %v", err),
		}, nil
	}

	if err := os.WriteFile(todoFilePath, data, 0644); err != nil {
		return &TodoWriteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to write todo file: %v", err),
		}, nil
	}

	return &TodoWriteResponse{
		Success:      true,
		Message:      "TODO list updated successfully",
		TodosWritten: len(req.Todos),
		Todos:        todoList.Todos,
	}, nil
}

func ParseJSONRequest(jsonInput string) (TodoWriteRequest, error) {
	var req TodoWriteRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return TodoWriteRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the todo_write tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req TodoWriteRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := TodoWrite(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
