package service_test

import (
	"reflect"
	"testing"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/service"
)

func TestParseAINodes(t *testing.T) {
	tests := []struct {
		name      string
		rawResult string
		expected  []model.GeneratedNode
	}{
		{
			name:      "EmptyResult",
			rawResult: "",
			expected:  nil,
		},
		{
			name:      "SingleNode",
			rawResult: "Node 1: Introduction to Go",
			expected: []model.GeneratedNode{
				{Sequence: 1, Title: "Introduction to Go"},
			},
		},
		{
			name:      "MultipleNodes",
			rawResult: "Node 1: Getting Started, Node 2: Advanced Concepts",
			expected: []model.GeneratedNode{
				{Sequence: 1, Title: "Getting Started"},
				{Sequence: 2, Title: "Advanced Concepts"},
			},
		},
		{
			name:      "InvalidFormatIgnored",
			rawResult: "Node 1: Valid format, Random text without node format",
			expected: []model.GeneratedNode{
				{Sequence: 1, Title: "Valid format"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ParseAINodes(tt.rawResult)
			if !reflect.DeepEqual(got, tt.expected) && (len(got) != 0 || len(tt.expected) != 0) {
				t.Errorf("service.ParseAINodes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
