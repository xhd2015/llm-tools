package batch_read_file

import (
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
llm-tools batch_read_file reads multiple files in a batch operation

Usage: llm-tools batch_read_file [OPTIONS]

Options:
  --workspace-root <path>      workspace root directory (defaults to current directory)
  --file <path>                file to read (can be specified multiple times)
  --entire-file                read entire files
  --start-line <num>           start line number (1-indexed)
  --end-line <num>             end line number (1-indexed, inclusive)
  --max-lines <num>            maximum lines per file
  --global-max-lines <num>     global maximum lines per file (default: 250)
  --global-min-lines <num>     global minimum lines per file (default: 200)
  --continue-on-error          continue processing if one file fails
  --include-outline            include outline in responses
  --explanation <text>         explanation for the operation

Examples:
  llm-tools batch_read_file --file file1.go --file file2.go --entire-file
  llm-tools batch_read_file --workspace-root /path/to/workspace --file file1.go --start-line 1 --end-line 50
`

func HandleCli(args []string) error {
	var workspaceRoot string
	var files []string
	var entireFile bool
	var startLine int
	var endLine int
	var maxLines int
	var globalMaxLines int
	var globalMinLines int
	var continueOnError bool
	var includeOutline bool
	var explanation string

	args, err := flags.String("--workspace-root", &workspaceRoot).
		StringSlice("--file", &files).
		Bool("--entire-file", &entireFile).
		Int("--start-line", &startLine).
		Int("--end-line", &endLine).
		Int("--max-lines", &maxLines).
		Int("--global-max-lines", &globalMaxLines).
		Int("--global-min-lines", &globalMinLines).
		Bool("--continue-on-error", &continueOnError).
		Bool("--include-outline", &includeOutline).
		String("--explanation", &explanation).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments")
	}

	if len(files) == 0 {
		return fmt.Errorf("at least one file must be specified with --file")
	}

	// Use current working directory if workspace_root is not provided
	if workspaceRoot == "" {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Set defaults
	if globalMaxLines == 0 {
		globalMaxLines = 250
	}
	if globalMinLines == 0 {
		globalMinLines = 200
	}

	// Build file requests
	var fileRequests []FileReadRequest
	for _, file := range files {
		fileReq := FileReadRequest{
			TargetFile:           file,
			ShouldReadEntireFile: entireFile,
		}

		if !entireFile {
			if startLine > 0 {
				fileReq.StartLineOneIndexed = startLine
			}
			if endLine > 0 {
				fileReq.EndLineOneIndexedInclusive = endLine
			}
		}

		if maxLines > 0 {
			fileReq.MaxLines = maxLines
		}

		fileRequests = append(fileRequests, fileReq)
	}

	req := BatchReadFileRequest{
		WorkspaceRoot:   workspaceRoot,
		Files:           fileRequests,
		GlobalMaxLines:  globalMaxLines,
		GlobalMinLines:  globalMinLines,
		ContinueOnError: continueOnError,
		IncludeOutline:  includeOutline,
		Explanation:     explanation,
	}

	response, err := BatchReadFile(req)
	if err != nil {
		return err
	}

	// Print results
	fmt.Printf("Total files: %d, Success: %d, Errors: %d\n", response.TotalFiles, response.SuccessCount, response.ErrorCount)
	fmt.Println()

	for _, fileResp := range response.Files {
		fmt.Printf("=== %s ===\n", fileResp.TargetFile)
		if fileResp.Error != "" {
			fmt.Printf("Error: %s\n", fileResp.Error)
		} else {
			fmt.Printf("Lines: %s (Total: %d)\n", fileResp.LinesShown, fileResp.TotalLines)
			if fileResp.Outline != "" {
				fmt.Printf("Outline: %s\n", fileResp.Outline)
			}
			fmt.Println(fileResp.Contents)
		}
		fmt.Println()
	}

	return nil
}
