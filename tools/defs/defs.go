package defs

import "github.com/xhd2015/llm-tools/jsonschema"

// ToolDefinition represents the JSON schema for the grep_search tool
type ToolDefinition struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Parameters  *jsonschema.JsonSchema `json:"parameters"`
}
