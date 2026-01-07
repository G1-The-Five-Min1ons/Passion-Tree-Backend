package aiclient

// SearchRequest represents the search request payload for AI service
type SearchRequest struct {
	Query        string                 `json:"query"`
	TopK         int                    `json:"top_k"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	ResourceType string                 `json:"resource_type,omitempty"`
}

// SearchResult represents a single search result from AI service
type SearchResult struct {
	ID      interface{}            `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchResponse represents the search response from AI service
type SearchResponse struct {
	Query   string         `json:"query"`
	Total   int            `json:"total"`
	Results []SearchResult `json:"results"`
}
