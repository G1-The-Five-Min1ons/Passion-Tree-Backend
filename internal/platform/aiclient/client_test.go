package aiclient_test

import (
	"context"
	"strings"
	"testing"

	"passiontree/internal/platform/aiclient"
)

func TestSearch_NegativeTopK_ReturnsValidationError(t *testing.T) {
	tests := []struct {
		name    string
		topK    int
		wantErr string
	}{
		{
			name:    "NegativeOne",
			topK:    -1,
			wantErr: "top_k must be >= 1, got -1",
		},
		{
			name:    "LargeNegative",
			topK:    -100,
			wantErr: "top_k must be >= 1, got -100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := aiclient.NewAIClient("http://not-used-validation-only")
			_, err := client.Search(context.Background(), aiclient.SearchRequest{
				Query: "test query",
				TopK:  tt.topK,
			})
			if err == nil {
				t.Fatalf("Expected validation error for top_k=%d, got nil", tt.topK)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestSearch_ZeroTopK_UsesDefault(t *testing.T) {
	// top_k=0 is a valid sentinel that gets defaulted to 7; no validation error expected.
	// A real HTTP call is not made here — validation occurs before the HTTP layer.
	// We just confirm no error is returned for the validation gate itself.
	// (Full default-value behaviour is covered in integration / mock-server tests.)
	client := aiclient.NewAIClient("http://not-used-validation-only")
	_, err := client.Search(context.Background(), aiclient.SearchRequest{
		Query: "test",
		TopK:  0,
	})
	// Expect a network / connection error, NOT a validation error.
	if err != nil && strings.Contains(err.Error(), "top_k must be >= 1") {
		t.Errorf("top_k=0 should not trigger validation error, got: %v", err)
	}
}
