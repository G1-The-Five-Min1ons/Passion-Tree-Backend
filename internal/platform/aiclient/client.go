package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"passiontree/internal/learning-path/model"
	batchmodel "passiontree/internal/recommendation/model"

	"github.com/go-resty/resty/v2"
)

const defaultAIClientTimeout = 60 * time.Second

// AIClient represents a client for AI service
type AIClient struct {
	baseURL string
	client  *resty.Client
}

// NewAIClient creates a new AI service client
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		client:  resty.New().SetTimeout(defaultAIClientTimeout),
	}
}

// DebugClientPointer logs the pointer of the internal HTTP client for debugging
func (c *AIClient) DebugClientPointer() {
	fmt.Printf("[DEBUG] AIClient.client pointer: %p\n", c.client)
}

// Search performs a semantic search via AI service
func (c *AIClient) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.TopK < 0 {
		return nil, fmt.Errorf("top_k must be >= 1, got %d", req.TopK)
	}
	if req.TopK == 0 {
		req.TopK = 7
	}
	if req.ResourceType == "" {
		req.ResourceType = "learning_paths"
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/api/v1/search/")
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("search cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var searchResp SearchResponse
	if err := json.Unmarshal(resp.Body(), &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &searchResp, nil
}

// Ping checks if AI service is reachable
func (c *AIClient) Ping(ctx context.Context) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Get(c.baseURL)
	if err != nil {
		return fmt.Errorf("failed to ping AI service: %w", err)
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 404 {
		return fmt.Errorf("AI service returned unexpected status: %d", resp.StatusCode())
	}
	return nil
}

// GenerateLearningPath calls AI to generate a roadmap
func (c *AIClient) GenerateLearningPath(ctx context.Context, topic string) (*model.AIPathGenerationResponse, error) {
	reqBody := model.AIGeneratePathRequest{Topic: topic}
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(c.baseURL + "/api/v1/generator/learning-path")
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("generation cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("connection error: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status: %d body: %s", resp.StatusCode(), resp.String())
	}
	var aiResponse model.AIPathGenerationResponse
	if err := json.Unmarshal(resp.Body(), &aiResponse); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return &aiResponse, nil
}

// AnalyzeSentiment calls AI service to analyze learning reflection
func (c *AIClient) AnalyzeSentiment(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/api/v1/sentiment/analyze")
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("sentiment analysis cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("connection error: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status: %d body: %s", resp.StatusCode(), resp.String())
	}
	var sentimentResp SentimentResponse
	if err := json.Unmarshal(resp.Body(), &sentimentResp); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return &sentimentResp, nil
}

// GetCollectionInfo retrieves debug information about a collection from AI service
func (c *AIClient) GetCollectionInfo(collectionName string) (*CollectionInfoResponse, error) {
	url := fmt.Sprintf("%s/api/v1/search/debug/collection/%s", c.baseURL, collectionName)
	resp, err := c.client.R().
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var info CollectionInfoResponse
	if err := json.Unmarshal(resp.Body(), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &info, nil
}

func (c *AIClient) SyncLearningPath(ctx context.Context, req SyncLearningPathRequest) (*SyncLearningPathResponse, error) {
	if req.CollectionName == "" {
		req.CollectionName = "learning_paths"
	}
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/api/v1/search/sync")
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var syncResp SyncLearningPathResponse
	if err := json.Unmarshal(resp.Body(), &syncResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &syncResp, nil
}

// SyncDeletePath removes a learning path point from Qdrant.
func (c *AIClient) SyncDeletePath(ctx context.Context, pathID string, collectionName string) (*SyncLearningPathResponse, error) {
	if collectionName == "" {
		collectionName = "learning_paths"
	}

	url := fmt.Sprintf("%s/api/v1/search/sync/%s", c.baseURL, pathID)

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("collection_name", collectionName).
		Delete(url)

	if err != nil {
		return nil, fmt.Errorf("failed to send delete request: %w", err) // ใช้ %w
	}

	if resp.IsError() { // เช็ค Status code แบบครอบคลุม
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var syncResp SyncLearningPathResponse
	if err := json.Unmarshal(resp.Body(), &syncResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delete response: %w", err)
	}	
	return &syncResp, nil
}

// BulkSyncLearningPaths upserts many paths in one call.
func (c *AIClient) BulkSyncLearningPaths(ctx context.Context, req BulkSyncRequest) (*BulkSyncResponse, error) {
	if req.CollectionName == "" {
		req.CollectionName = "learning_paths"
	}
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/api/v1/search/sync/bulk")
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk sync request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var bulkResp BulkSyncResponse
	if err := json.Unmarshal(resp.Body(), &bulkResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bulk sync response: %w", err)
	}
	return &bulkResp, nil
}

// ListCollectionIDs returns every point id inside a Qdrant collection.
func (c *AIClient) ListCollectionIDs(ctx context.Context, collectionName string) (*ListIDsResponse, error) {
	if collectionName == "" {
		collectionName = "learning_paths"
	}
	resp, err := c.client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/api/v1/search/ids/%s", c.baseURL, collectionName))
	if err != nil {
		return nil, fmt.Errorf("failed to send list ids request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var ids ListIDsResponse
	if err := json.Unmarshal(resp.Body(), &ids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list ids response: %w", err)
	}
	return &ids, nil
}

func (c *AIClient) ComputeBatchRecommendations(ctx context.Context, payload batchmodel.BatchRecommendPayload) (*batchmodel.BatchRecommendResponse, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(c.baseURL + "/api/v1/recommend/batch-compute")

	if err != nil {
		return nil, fmt.Errorf("failed to send batch request: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var result batchmodel.BatchRecommendResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}
	return &result, nil
}
