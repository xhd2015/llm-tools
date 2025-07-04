package jsonschema

type ParamType string

const (
	ParamTypeObject ParamType = "object"
	ParamTypeString ParamType = "string"
	ParamTypeArray  ParamType = "array"
	ParamTypeNumber ParamType = "number"
)

type JsonSchema struct {
	Type        ParamType              `json:"type"`
	Properties  map[string]*JsonSchema `json:"properties,omitempty"`
	Description string                 `json:"description,omitempty"`
	Items       *JsonSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
}
