package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTreeToolDefNewFeatures(t *testing.T) {
	// Create a temporary test directory structure
	tempDir := t.TempDir()

	// Create test structure
	testDirs := []string{
		"dir1/subdir1",
		"dir1/subdir2",
		"dir1/subdir3",
		"dir2/subdir1",
		"dir2/subdir2",
		"dir3/subdir1",
		"dir4/subdir1",
		"dir5/subdir1",
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create some test files
	testFiles := []string{
		"dir1/file1.txt",
		"dir1/file2.txt",
		"dir1/subdir1/file3.txt",
		"dir2/file4.txt",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name      string
		req       TreeRequest
		checkFunc func(t *testing.T, response *TreeResponse)
	}{
		{
			name: "depth_limit",
			req: TreeRequest{
				WorkspaceRoot:         tempDir,
				RelativeWorkspacePath: ".",
				Explanation:           "Testing depth limit",
				Depth:                 1,
				MaxEntriesPerDir:      10,
			},
			checkFunc: func(t *testing.T, response *TreeResponse) {
				// Should only show directories at depth 1
				if !strings.Contains(response.Tree, "dir1") {
					t.Error("Expected dir1 in output")
				}
				// Should not show subdirectories at depth 2
				if strings.Contains(response.Tree, "subdir1") {
					t.Error("Should not show subdir1 due to depth limit")
				}
			},
		},
		{
			name: "max_entries_limit",
			req: TreeRequest{
				WorkspaceRoot:         tempDir,
				RelativeWorkspacePath: ".",
				Explanation:           "Testing max entries limit",
				Depth:                 2,
				MaxEntriesPerDir:      3,
			},
			checkFunc: func(t *testing.T, response *TreeResponse) {
				// Count the number of top-level directories shown
				lines := strings.Split(response.Tree, "\n")
				dirCount := 0
				for _, line := range lines {
					if strings.Contains(line, "dir") && !strings.Contains(line, "subdir") {
						dirCount++
					}
				}
				if dirCount > 3 {
					t.Errorf("Expected at most 3 directories, got %d", dirCount)
				}
			},
		},
		{
			name: "include_files",
			req: TreeRequest{
				WorkspaceRoot:         tempDir,
				RelativeWorkspacePath: ".",
				Explanation:           "Testing include files",
				Depth:                 2,
				MaxEntriesPerDir:      10,
				IncludeFiles:          true,
			},
			checkFunc: func(t *testing.T, response *TreeResponse) {
				// Should show files when IncludeFiles is true
				if !strings.Contains(response.Tree, "file1.txt") {
					t.Error("Expected file1.txt in output when IncludeFiles is true")
				}
			},
		},
		{
			name: "exclude_files_default",
			req: TreeRequest{
				WorkspaceRoot:         tempDir,
				RelativeWorkspacePath: ".",
				Explanation:           "Testing exclude files by default",
				Depth:                 2,
				MaxEntriesPerDir:      10,
				IncludeFiles:          false,
			},
			checkFunc: func(t *testing.T, response *TreeResponse) {
				// Should not show files when IncludeFiles is false (default)
				if strings.Contains(response.Tree, "file1.txt") {
					t.Error("Should not show file1.txt when IncludeFiles is false")
				}
			},
		},
		{
			name: "expand_dirs",
			req: TreeRequest{
				WorkspaceRoot:         tempDir,
				RelativeWorkspacePath: ".",
				Explanation:           "Testing expand directories",
				Depth:                 1,
				MaxEntriesPerDir:      10,
				ExpandDirs:            []string{"dir1"},
			},
			checkFunc: func(t *testing.T, response *TreeResponse) {
				// Should show subdirectories of dir1 even with depth limit of 1
				if !strings.Contains(response.Tree, "subdir1") {
					t.Error("Expected subdir1 in output due to expand dirs")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ExecuteTree(tt.req)
			if err != nil {
				t.Fatalf("ExecuteTree() error: %v", err)
			}

			if response == nil {
				t.Fatal("ExecuteTree() returned nil response")
			}

			if response.Tree == "" {
				t.Error("ExecuteTree() returned empty tree")
			}

			tt.checkFunc(t, response)
		})
	}
}

func TestTreeToolDefDefaults(t *testing.T) {
	tempDir := t.TempDir()

	// Create test structure
	err := os.MkdirAll(filepath.Join(tempDir, "testdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	req := TreeRequest{
		WorkspaceRoot:         tempDir,
		RelativeWorkspacePath: ".",
		Explanation:           "Testing defaults",
		// Don't set Depth and MaxEntriesPerDir to test defaults
	}

	response, err := ExecuteTree(req)
	if err != nil {
		t.Fatalf("ExecuteTree() error: %v", err)
	}

	if response == nil {
		t.Fatal("ExecuteTree() returned nil response")
	}

	if response.Tree == "" {
		t.Error("ExecuteTree() returned empty tree")
	}

	// Should contain the test directory
	if !strings.Contains(response.Tree, "testdir") {
		t.Error("Expected testdir in output")
	}
}
