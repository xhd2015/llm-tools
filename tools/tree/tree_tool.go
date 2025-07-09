package tree

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools tree prints the tree of a directory

Usage: llm-tools tree <dir> [OPTIONS]

Options:
  --exclude <pattern>  exclude files/directories matching the pattern
  --include <pattern>  include files/directories matching the pattern
  --dir-only           only show directories
  --collapse-pattern   collapse duplicate patterns
  --collapse-repeated  collapse repeated entries
  --collapse           collapse both patterns and repeated entries
  --find-path <path>   find the path in the tree

Examples:
  llm-tools tree                       current directory
  llm-tools tree --exclude .git        exclude .git directory
`

func HandleCli(args []string) error {
	var collapsePattern bool
	var collapseRepeated bool
	var collapse bool

	var collapseDir []string

	var exclude []string
	var include []string
	var dirOnly bool

	var findPath string
	args, err := flags.Bool("--collapse-pattern", &collapsePattern).
		Bool("--collapse-repeated", &collapseRepeated).
		Bool("--collapse", &collapse).
		StringSlice("--collapse-dir", &collapseDir).
		StringSlice("--exclude", &exclude).
		StringSlice("--include", &include).
		Bool("--dir-only", &dirOnly).
		String("--find-path", &findPath).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	var dir string
	if len(args) > 0 {
		dir = args[0]
		args = args[1:]

		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra")
		}
	} else {
		dir = "."
	}

	if collapse {
		collapseRepeated = true
		collapsePattern = true
	}

	item, err := traverseTreeItem(dir, TreeOptions{
		IncludePatterns:  include,
		ExcludePatterns:  exclude,
		DirectoriesOnly:  dirOnly,
		CollapseRepeated: collapseRepeated,
		CollapsePattern:  collapsePattern,
		CollapsedDirs:    collapseDir,
	})
	if err != nil {
		return err
	}

	if findPath != "" {
		splittedPath, err := splitPath(findPath)
		if err != nil {
			return err
		}
		if len(splittedPath) == 0 {
			return fmt.Errorf("path is empty")
		}

		paths := findPathInTree([]string{}, item, splittedPath)
		if len(paths) == 0 {
			return fmt.Errorf("path not found: %s", findPath)
		}

		for _, path := range paths {
			fmt.Println(filepath.Join(path...))
		}
		return nil
	}

	output := PrintItem(item)
	fmt.Println(output)
	return nil
}

func splitPath(path string) ([]string, error) {
	parts := strings.Split(path, "/")

	j := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		parts[j] = part
		j++
	}
	return parts[:j], nil
}

type TreeToolRequest struct {
	WorkspaceRoot         string `json:"workspace_root"`
	RelativeWorkspacePath string `json:"relative_workspace_path"`
	Explanation           string `json:"explanation"`
}

type TreeToolResponse struct {
	Tree string `json:"tree"`
}

func (c TreeToolResponse) ToLLMOutput() string {
	//

	return ""
}

func TreeTool(req TreeToolRequest) (TreeToolResponse, error) {
	// tree, err := TreeCollapsed(req.WorkspaceRoot, req.RelativeWorkspacePath)
	// if err != nil {
	// 	return TreeToolResponse{}, err
	// }
	// return TreeToolResponse{Tree: tree}, nil

	// TODO
	return TreeToolResponse{}, nil
}
