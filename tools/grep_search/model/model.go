package model

// GrepSearchRequest represents the input parameters for the grep_search tool
type GrepSearchRequest struct {
	WorkspaceRoot        string `json:"workspace_root"`
	RelativePathToSearch string `json:"relative_path_to_search"`
	Query                string `json:"query"`
	CaseSensitive        bool   `json:"case_sensitive,omitempty"`
	ExcludePattern       string `json:"exclude_pattern,omitempty"`
	IncludePattern       string `json:"include_pattern,omitempty"`
	Explanation          string `json:"explanation"`
}

// GrepSearchMatch represents a single search match result
type GrepSearchMatch struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column,omitempty"`
	Content    string `json:"content"`
	MatchStart int    `json:"match_start,omitempty"`
	MatchEnd   int    `json:"match_end,omitempty"`
}

// GrepSearchResponse represents the output of the grep_search tool
type GrepSearchResponse struct {
	Matches      []GrepSearchMatch `json:"matches"`
	TotalMatches int               `json:"total_matches"`
	SearchQuery  string            `json:"search_query"`
	Truncated    bool              `json:"truncated"`
}

// GrepSearcher defines the interface for different grep search implementations
type GrepSearcher interface {
	Search(req GrepSearchRequest) (*GrepSearchResponse, error)
	IsAvailable() bool
}
