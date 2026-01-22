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

// SamplePoint represents a sample point in the collection
type SamplePoint struct {
	ID      interface{}            `json:"id"`
	Payload map[string]interface{} `json:"payload"`
}

// CollectionInfoResponse represents collection debug information from AI service
type CollectionInfoResponse struct {
	CollectionName string        `json:"collection_name"`
	PointsCount    interface{}   `json:"points_count"`
	VectorsConfig  string        `json:"vectors_config"`
	SamplePoints   []SamplePoint `json:"sample_points"`
	TotalScrolled  int           `json:"total_scrolled"`
}

// SyncLearningPathRequest represents the request to sync a learning path to Qdrant
type SyncLearningPathRequest struct {
	PathID         int                    `json:"path_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CollectionName string                 `json:"collection_name,omitempty"`
}

// SyncLearningPathResponse represents the response from AI service for sync operations
type SyncLearningPathResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PathID  int    `json:"path_id,omitempty"`
}
