package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"passiontree/internal/learning-path/model"

	"github.com/gofiber/fiber/v2"
)

// AIClient represents a client for AI service
type AIClient struct {
	baseURL string
	client  *fiber.Client
}

// NewAIClient creates a new AI service client
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		client:  fiber.AcquireClient(), // ใช้ Client Pool เพื่อประสิทธิภาพสูงสุด
	}
}

// Search performs a semantic search via AI service
func (c *AIClient) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.TopK == 0 {
		req.TopK = 7
	}
	if req.ResourceType == "" {
		req.ResourceType = "learning_paths"
	}

	agent := c.client.Post(c.baseURL + "/api/v1/search/")

	if deadline, ok := ctx.Deadline(); ok {
		agent.Timeout(time.Until(deadline))
	}

	agent.JSON(req)

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("search cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to send request: %v", errs[0])
	}

	if statusCode != fiber.StatusOK {
		return nil, fmt.Errorf("AI service returned status %d: %s", statusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResp, nil
}

// Ping checks if AI service is reachable
func (c *AIClient) Ping(ctx context.Context) error {
	agent := c.client.Get(c.baseURL)

	if deadline, ok := ctx.Deadline(); ok {
		agent.Timeout(time.Until(deadline))
	}

	statusCode, _, errs := agent.Bytes()
	if len(errs) > 0 {
		return fmt.Errorf("failed to ping AI service: %v", errs[0])
	}

	if statusCode != fiber.StatusOK && statusCode != fiber.StatusNotFound {
		return fmt.Errorf("AI service returned unexpected status: %d", statusCode)
	}

	return nil
}

// GenerateLearningPath calls AI to generate a roadmap
func (c *AIClient) GenerateLearningPath(ctx context.Context, topic string) (*model.AIPathGenerationResponse, error) {
	reqBody := model.AIGeneratePathRequest{Topic: topic}
	agent := c.client.Post(c.baseURL + "/api/v1/generator/learning-path")

	if deadline, ok := ctx.Deadline(); ok {
		agent.Timeout(time.Until(deadline))
	}

	agent.JSON(reqBody)
	statusCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("generation cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("connection error: %v", errs[0])
	}

	if statusCode != fiber.StatusOK {
		return nil, fmt.Errorf("AI service returned status: %d body: %s", statusCode, string(body))
	}

	var aiResponse model.AIPathGenerationResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &aiResponse, nil
}

// AnalyzeSentiment calls AI service to analyze learning reflection
func (c *AIClient) AnalyzeSentiment(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	agent := c.client.Post(c.baseURL + "/api/v1/sentiment/analyze")

	if deadline, ok := ctx.Deadline(); ok {
		agent.Timeout(time.Until(deadline))
	}

	agent.JSON(req)
	statusCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("sentiment analysis cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("connection error: %v", errs[0])
	}

	if statusCode != fiber.StatusOK {
		return nil, fmt.Errorf("AI service returned status: %d body: %s", statusCode, string(body))
	}

	var sentimentResp SentimentResponse
	if err := json.Unmarshal(body, &sentimentResp); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &sentimentResp, nil
}
