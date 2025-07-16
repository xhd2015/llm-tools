package delete_file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
	"github.com/xhd2015/llm-tools/tools/dirs"
)

// DeleteFileRequest represents the input parameters for the delete_file tool
type DeleteFileRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	TargetFile    string `json:"target_file"`
	Explanation   string `json:"explanation"`
}

// DeleteFileResponse represents the output of the delete_file tool
type DeleteFileResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	DeletedFile string `json:"deleted_file"`
}

// GetToolDefinition returns the JSON schema definition for the delete_file tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: "Deletes a file at the specified path. The operation will fail gracefully if:\n    - The file doesn't exist\n    - The operation is rejected for security reasons\n    - The file cannot be deleted",
		Name:        "delete_file",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"workspace_root": {
					Type:        jsonschema.ParamTypeString,
					Description: "The absolute path of the workspace root directory. This is used to resolve relative paths to files.",
				},
				"target_file": {
					Type:        jsonschema.ParamTypeString,
					Description: "The path of the file to delete, relative to the workspace root.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"target_file"},
		},
	}
}

// DeleteFile executes the delete_file tool with the given parameters
func DeleteFile(req DeleteFileRequest) (*DeleteFileResponse, error) {
	if req.WorkspaceRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		req.WorkspaceRoot = cwd
	}

	// Resolve the target file path
	filePath, err := dirs.GetPath(req.WorkspaceRoot, req.TargetFile, "target_file", false)
	if err != nil {
		return &DeleteFileResponse{
			Success:     false,
			Message:     fmt.Sprintf("Invalid file path: %v", err),
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Security check: ensure the file is within the workspace
	absWorkspaceRoot, err := filepath.Abs(req.WorkspaceRoot)
	if err != nil {
		return &DeleteFileResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to resolve workspace root: %v", err),
			DeletedFile: req.TargetFile,
		}, nil
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return &DeleteFileResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to resolve file path: %v", err),
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Check if the file is within the workspace
	if !strings.HasPrefix(absFilePath, absWorkspaceRoot) {
		return &DeleteFileResponse{
			Success:     false,
			Message:     "Security violation: file is outside workspace root",
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return &DeleteFileResponse{
			Success:     false,
			Message:     "File does not exist",
			DeletedFile: req.TargetFile,
		}, nil
	}
	if err != nil {
		return &DeleteFileResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to stat file: %v", err),
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return &DeleteFileResponse{
			Success:     false,
			Message:     "Cannot delete directory (use appropriate directory deletion tool)",
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Additional security checks for sensitive files
	if isSensitiveFile(filePath) {
		return &DeleteFileResponse{
			Success:     false,
			Message:     "Security violation: cannot delete sensitive system files",
			DeletedFile: req.TargetFile,
		}, nil
	}

	// Attempt to delete the file
	err = os.Remove(filePath)
	if err != nil {
		return &DeleteFileResponse{
			Success:     false,
			Message:     fmt.Sprintf("Failed to delete file: %v", err),
			DeletedFile: req.TargetFile,
		}, nil
	}

	return &DeleteFileResponse{
		Success:     true,
		Message:     "File deleted successfully",
		DeletedFile: req.TargetFile,
	}, nil
}

// isSensitiveFile checks if a file is considered sensitive and should not be deleted
func isSensitiveFile(filePath string) bool {
	// Convert to absolute path for consistent checking
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return true // Err on the side of caution
	}

	// Check for sensitive file patterns
	sensitivePatterns := []string{
		// System files
		"/etc/",
		"/usr/",
		"/var/",
		"/bin/",
		"/sbin/",
		"/lib/",
		"/lib64/",
		"/boot/",
		"/dev/",
		"/proc/",
		"/sys/",
		// Windows system files
		"C:\\Windows\\",
		"C:\\Program Files\\",
		"C:\\Program Files (x86)\\",
		// Important config files
		"/.ssh/",
		"/.aws/",
		"/.docker/",
		"/.kube/",
		// Version control
		"/.git/config",
		"/.gitconfig",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(absPath, pattern) {
			return true
		}
	}

	// Check for sensitive file extensions
	ext := strings.ToLower(filepath.Ext(absPath))
	sensitiveExts := []string{
		".key", ".pem", ".crt", ".p12", ".pfx", ".jks",
		".keystore", ".truststore", ".ssh", ".gpg",
	}

	for _, sensitiveExt := range sensitiveExts {
		if ext == sensitiveExt {
			return true
		}
	}

	// Check for sensitive filenames
	filename := strings.ToLower(filepath.Base(absPath))
	sensitiveNames := []string{
		"id_rsa", "id_dsa", "id_ecdsa", "id_ed25519",
		"private.key", "server.key", "client.key",
		"credentials", "secrets", "password", "passwd",
		".env", ".env.local", ".env.production",
		"config.json", "settings.json", "secrets.json",
	}

	for _, sensitiveName := range sensitiveNames {
		if filename == sensitiveName {
			return true
		}
	}

	return false
}

func ParseJSONRequest(jsonInput string) (DeleteFileRequest, error) {
	var req DeleteFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return DeleteFileRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the delete_file tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req DeleteFileRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := DeleteFile(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
