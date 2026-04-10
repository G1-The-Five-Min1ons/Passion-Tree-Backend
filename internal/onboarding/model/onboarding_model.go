package model

import "time"

// SaveOnboardingRequest is the request body from the client
type SaveOnboardingRequest struct {
	Subjects        []string `json:"subjects"`         // multiple selections
	KnowledgeLevel  string   `json:"knowledge_level"`  // single selection
	Motivation      string   `json:"motivation"`       // single selection
	DailyGoal       string   `json:"daily_goal"`       // single selection
	LearningStyles  []string `json:"learning_styles"`  // multiple selections
	ReflectionHabit string   `json:"reflection_habit"` // single selection
}

// OnboardingData represents a row in the onboarding_answers table
type OnboardingData struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Subjects        string     `json:"subjects"`         // stored as comma-separated
	KnowledgeLevel  string     `json:"knowledge_level"`
	Motivation      string     `json:"motivation"`
	DailyGoal       string     `json:"daily_goal"`
	LearningStyles  string     `json:"learning_styles"`  // stored as comma-separated
	ReflectionHabit string     `json:"reflection_habit"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}
