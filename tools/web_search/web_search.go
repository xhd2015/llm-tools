package web_search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xhd2015/llm-tools/jsonschema"
	"github.com/xhd2015/llm-tools/tools/defs"
)

// WebSearchRequest represents the input parameters for the web_search tool
type WebSearchRequest struct {
	SearchTerm  string `json:"search_term"`
	Explanation string `json:"explanation"`
}

// WebSearchResult represents a single search result
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearchResponse represents the output of the web_search tool
type WebSearchResponse struct {
	SearchTerm string            `json:"search_term"`
	Results    []WebSearchResult `json:"results"`
	Message    string            `json:"message"`
}

// GetToolDefinition returns the JSON schema definition for the web_search tool
func GetToolDefinition() defs.ToolDefinition {
	return defs.ToolDefinition{
		Description: "Search the web for real-time information about any topic. Use this tool when you need up-to-date information that might not be available in your training data, or when you need to verify current facts. The search results will include relevant snippets and URLs from web pages. This is particularly useful for questions about current events, technology updates, or any topic that requires recent information.",
		Name:        "web_search",
		Parameters: &jsonschema.JsonSchema{
			Type: jsonschema.ParamTypeObject,
			Properties: map[string]*jsonschema.JsonSchema{
				"search_term": {
					Type:        jsonschema.ParamTypeString,
					Description: "The search term to look up on the web. Be specific and include relevant keywords for better results. For technical queries, include version numbers or dates if relevant.",
				},
				"explanation": {
					Type:        jsonschema.ParamTypeString,
					Description: "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
				},
			},
			Required: []string{"search_term"},
		},
	}
}

// WebSearch executes the web_search tool with the given parameters
func WebSearch(req WebSearchRequest) (*WebSearchResponse, error) {
	// This is a basic implementation that simulates web search
	// In a real implementation, this would integrate with a search API like Google Custom Search, Bing, or DuckDuckGo

	// For now, we'll return a simulated response indicating that web search is not fully implemented
	response := &WebSearchResponse{
		SearchTerm: req.SearchTerm,
		Results:    []WebSearchResult{},
		Message:    "Web search functionality is not fully implemented in this version. This tool would normally search the web for real-time information.",
	}

	// Try to make a basic web request to demonstrate the concept
	// In a real implementation, this would use a proper search API
	searchURL := fmt.Sprintf("https://duckduckgo.com/html/?q=%s", url.QueryEscape(req.SearchTerm))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Get(searchURL)
	if err != nil {
		response.Message = fmt.Sprintf("Web search simulation failed: %v. In a real implementation, this would search the web for: %s", err, req.SearchTerm)
		return response, nil
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to read search response: %v. In a real implementation, this would search the web for: %s", err, req.SearchTerm)
		return response, nil
	}

	// Basic parsing simulation (in a real implementation, this would parse actual search results)
	if len(body) > 0 {
		response.Results = []WebSearchResult{
			{
				Title:   fmt.Sprintf("Search results for: %s", req.SearchTerm),
				URL:     searchURL,
				Snippet: "This is a simulated web search result. In a real implementation, this would contain actual search results from the web with relevant snippets and URLs.",
			},
		}
		response.Message = "Web search simulation completed. In a real implementation, this would return actual search results from the web."
	}

	return response, nil
}

// validateSearchTerm validates the search term
func validateSearchTerm(searchTerm string) error {
	if strings.TrimSpace(searchTerm) == "" {
		return fmt.Errorf("search term cannot be empty")
	}

	if len(searchTerm) > 1000 {
		return fmt.Errorf("search term is too long (max 1000 characters)")
	}

	return nil
}

func ParseJSONRequest(jsonInput string) (WebSearchRequest, error) {
	var req WebSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return WebSearchRequest{}, fmt.Errorf("failed to parse JSON input: %w", err)
	}
	return req, nil
}

// ExecuteFromJSON executes the web_search tool from JSON input
func ExecuteFromJSON(jsonInput string) (string, error) {
	var req WebSearchRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Validate search term
	if err := validateSearchTerm(req.SearchTerm); err != nil {
		return "", err
	}

	response, err := WebSearch(req)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonOutput), nil
}
