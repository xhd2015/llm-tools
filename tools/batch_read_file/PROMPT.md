# Use batch_read_file Instead of Multiple read_file Calls

## When to Use batch_read_file

**ALWAYS use `batch_read_file` instead of multiple `read_file` calls when:**

1. **Reading 2 or more files** - Even for just 2 files, batch_read_file is more efficient
2. **Examining related files** - Such as imports, dependencies, or related components
3. **Comparing implementations** - When you need to see multiple files side-by-side
4. **Gathering context** - When understanding how multiple files work together
5. **Code exploration** - When diving into a codebase and need to read several files

## Benefits of batch_read_file

- **Efficiency**: Single tool call instead of multiple sequential calls
- **Consistency**: All files processed with same settings and limits
- **Error handling**: Continue processing other files if one fails
- **Structured output**: Organized response with success/error counts
- **Parallel processing**: Files can be processed simultaneously

## Parameter Usage Guidelines

### For Entire Files
```json
{
  "files": [
    {
      "target_file": "path/to/file1.go",
      "should_read_entire_file": true
    },
    {
      "target_file": "path/to/file2.go", 
      "should_read_entire_file": true
    }
  ]
}
```

### For Specific Line Ranges
```json
{
  "files": [
    {
      "target_file": "path/to/file1.go",
      "should_read_entire_file": false,
      "start_line_one_indexed": 1,
      "end_line_one_indexed_inclusive": 50
    },
    {
      "target_file": "path/to/file2.go",
      "should_read_entire_file": false, 
      "start_line_one_indexed": 100,
      "end_line_one_indexed_inclusive": 200
    }
  ]
}
```

### Mixed Usage (Some Entire, Some Ranges)
```json
{
  "files": [
    {
      "target_file": "main.go",
      "should_read_entire_file": true
    },
    {
      "target_file": "large_file.go",
      "should_read_entire_file": false,
      "start_line_one_indexed": 500,
      "end_line_one_indexed_inclusive": 600
    }
  ]
}
```

## CRITICAL: Avoid Parameter Conflicts

**DO NOT** provide both `should_read_entire_file=true` AND line range parameters:

❌ **WRONG - Creates ambiguity:**
```json
{
  "target_file": "file.go",
  "should_read_entire_file": true,
  "start_line_one_indexed": 1,
  "end_line_one_indexed_inclusive": 100
}
```

✅ **CORRECT - Choose one approach:**
```json
{
  "target_file": "file.go",
  "should_read_entire_file": true
}
```

## Examples of When to Use batch_read_file

### Example 1: Understanding a Component
Instead of:
```
read_file component.go
read_file component_test.go  
read_file component_types.go
```

Use:
```json
{
  "files": [
    {"target_file": "component.go", "should_read_entire_file": true},
    {"target_file": "component_test.go", "should_read_entire_file": true},
    {"target_file": "component_types.go", "should_read_entire_file": true}
  ],
  "explanation": "Reading component files to understand the implementation"
}
```

### Example 2: Examining Imports
Instead of:
```
read_file main.go lines 1-20
read_file utils.go lines 1-30
read_file config.go lines 1-25
```

Use:
```json
{
  "files": [
    {"target_file": "main.go", "should_read_entire_file": false, "start_line_one_indexed": 1, "end_line_one_indexed_inclusive": 20},
    {"target_file": "utils.go", "should_read_entire_file": false, "start_line_one_indexed": 1, "end_line_one_indexed_inclusive": 30},
    {"target_file": "config.go", "should_read_entire_file": false, "start_line_one_indexed": 1, "end_line_one_indexed_inclusive": 25}
  ],
  "explanation": "Examining import statements across multiple files"
}
```

## Default Settings

- `continue_on_error`: true (keep processing other files if one fails)
- `include_outline`: true (provide code outline for context)
- `global_max_lines`: 250 (same as read_file)
- `global_min_lines`: 200 (same as read_file minimum expansion)

## Remember

- **Think batch first**: Before making multiple read_file calls, consider if batch_read_file is more appropriate
- **One tool call**: Replace 2+ read_file calls with a single batch_read_file call
- **Clear parameters**: Use either entire file OR line ranges, never both
- **Meaningful explanations**: Always provide context for why you're reading these files together
