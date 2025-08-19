# llm tools

This is a collection of llm tools, including the ones reverse engineered from Cursor.

## Implemented Tools

|Tool|Args|Description|
|-|-|-|
|`read_file`|`target_file`, `should_read_entire_file`, `start_line_one_indexed`, `end_line_one_indexed_inclusive`, `explanation`|Read the contents of a file with line range support. Supports reading entire files or specific line ranges (max 250 lines, min 200 lines for partial reads). Returns structured output with file contents, total lines, lines shown, and code outline.|
|`batch_read_file`|`files[]`, `global_max_lines`, `global_min_lines`, `continue_on_error`, `include_outline`, `explanation`|Read multiple files in a single batch operation for improved efficiency. Each file can have individual line range settings. Supports global and per-file line limits, error handling, and optional outline generation.|
|`grep_search`|`query`, `case_sensitive`, `exclude_pattern`, `include_pattern`, `explanation`|Fast regex search over text files using the ripgrep engine. Results are capped at 50 matches. Supports include/exclude patterns for file filtering and case-sensitive/insensitive search.|
|`run_terminal_cmd`|`command`, `is_background`, `explanation`|Execute terminal commands on behalf of the user. Supports foreground and background execution, cross-platform shell detection, output capture, and safety validation. Returns exit codes, command output, and execution context.|

## Tool Features

### `read_file`
- **Line Range Control**: Read specific line ranges with 1-indexed line numbers
- **Entire File Support**: Option to read complete files
- **Smart Context**: Automatically expands ranges to meet minimum requirements
- **Multi-Language Outlines**: Generates code outlines for Go, JavaScript/TypeScript, Python, Java, C/C++
- **Structured Output**: Returns contents, total lines, lines shown, and outline

### `batch_read_file`
- **Batch Processing**: Read multiple files in a single operation
- **Individual Control**: Each file can have its own line range settings
- **Global Settings**: Set global max/min lines with per-file overrides
- **Error Resilience**: Continue processing other files if one fails
- **Summary Statistics**: Returns total files, success count, and error count

### `grep_search`
- **Ripgrep Integration**: Uses the fast ripgrep engine for searching
- **Regex Support**: Full regex pattern matching with proper escaping
- **File Filtering**: Include/exclude patterns using glob syntax
- **Rich Results**: Returns file path, line number, column, content, and match positions
- **Fallback Parser**: JSON-based parsing with text fallback

### `run_terminal_cmd`
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Background Support**: Run long-running commands in background
- **Shell Detection**: Automatically uses appropriate shell (bash, cmd, sh)
- **Output Capture**: Captures both stdout and stderr
- **Safety Features**: Basic validation to prevent dangerous commands
- **Process Management**: Utilities for process information and control

# Docs
https://gist.github.com/sshh12/25ad2e40529b269a88b80e7cf1c38084

https://blog.sshh.io/p/how-cursor-ai-ide-works

https://gist.github.com/ichim-david/bf24513616aa7dc5c74abcae35ddf706

# Some interesting prompt to the tool
Cursor: (by asking in the chat)

```
You have both the edit_file and search_replace tools at your disposal. Use the search_replace tool for files larger than 2500 lines, otherwise prefer the edit_file tool.
```