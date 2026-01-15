package aiclient

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/model"
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
		client:  fiber.AcquireClient(),
	}
}

// Search performs a semantic search via AI service
func (c *AIClient) Search(req SearchRequest) (*SearchResponse, error) {
	// Default values
	if req.TopK == 0 {
		req.TopK = 7
	}
	if req.ResourceType == "" {
		req.ResourceType = "learning_paths"
	}

	// Create request
	agent := c.client.Post(c.baseURL + "/api/v1/search/")
	agent.JSON(req)

	// Send request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to send request: %v", errs[0])
	}

	// Check status code
	if statusCode != fiber.StatusOK {
		return nil, fmt.Errorf("AI service returned status %d: %s", statusCode, string(body))
	}

	// Unmarshal response
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResp, nil
}

// Ping checks if AI service is reachable
func (c *AIClient) Ping() error {
	agent := c.client.Get(c.baseURL)
	statusCode, _, errs := agent.Bytes()

	if len(errs) > 0 {
		return fmt.Errorf("failed to ping AI service: %v", errs[0])
	}

	if statusCode != fiber.StatusOK && statusCode != fiber.StatusNotFound {
		return fmt.Errorf("AI service returned unexpected status: %d", statusCode)
	}

	return nil
}

func (c *AIClient) GenerateLearningPath(topic string) (*model.AIPathGenerationResponse, error) {
	reqBody := model.AIGeneratePathRequest{Topic: topic}
	agent := c.client.Post(c.baseURL + "/api/v1/generator/learning-path")
	agent.JSON(reqBody)
	statusCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
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
