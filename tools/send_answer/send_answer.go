package send_answer

import (
	"encoding/json"
	"fmt"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// serve as an echo

// SendAnswerRequest represents the input parameters for the send_answer tool
type SendAnswerRequest struct {
	Answer      []string `json:"answer"`
	Explanation string   `json:"explanation"`
}

// SendAnswerResponse represents the output of the send_answer tool
type SendAnswerResponse struct {
	Message string `json:"message"`
}

// AnswerData represents the structure of the data written to the answer file
type AnswerData struct {
	Answer []string `json:"answer"`
}

// GetToolDefinition returns the JSON schema definition for the send_answer tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: `Send a structured answer to another tool. It is used to provide structured responses that can be consumed by other tools or processes.`,
		Name:        "send_answer",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"answer": {
					Type:        jsonschema.ParamTypeArray,
					Description: "The structured answer as an array of strings.",
					Items: &jsonschema.JsonSchema{
						Type: jsonschema.ParamTypeString,
					},
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"answer"},
		},
	}
}

// SendAnswer executes the send_answer tool with the given parameters
func SendAnswer(req SendAnswerRequest) (*SendAnswerResponse, error) {
	return &SendAnswerResponse{
		Message: "Answer sent successfully",
	}, nil
}

func ParseJSONRequest(jsonInput string) (SendAnswerRequest, error) {
	var req SendAnswerRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return SendAnswerRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the send_answer tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req SendAnswerRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	response, err := SendAnswer(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
