package jsonschema

type ParamType string

const (
	ParamTypeObject  ParamType = "object"
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeArray   ParamType = "array"
	ParamTypeNumber  ParamType = "number"
)

type JsonSchema struct {
	Type        ParamType              `json:"type"`
	Properties  map[string]*JsonSchema `json:"properties,omitempty"`
	Description string                 `json:"description,omitempty"`
	Items       *JsonSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
}

func (c *JsonSchema) ToMap() map[string]any {
	if c == nil {
		return nil
	}
	m := make(map[string]any, 2)
	m["type"] = c.Type
	if c.Description != "" {
		m["description"] = c.Description
	}
	if c.Properties != nil {
		m["properties"] = c.Properties
	}
	if c.Items != nil {
		m["items"] = c.Items.ToMap()
	}
	if len(c.Required) > 0 {
		m["required"] = c.Required
	}
	return m
}

func (c *JsonSchema) PropertiesToMap() map[string]any {
	if c == nil || c.Properties == nil {
		return nil
	}
	m := make(map[string]any, len(c.Properties))
	for k, v := range c.Properties {
		m[k] = v.ToMap()
	}
	return m
}
