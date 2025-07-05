# Tool Definitions V2 - Complete Available Tools

This document contains all available tools and their JSON schema definitions, following the format established in TOOL_DEFS_JSON_SCHEMA.md.

## Available Tools (15 total)

### 1. codebase_search
```json
{
  "description": "`codebase_search`: semantic search that finds code by meaning, not exact text\n\n### When to Use This Tool\n\nUse `codebase_search` when you need to:\n- Explore unfamiliar codebases\n- Ask \"how / where / what\" questions to understand behavior\n- Find code by meaning rather than exact text\n\n### When NOT to Use\n\nSkip `codebase_search` for:\n1. Exact text matches (use `grep_search`)\n2. Reading known files (use `read_file`)\n3. Simple symbol lookups (use `grep_search`)\n4. Find file by name (use `file_search`)\n\n### Examples\n\n<example>\n  Query: \"Where is interface MyInterface implemented in the frontend?\"\n\n<reasoning>\n  Good: Complete question asking about implementation location with specific context (frontend).\n</reasoning>\n</example>\n\n<example>\n  Query: \"Where do we encrypt user passwords before saving?\"\n\n<reasoning>\n  Good: Clear question about a specific process with context about when it happens.\n</reasoning>\n</example>\n\n<example>\n  Query: \"MyInterface frontend\"\n\n<reasoning>\n  BAD: Too vague; use a specific question instead. This would be better as \"Where is MyInterface used in the frontend?\"\n</reasoning>\n</example>\n\n<example>\n  Query: \"AuthService\"\n\n<reasoning>\n  BAD: Single word searches should use `grep_search` for exact text matching instead.\n</reasoning>\n</example>\n\n<example>\n  Query: \"What is AuthService? How does AuthService work?\"\n\n<reasoning>\n  BAD: Combines two separate queries together. Semantic search is not good at looking for multiple things in parallel. Split into separate searches: first \"What is AuthService?\" then \"How does AuthService work?\"\n</reasoning>\n</example>\n\n### Target Directories\n\n- Provide ONE directory or file path; [] searches the whole repo. No globs or wildcards.\n  Good:\n  - [\"backend/api/\"]   - focus directory\n  - [\"src/components/Button.tsx\"] - single file\n  - [] - search everywhere when unsure\n  BAD:\n  - [\"frontend/\", \"backend/\"] - multiple paths\n  - [\"src/**/utils/**\"] - globs\n  - [\"*.ts\"] or [\"**/*\"] - wildcard paths\n\n### Search Strategy\n\n1. Start with exploratory queries - semantic search is powerful and often finds relevant context in one go. Begin broad with [].\n2. Review results; if a directory or file stands out, rerun with that as the target.\n3. Break large questions into smaller ones (e.g. auth roles vs session storage).\n4. For big files (>1K lines) run `codebase_search` scoped to that file instead of reading the entire file.\n\n<example>\n  Step 1: { \"query\": \"How does user authentication work?\", \"target_directories\": [], \"explanation\": \"Find auth flow\" }\n  Step 2: Suppose results point to backend/auth/ → rerun:\n          { \"query\": \"Where are user roles checked?\", \"target_directories\": [\"backend/auth/\"], \"explanation\": \"Find role logic\" }\n\n<reasoning>\n  Good strategy: Start broad to understand overall system, then narrow down to specific areas based on initial results.\n</reasoning>\n</example>\n\n<example>\n  Query: \"How are websocket connections handled?\"\n  Target: [\"backend/services/realtime.ts\"]\n\n<reasoning>\n  Good: We know the answer is in this specific file, but the file is too large to read entirely, so we use semantic search to find the relevant parts.\n</reasoning>\n</example>",
  "name": "codebase_search",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "query": {
        "description": "A complete question about what you want to understand. Ask as if talking to a colleague: 'How does X work?', 'What happens when Y?', 'Where is Z handled?'",
        "type": "string"
      },
      "target_directories": {
        "description": "Prefix directory paths to limit search scope (single directory only, no glob patterns)",
        "items": {
          "type": "string"
        },
        "type": "array"
      }
    },
    "required": ["explanation", "query", "target_directories"],
    "type": "object"
  }
}
```

### 2. read_file
```json
{
  "description": "Read the contents of a file. the output of this tool call will be the 1-indexed file contents from start_line_one_indexed to end_line_one_indexed_inclusive, together with a summary of the lines outside start_line_one_indexed and end_line_one_indexed_inclusive.\nNote that this call can view at most 250 lines at a time and 200 lines minimum.\n\nWhen using this tool to gather information, it's your responsibility to ensure you have the COMPLETE context. Specifically, each time you call this command you should:\n1) Assess if the contents you viewed are sufficient to proceed with your task.\n2) Take note of where there are lines not shown.\n3) If the file contents you have viewed are insufficient, and you suspect they may be in lines not shown, proactively call the tool again to view those lines.\n4) When in doubt, call this tool again to gather more information. Remember that partial file views may miss critical dependencies, imports, or functionality.\n\nIn some cases, if reading a range of lines is not enough, you may choose to read the entire file.\nReading entire files is often wasteful and slow, especially for large files (i.e. more than a few hundred lines). So you should use this option sparingly.\nReading the entire file is not allowed in most cases. You are only allowed to read the entire file if it has been edited or manually attached to the conversation by the user.",
  "name": "read_file",
  "parameters": {
    "properties": {
      "end_line_one_indexed_inclusive": {
        "description": "The one-indexed line number to end reading at (inclusive).",
        "type": "integer"
      },
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "should_read_entire_file": {
        "description": "Whether to read the entire file. Defaults to false.",
        "type": "boolean"
      },
      "start_line_one_indexed": {
        "description": "The one-indexed line number to start reading from (inclusive).",
        "type": "integer"
      },
      "target_file": {
        "description": "The path of the file to read. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
        "type": "string"
      }
    },
    "required": ["target_file", "should_read_entire_file", "start_line_one_indexed", "end_line_one_indexed_inclusive"],
    "type": "object"
  }
}
```

### 3. run_terminal_cmd
```json
{
  "description": "PROPOSE a command to run on behalf of the user.\nIf you have this tool, note that you DO have the ability to run commands directly on the USER's system.\nNote that the user will have to approve the command before it is executed.\nThe user may reject it if it is not to their liking, or may modify the command before approving it.  If they do change it, take those changes into account.\nThe actual command will NOT execute until the user approves it. The user may not approve it immediately. Do NOT assume the command has started running.\nIf the step is WAITING for user approval, it has NOT started running.\nIn using these tools, adhere to the following guidelines:\n1. Based on the contents of the conversation, you will be told if you are in the same shell as a previous step or a different shell.\n2. If in a new shell, you should `cd` to the appropriate directory and do necessary setup in addition to running the command. By default, the shell will initialize in the project root.\n3. If in the same shell, LOOK IN CHAT HISTORY for your current working directory.\n4. For ANY commands that would require user interaction, ASSUME THE USER IS NOT AVAILABLE TO INTERACT and PASS THE NON-INTERACTIVE FLAGS (e.g. --yes for npx).\n5. If the command would use a pager, append ` | cat` to the command.\n6. For commands that are long running/expected to run indefinitely until interruption, please run them in the background. To run jobs in the background, set `is_background` to true rather than changing the details of the command.\n7. Dont include any newlines in the command.",
  "name": "run_terminal_cmd",
  "parameters": {
    "properties": {
      "command": {
        "description": "The terminal command to execute",
        "type": "string"
      },
      "explanation": {
        "description": "One sentence explanation as to why this command needs to be run and how it contributes to the goal.",
        "type": "string"
      },
      "is_background": {
        "description": "Whether the command should be run in the background",
        "type": "boolean"
      }
    },
    "required": ["command", "is_background"],
    "type": "object"
  }
}
```

### 4. list_dir
```json
{
  "description": "List the contents of a directory.",
  "name": "list_dir",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "relative_workspace_path": {
        "description": "Path to list contents of, relative to the workspace root.",
        "type": "string"
      }
    },
    "required": ["relative_workspace_path"],
    "type": "object"
  }
}
```

### 5. grep_search
```json
{
  "description": "### Instructions:\nThis is best for finding exact text matches or regex patterns.\nThis is preferred over semantic search when we know the exact symbol/function name/etc. to search in some set of directories/file types.\n\nUse this tool to run fast, exact regex searches over text files using the `ripgrep` engine.\nTo avoid overwhelming output, the results are capped at 50 matches.\nUse the include or exclude patterns to filter the search scope by file type or specific paths.\n\n- Always escape special regex characters: ( ) [ ] { } + * ? ^ $ | . \\\n- Use `\\` to escape any of these characters when they appear in your search string.\n- Do NOT perform fuzzy or semantic matches.\n- Return only a valid regex pattern string.\n\n### Examples:\n| Literal               | Regex Pattern            |\n|-----------------------|--------------------------|\n| function(             | function\\(              |\n| value[index]          | value\\[index\\]         |\n| file.txt               | file\\.txt                |\n| user|admin            | user\\|admin             |\n| path\\to\\file         | path\\\\to\\\\file        |\n| hello world           | hello world              |\n| foo\\(bar\\)          | foo\\\\(bar\\\\)         |",
  "name": "grep_search",
  "parameters": {
    "properties": {
      "case_sensitive": {
        "description": "Whether the search should be case sensitive",
        "type": "boolean"
      },
      "exclude_pattern": {
        "description": "Glob pattern for files to exclude",
        "type": "string"
      },
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "include_pattern": {
        "description": "Glob pattern for files to include (e.g. '*.ts' for TypeScript files)",
        "type": "string"
      },
      "query": {
        "description": "The regex pattern to search for",
        "type": "string"
      }
    },
    "required": ["query"],
    "type": "object"
  }
}
```

### 6. edit_file
```json
{
  "description": "Use this tool to propose an edit to an existing file or create a new file.\n\nThis will be read by a less intelligent model, which will quickly apply the edit. You should make it clear what the edit is, while also minimizing the unchanged code you write.\nWhen writing the edit, you should specify each edit in sequence, with the special comment `// ... existing code ...` to represent unchanged code in between edited lines.\n\nFor example:\n\n```\n// ... existing code ...\nFIRST_EDIT\n// ... existing code ...\nSECOND_EDIT\n// ... existing code ...\nTHIRD_EDIT\n// ... existing code ...\n```\n\nYou should still bias towards repeating as few lines of the original file as possible to convey the change.\nBut, each edit should contain sufficient context of unchanged lines around the code you're editing to resolve ambiguity.\nDO NOT omit spans of pre-existing code (or comments) without using the `// ... existing code ...` comment to indicate its absence. If you omit the existing code comment, the model may inadvertently delete these lines.\nMake sure it is clear what the edit should be, and where it should be applied.\nTo create a new file, simply specify the content of the file in the `code_edit` field.\n\nYou should specify the following arguments before the others: [target_file]\n\nALWAYS make all edits to a file in a single edit_file instead of multiple edit_file calls to the same file. The apply model can handle many distinct edits at once. When editing multiple files, ALWAYS make parallel edit_file calls.",
  "name": "edit_file",
  "parameters": {
    "properties": {
      "code_edit": {
        "description": "Specify ONLY the precise lines of code that you wish to edit. **NEVER specify or write out unchanged code**. Instead, represent all unchanged code using the comment of the language you're editing in - example: `// ... existing code ...`",
        "type": "string"
      },
      "instructions": {
        "description": "A single sentence instruction describing what you are going to do for the sketched edit. This is used to assist the less intelligent model in applying the edit. Please use the first person to describe what you are going to do. Dont repeat what you have said previously in normal messages. And use it to disambiguate uncertainty in the edit.",
        "type": "string"
      },
      "target_file": {
        "description": "The target file to modify. Always specify the target file as the first argument. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
        "type": "string"
      }
    },
    "required": ["target_file", "instructions", "code_edit"],
    "type": "object"
  }
}
```

### 7. search_replace
```json
{
  "description": "Use this tool to propose a search and replace operation on an existing file.\n\nThe tool will replace ONE occurrence of old_string with new_string in the specified file.\n\nCRITICAL REQUIREMENTS FOR USING THIS TOOL:\n\n1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:\n   - Include AT LEAST 3-5 lines of context BEFORE the change point\n   - Include AT LEAST 3-5 lines of context AFTER the change point\n   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file\n\n2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:\n   - Make separate calls to this tool for each instance\n   - Each call must uniquely identify its specific instance using extensive context\n\n3. VERIFICATION: Before using this tool:\n   - If multiple instances exist, gather enough context to uniquely identify each one\n   - Plan separate tool calls for each instance\n",
  "name": "search_replace",
  "parameters": {
    "properties": {
      "file_path": {
        "description": "The path to the file you want to search and replace in. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
        "type": "string"
      },
      "new_string": {
        "description": "The edited text to replace the old_string (must be different from the old_string)",
        "type": "string"
      },
      "old_string": {
        "description": "The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)",
        "type": "string"
      }
    },
    "required": ["file_path", "old_string", "new_string"],
    "type": "object"
  }
}
```

### 8. file_search
```json
{
  "description": "Fast file search based on fuzzy matching against file path. Use if you know part of the file path but don't know where it's located exactly. Response will be capped to 10 results. Make your query more specific if need to filter results further.",
  "name": "file_search",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "query": {
        "description": "Fuzzy filename to search for",
        "type": "string"
      }
    },
    "required": ["query", "explanation"],
    "type": "object"
  }
}
```

### 9. delete_file
```json
{
  "description": "Deletes a file at the specified path. The operation will fail gracefully if:\n    - The file doesn't exist\n    - The operation is rejected for security reasons\n    - The file cannot be deleted",
  "name": "delete_file",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "target_file": {
        "description": "The path of the file to delete, relative to the workspace root.",
        "type": "string"
      }
    },
    "required": ["target_file"],
    "type": "object"
  }
}
```

### 10. reapply
```json
{
  "description": "Calls a smarter model to apply the last edit to the specified file.\nUse this tool immediately after the result of an edit_file tool call ONLY IF the diff is not what you expected, indicating the model applying the changes was not smart enough to follow your instructions.",
  "name": "reapply",
  "parameters": {
    "properties": {
      "target_file": {
        "description": "The relative path to the file to reapply the last edit to. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
        "type": "string"
      }
    },
    "required": ["target_file"],
    "type": "object"
  }
}
```

### 11. web_search
```json
{
  "description": "Search the web for real-time information about any topic. Use this tool when you need up-to-date information that might not be available in your training data, or when you need to verify current facts. The search results will include relevant snippets and URLs from web pages. This is particularly useful for questions about current events, technology updates, or any topic that requires recent information.",
  "name": "web_search",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      },
      "search_term": {
        "description": "The search term to look up on the web. Be specific and include relevant keywords for better results. For technical queries, include version numbers or dates if relevant.",
        "type": "string"
      }
    },
    "required": ["search_term"],
    "type": "object"
  }
}
```

### 12. create_diagram
```json
{
  "description": "Creates a Mermaid diagram that will be rendered in the chat UI. Provide the raw Mermaid DSL string via `content`.\nUse <br/> for line breaks, always wrap diagram texts/tags in double quotes, do not use custom colors, do not use :::, and do not use beta features.\n\n⚠️  Security note: Do **NOT** embed remote images (e.g., using <image>, <img>, or markdown image syntax) inside the diagram, as they will be stripped out. If you need an image it must be a trusted local asset (e.g., data URI or file on disk).\nThe diagram will be pre-rendered to validate syntax – if there are any Mermaid syntax errors, they will be returned in the response so you can fix them.",
  "name": "create_diagram",
  "parameters": {
    "properties": {
      "content": {
        "description": "Raw Mermaid diagram definition (e.g., 'graph TD; A-->B;').",
        "type": "string"
      }
    },
    "required": ["content"],
    "type": "object"
  }
}
```

### 13. edit_notebook
```json
{
  "description": "Use this tool to edit a jupyter notebook cell. Use ONLY this tool to edit notebooks.\n\nThis tool supports editing existing cells and creating new cells:\n\t- If you need to edit an existing cell, set 'is_new_cell' to false and provide the 'old_string' and 'new_string'.\n\t\t-- The tool will replace ONE occurrence of 'old_string' with 'new_string' in the specified cell.\n\t- If you need to create a new cell, set 'is_new_cell' to true and provide the 'new_string' (and keep 'old_string' empty).\n\t- It's critical that you set the 'is_new_cell' flag correctly!\n\t- This tool does NOT support cell deletion, but you can delete the content of a cell by passing an empty string as the 'new_string'.\n\nOther requirements:\n\t- Cell indices are 0-based.\n\t- 'old_string' and 'new_string' should be a valid cell content, i.e. WITHOUT any JSON syntax that notebook files use under the hood.\n\t- The old_string MUST uniquely identify the specific instance you want to change. This means:\n\t\t-- Include AT LEAST 3-5 lines of context BEFORE the change point\n\t\t-- Include AT LEAST 3-5 lines of context AFTER the change point\n\t- This tool can only change ONE instance at a time. If you need to change multiple instances:\n\t\t-- Make separate calls to this tool for each instance\n\t\t-- Each call must uniquely identify its specific instance using extensive context\n\t- This tool might save markdown cells as \"raw\" cells. Don't try to change it, it's fine. We need it to properly display the diff.\n\t- If you need to create a new notebook, just set 'is_new_cell' to true and cell_idx to 0.\n\t- ALWAYS generate arguments in the following order: target_notebook, cell_idx, is_new_cell, cell_language, old_string, new_string.\n\t- Prefer editing existing cells over creating new ones!\n",
  "name": "edit_notebook",
  "parameters": {
    "properties": {
      "cell_idx": {
        "description": "The index of the cell to edit (0-based)",
        "type": "number"
      },
      "cell_language": {
        "description": "The language of the cell to edit. Should be STRICTLY one of these: 'python', 'markdown', 'javascript', 'typescript', 'r', 'sql', 'shell', 'raw' or 'other'.",
        "type": "string"
      },
      "is_new_cell": {
        "description": "If true, a new cell will be created at the specified cell index. If false, the cell at the specified cell index will be edited.",
        "type": "boolean"
      },
      "new_string": {
        "description": "The edited text to replace the old_string or the content for the new cell.",
        "type": "string"
      },
      "old_string": {
        "description": "The text to replace (must be unique within the cell, and must match the cell contents exactly, including all whitespace and indentation).",
        "type": "string"
      },
      "target_notebook": {
        "description": "The path to the notebook file you want to edit. You can use either a relative path in the workspace or an absolute path. If an absolute path is provided, it will be preserved as is.",
        "type": "string"
      }
    },
    "required": ["target_notebook", "cell_idx", "is_new_cell", "cell_language", "old_string", "new_string"],
    "type": "object"
  }
}
```

### 14. todo_write
```json
{
  "description": "Use this tool to create and manage a structured task list for your current coding session. This helps track progress, organize complex tasks, and demonstrate thoroughness.\n\n### When to Use This Tool\n\nUse proactively for:\n1. Complex multi-step tasks (3+ distinct steps)\n2. Non-trivial tasks requiring careful planning\n3. User explicitly requests todo list\n4. User provides multiple tasks (numbered/comma-separated)\n5. After receiving new instructions - capture requirements as todos (use merge=false to add new ones)\n6. After completing tasks - mark complete with merge=true and add follow-ups\n7. When starting new tasks - mark as in_progress (ideally only one at a time)\n\n### When NOT to Use\n\nSkip for:\n1. Single, straightforward tasks\n2. Trivial tasks with no organizational benefit\n3. Tasks completable in < 3 trivial steps\n4. Purely conversational/informational requests\n5. Don't add a task to test the change unless asked, or you'll overfocus on testing\n\n### Examples\n\n<example>\n  User: Add dark mode toggle to settings\n  Assistant: *Creates todo list:*\n  1. Add state management - no dependencies\n  2. Implement styles - depends on task 1\n  3. Create toggle component - depends on tasks 1, 2\n  4. Update components - depends on tasks 1, 2\n<reasoning>\n  Multi-step feature with dependencies; user requested tests/build afterward.\n</reasoning>\n</example>\n\n<example>\n  User: Rename getCwd to getCurrentWorkingDirectory across my project\n  Assistant: *Searches codebase, finds 15 instances across 8 files*\n  *Creates todo list with specific items for each file that needs updating*\n\n<reasoning>\n  Complex refactoring requiring systematic tracking across multiple files.\n</reasoning>\n</example>\n\n<example>\n  User: Implement user registration, product catalog, shopping cart, checkout flow.\n  Assistant: *Creates todo list breaking down each feature into specific tasks*\n\n<reasoning>\n  Multiple complex features provided as list requiring organized task management.\n</reasoning>\n</example>\n\n<example>\n  User: Optimize my React app - it's rendering slowly.\n  Assistant: *Analyzes codebase, identifies issues*\n  *Creates todo list: 1) Memoization, 2) Virtualization, 3) Image optimization, 4) Fix state loops, 5) Code splitting*\n\n<reasoning>\n  Performance optimization requires multiple steps across different components.\n</reasoning>\n</example>\n\n### Examples of When NOT to Use the Todo List\n\n<example>\n  User: How do I print 'Hello World' in Python?\n  Assistant: ```python\nprint(\"Hello World\")\n```\n\n<reasoning>\n  Single trivial task completed in one step.\n</reasoning>\n</example>\n\n<example>\n  User: What does git status do?\n  Assistant: Shows current state of working directory and staging area...\n\n<reasoning>\n  Informational request with no coding task to complete.\n</reasoning>\n</example>\n\n<example>\n  User: Add comment to calculateTotal function.\n  Assistant: *Uses edit tool to add comment*\n\n<reasoning>\n  Single straightforward task in one location.\n</reasoning>\n</example>\n\n<example>\n  User: Run npm install for me.\n  Assistant: *Executes npm install* Command completed successfully...\n\n<reasoning>\n  Single command execution with immediate results.\n</reasoning>\n</example>\n\n### Task States and Management\n\n1. **Task States:**\n  - pending: Not yet started\n  - in_progress: Currently working on\n  - completed: Finished successfully\n  - cancelled: No longer needed\n\n2. **Task Management:**\n  - Update status in real-time\n  - Mark complete IMMEDIATELY after finishing\n  - Only ONE task in_progress at a time\n  - Complete current tasks before starting new ones\n\n3. **Task Breakdown:**\n  - Create specific, actionable items\n  - Break complex tasks into manageable steps\n  - Use clear, descriptive names\n\n4. **Task Dependencies:**\n  - Use dependencies field for natural prerequisites\n  - Avoid circular dependencies\n  - Independent tasks can run in parallel\n\nWhen in doubt, use this tool. Proactive task management demonstrates attentiveness and ensures complete requirements.",
  "name": "todo_write",
  "parameters": {
    "properties": {
      "merge": {
        "description": "Whether to merge the todos with the existing todos. If true, the todos will be merged into the existing todos based on the id field. You can leave unchanged properties undefined. If false, the new todos will replace the existing todos.",
        "type": "boolean"
      },
      "todos": {
        "description": "Array of TODO items to write to the workspace",
        "items": {
          "properties": {
            "content": {
              "description": "The description/content of the TODO item",
              "type": "string"
            },
            "dependencies": {
              "description": "List of other task IDs that are prerequisites for this task, i.e. we cannot complete this task until these tasks are done",
              "items": {
                "type": "string"
              },
              "type": "array"
            },
            "id": {
              "description": "Unique identifier for the TODO item",
              "type": "string"
            },
            "status": {
              "description": "The current status of the TODO item",
              "enum": ["pending", "in_progress", "completed", "cancelled"],
              "type": "string"
            }
          },
          "required": ["content", "status", "id", "dependencies"],
          "type": "object"
        },
        "minItems": 2,
        "type": "array"
      }
    },
    "required": ["merge", "todos"],
    "type": "object"
  }
}
```

### 15. mcp_mcp-batch-read-file_batch_read_file (bad version 1)
```json
{
  "description": "Read the contents of multiple files in a single batch operation. This tool improves efficiency by reading multiple files at once instead of making separate read_file calls.\n\nParameter usage:\n- Set should_read_entire_file=true to read the entire file (line range parameters are ignored)\n- Set should_read_entire_file=false and provide start_line_one_indexed and end_line_one_indexed_inclusive for specific ranges\n- Do NOT provide both should_read_entire_file=true AND line range parameters as this creates ambiguity\n\nKey features:\n- Batch processing of multiple files\n- Individual line range control per file\n- Global and per-file line limits\n- Error handling with continue-on-error option\n- Optional outline generation\n- Structured response with success/error counts\n\nThis tool is particularly useful when you need to read multiple related files (e.g., examining imports, comparing implementations, or gathering context from multiple source files).",
  "name": "mcp_mcp-batch-read-file_batch_read_file",
  "parameters": {
    "properties": {
      "explanation": {
        "description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
        "type": "string"
      }
    },
    "required": ["explanation"],
    "type": "object"
  }
}
```

## Tool Usage Statistics
- **Total tools available**: 15
- **Tool calls used in this session**: 34 (including this one)

## Tool Categories

### File Operations (6 tools)
- `read_file` - Read file contents with line range support
- `mcp_mcp-batch-read-file_batch_read_file` - Read multiple files efficiently in a single operation
- `edit_file` - Edit existing files or create new files
- `search_replace` - Search and replace operations
- `delete_file` - Delete files
- `reapply` - Re-apply edits with a smarter model

### Search & Discovery (4 tools)
- `codebase_search` - Semantic search across codebase
- `grep_search` - Regex/text pattern search
- `file_search` - Fuzzy file path search
- `list_dir` - Directory listing

### Execution & System (1 tool)
- `run_terminal_cmd` - Execute terminal commands

### Content Creation (2 tools)
- `create_diagram` - Create Mermaid diagrams
- `edit_notebook` - Edit Jupyter notebooks

### Task Management (1 tool)
- `todo_write` - Create and manage structured task lists

### External Information (1 tool)
- `web_search` - Web search for external information

## Key Tool Usage Guidelines

1. **Batch Operations**: Use `mcp_mcp-batch-read-file_batch_read_file` instead of multiple `read_file` calls when reading 2+ files
2. **Search Strategy**: Use `codebase_search` for semantic/conceptual searches, `grep_search` for exact text/regex patterns
3. **File Discovery**: Use `file_search` for finding files by name, `list_dir` for exploring directory structure
4. **Editing**: Prefer `edit_file` for new files or major changes, `search_replace` for targeted replacements
5. **Task Management**: Use `todo_write` for complex multi-step tasks or when user requests task tracking

All tools require proper parameter validation and follow specific usage patterns as defined in their schemas.
