# LLM Tools MCP Server

This is a Model Context Protocol (MCP) server that exposes the `batch_read_file` tool from the llm-tools project.

## Overview

The MCP server provides a `batch_read_file` tool that allows efficient reading of multiple files in a single operation, which is much more efficient than making multiple separate `read_file` calls.

## Building and Running

### Build the server:
```bash
go build -o bin/mcp-server script/mcp/main.go
```

### Run the server:
```bash
./bin/mcp-server
```

The server communicates via stdin/stdout using the JSON-RPC protocol.

## Tool: batch_read_file

### Description
Read the contents of multiple files in a single batch operation. This tool improves efficiency by reading multiple files at once instead of making separate read_file calls.

### Parameters

The tool accepts the following parameters:

- `explanation` (required, string): One sentence explanation as to why this tool is being used, and how it contributes to the goal.
- `files` (required, array): Array of file read requests, each with its own target file and line range settings
- `global_max_lines` (optional, number): Global maximum lines per file (default: 250). Can be overridden per file.
- `global_min_lines` (optional, number): Global minimum lines per file for partial reads (default: 200). Applied when expanding ranges.
- `continue_on_error` (optional, boolean): Whether to continue processing other files if one fails (default: true).
- `include_outline` (optional, boolean): Whether to include outline in responses (default: true).

### File Request Structure

Each file in the `files` array should have:

- `target_file` (required, string): The path of the file to read. You can use either a relative path in the workspace or an absolute path.
- `should_read_entire_file` (optional, boolean): Whether to read the entire file. When true, line range parameters are ignored. When false, line range parameters are required. Defaults to false.
- `start_line_one_indexed` (optional, number): The one-indexed line number to start reading from (inclusive). Required when should_read_entire_file is false.
- `end_line_one_indexed_inclusive` (optional, number): The one-indexed line number to end reading at (inclusive). Required when should_read_entire_file is false.
- `max_lines` (optional, number): Optional per-file maximum lines limit. Overrides global_max_lines for this file.

### Usage Examples

#### Example 1: Read entire files
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "batch_read_file",
    "arguments": {
      "explanation": "Reading configuration and source files to understand the project structure",
      "files": [
        {
          "target_file": "go.mod",
          "should_read_entire_file": true
        },
        {
          "target_file": "main.go",
          "should_read_entire_file": true
        }
      ],
      "continue_on_error": true,
      "include_outline": true
    }
  }
}
```

#### Example 2: Read specific line ranges
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "batch_read_file",
    "arguments": {
      "explanation": "Reading the first 20 lines of multiple files to understand their structure",
      "files": [
        {
          "target_file": "README.md",
          "should_read_entire_file": false,
          "start_line_one_indexed": 1,
          "end_line_one_indexed_inclusive": 20
        },
        {
          "target_file": "script/mcp/main.go",
          "should_read_entire_file": false,
          "start_line_one_indexed": 1,
          "end_line_one_indexed_inclusive": 50
        }
      ],
      "global_max_lines": 100,
      "continue_on_error": true,
      "include_outline": false
    }
  }
}
```

### Key Features

- **Batch processing**: Read multiple files in a single operation
- **Individual line range control**: Each file can have its own line range settings
- **Global and per-file line limits**: Control output size with flexible limits
- **Error handling**: Continue processing other files if one fails
- **Optional outline generation**: Get code structure summaries
- **Structured response**: Includes success/error counts and detailed results

### Parameter Usage Guidelines

- Set `should_read_entire_file=true` to read the entire file (line range parameters are ignored)
- Set `should_read_entire_file=false` and provide `start_line_one_indexed` and `end_line_one_indexed_inclusive` for specific ranges
- **DO NOT** provide both `should_read_entire_file=true` AND line range parameters as this creates ambiguity

### Use Cases

This tool is particularly useful when you need to read multiple related files, such as:
- Examining imports across multiple source files
- Comparing implementations between different files
- Gathering context from multiple configuration files
- Reading test files alongside their corresponding source files

## Integration with MCP Clients

To use this server with MCP-compatible clients, configure it as a stdio server. For example, with Claude Desktop or other MCP hosts:

```json
{
  "mcpServers": {
    "llm-tools-batch-read": {
      "command": "/path/to/bin/mcp-server",
      "args": []
    }
  }
}
```

## Dependencies

- Go 1.21 or later
- github.com/mark3labs/mcp-go v0.32.0
- github.com/xhd2015/llm-tools (local dependency) 



## MCP Server

This project includes a Model Context Protocol (MCP) server that exposes the `batch_read_file` tool for use with MCP-compatible clients like Claude Desktop.

### Quick Start

Build and run the MCP server:

```bash
# Build the server
go build -o bin/mcp-server script/mcp/main.go

# Run the server (communicates via stdin/stdout)
./bin/mcp-server

# Test the server
./script/mcp/test.sh
```

### Integration with MCP Clients

Configure the server in your MCP client (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "llm-tools-batch-read": {
      "command": "/path/to/llm-tools/bin/mcp-server",
      "args": []
    }
  }
}
```

### Features

- **Standards Compliant**: Implements the Model Context Protocol specification
- **Efficient File Reading**: Exposes the powerful `batch_read_file` tool
- **Error Handling**: Provides helpful error messages and validation
- **JSON-RPC Protocol**: Uses standard JSON-RPC 2.0 for communication
- **Tool Discovery**: Supports MCP tool listing and introspection

For detailed documentation, see [script/mcp/README.md](script/mcp/README.md).
