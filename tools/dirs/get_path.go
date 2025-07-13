package dirs

import (
	"fmt"
	"path/filepath"
)

func GetPath(workspaceRoot string, relativeFile string, relativeFileName string, defaultRoot bool) (string, error) {
	// Validate input parameters
	if relativeFile == "" {
		if defaultRoot {
			if workspaceRoot == "" {
				return "", fmt.Errorf("requires workspace_root when %s is a relative path", relativeFileName)
			}
			return workspaceRoot, nil
		}
		return "", fmt.Errorf("requires %s", relativeFileName)
	}

	filePath := relativeFile
	if !filepath.IsAbs(filePath) {
		if workspaceRoot == "" {
			return "", fmt.Errorf("requires workspace_root when %s is a relative path", relativeFileName)
		}
		filePath = filepath.Join(workspaceRoot, filePath)
	}
	return filePath, nil
}
