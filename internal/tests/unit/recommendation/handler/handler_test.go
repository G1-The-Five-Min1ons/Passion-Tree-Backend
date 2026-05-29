package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"passiontree/internal/pkg/apperror"
	rechandler "passiontree/internal/recommendation/handler"
	"passiontree/internal/recommendation/model"

	"github.com/gofiber/fiber/v2"
)

// mockService satisfies service.Service for handler unit tests.
type mockService struct {
	recommendPathsFunc     func(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error)
	recommendHomePathsFunc func(ctx context.Context, userID string) (*model.RecommendPathResponse, error)
	runBatchFunc           func(ctx context.Context) error
	recomputeForUserFunc   func(ctx context.Context, userID string) error
}

func (m *mockService) RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
	if m.recommendPathsFunc != nil {
		return m.recommendPathsFunc(ctx, userID, treeID)
	}
	return &model.RecommendPathResponse{}, nil
}

func (m *mockService) RecommendHomePathsForUser(ctx context.Context, userID string) (*model.RecommendPathResponse, error) {
	if m.recommendHomePathsFunc != nil {
		return m.recommendHomePathsFunc(ctx, userID)
	}
	return &model.RecommendPathResponse{}, nil
}

func (m *mockService) RunDailyRecommendationBatch(ctx context.Context) error {
	if m.runBatchFunc != nil {
		return m.runBatchFunc(ctx)
	}
	return nil
}

func (m *mockService) RecomputeForUser(ctx context.Context, userID string) error {
	if m.recomputeForUserFunc != nil {
		return m.recomputeForUserFunc(ctx, userID)
	}
	return nil
}

func newTestApp(svc *mockService) (*rechandler.Handler, *fiber.App) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := rechandler.NewHandler(svc, logger, nil)
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	return h, app
}

func parseBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return body
}

// ── GetRecommendations ────────────────────────────────────────────────────────

func TestGetRecommendations_NoAuth_Returns401(t *testing.T) {
	svc := &mockService{}
	h, app := newTestApp(svc)

	// No user_id local set → handler returns 401
	app.Get("/reflect/recommendation", h.GetRecommendations)

	req := httptest.NewRequest(http.MethodGet, "/reflect/recommendation?tree_id=tree-1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestGetRecommendations_MissingTreeID_Returns400(t *testing.T) {
	svc := &mockService{}
	h, app := newTestApp(svc)

	app.Get("/reflect/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reflect/recommendation", nil) // no tree_id
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if _, ok := body["error"]; !ok {
		t.Errorf("Expected 'error' key in response body, got %v", body)
	}
}

func TestGetRecommendations_ServiceInternalError_Returns500(t *testing.T) {
	svc := &mockService{
		recommendPathsFunc: func(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
			return nil, apperror.NewInternal("recommendation engine failed")
		},
	}
	h, app := newTestApp(svc)

	app.Get("/reflect/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reflect/recommendation?tree_id=tree-1", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if body["error"] != "internal server error" {
		t.Errorf("Expected 'internal server error', got %v", body["error"])
	}
}

func TestGetRecommendations_ServiceRawError_Returns500(t *testing.T) {
	svc := &mockService{
		recommendPathsFunc: func(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
			return nil, errors.New("unexpected panic in recommendation pipeline")
		},
	}
	h, app := newTestApp(svc)

	app.Get("/reflect/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reflect/recommendation?tree_id=tree-1", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", resp.StatusCode)
	}
}

func TestGetRecommendations_Success_Returns200(t *testing.T) {
	svc := &mockService{
		recommendPathsFunc: func(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
			return &model.RecommendPathResponse{UserPersonaQuery: "test query"}, nil
		},
	}
	h, app := newTestApp(svc)

	app.Get("/reflect/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reflect/recommendation?tree_id=tree-1", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if body["success"] != true {
		t.Errorf("Expected success=true, got %v", body["success"])
	}
}

// ── GetHomeRecommendations ────────────────────────────────────────────────────

func TestGetHomeRecommendations_NoAuth_Returns401(t *testing.T) {
	svc := &mockService{}
	h, app := newTestApp(svc)

	app.Get("/home/recommendation", h.GetHomeRecommendations)

	req := httptest.NewRequest(http.MethodGet, "/home/recommendation", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

func TestGetHomeRecommendations_ServiceError_Returns500(t *testing.T) {
	svc := &mockService{
		recommendHomePathsFunc: func(ctx context.Context, userID string) (*model.RecommendPathResponse, error) {
			return nil, apperror.NewInternal("home recommendation fetch failed")
		},
	}
	h, app := newTestApp(svc)

	app.Get("/home/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetHomeRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/home/recommendation", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", resp.StatusCode)
	}
}

func TestGetHomeRecommendations_Success_Returns200(t *testing.T) {
	svc := &mockService{
		recommendHomePathsFunc: func(ctx context.Context, userID string) (*model.RecommendPathResponse, error) {
			return &model.RecommendPathResponse{}, nil
		},
	}
	h, app := newTestApp(svc)

	app.Get("/home/recommendation", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.GetHomeRecommendations(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/home/recommendation", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

// ── TriggerBatchRecommendation ────────────────────────────────────────────────

func TestTriggerBatch_NoAdminRole_Returns403(t *testing.T) {
	svc := &mockService{}
	h, app := newTestApp(svc)

	// Authenticated but role is "student", not "admin"
	app.Post("/batch/recommendation/trigger", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		c.Locals("role", "student")
		return h.TriggerBatchRecommendation(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/batch/recommendation/trigger", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if body["error"] != "admin access required" {
		t.Errorf("Expected 'admin access required', got %v", body["error"])
	}
}

func TestTriggerBatch_NoRoleSet_Returns403(t *testing.T) {
	svc := &mockService{}
	h, app := newTestApp(svc)

	// user_id set but no role local (simulates missing role claim)
	app.Post("/batch/recommendation/trigger", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return h.TriggerBatchRecommendation(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/batch/recommendation/trigger", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}

// DeleteError mapping: when SaveBatchRecommendations fails (delete step inside reconcile),
// the raw error propagates up and handleError maps it to 500.
func TestTriggerBatch_DeleteErrorMapsTo202(t *testing.T) {
	svc := &mockService{
		runBatchFunc: func(ctx context.Context) error {
			return errors.New("failed to delete old recommendations for user u1: deadlock detected")
		},
	}
	h, app := newTestApp(svc)

	app.Post("/batch/recommendation/trigger", func(c *fiber.Ctx) error {
		c.Locals("user_id", "admin-1")
		c.Locals("role", "admin")
		return h.TriggerBatchRecommendation(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/batch/recommendation/trigger", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202 for async batch trigger, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if body["success"] != true {
		t.Errorf("Expected success=true, got %v", body["success"])
	}
}

func TestTriggerBatch_AdminSuccess_Returns202(t *testing.T) {
	svc := &mockService{
		runBatchFunc: func(ctx context.Context) error { return nil },
	}
	h, app := newTestApp(svc)

	app.Post("/batch/recommendation/trigger", func(c *fiber.Ctx) error {
		c.Locals("user_id", "admin-1")
		c.Locals("role", "admin")
		return h.TriggerBatchRecommendation(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/batch/recommendation/trigger", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202, got %d", resp.StatusCode)
	}
	body := parseBody(t, resp)
	if body["success"] != true {
		t.Errorf("Expected success=true, got %v", body["success"])
	}
}
