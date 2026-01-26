package model

import (
	"time"
)

type Reflection struct {
    ReflectID        string    `json:"reflect_id"`
    ReflectScore     string    `json:"reflect_score"`
    ReflectDescription string  `json:"reflect_description"`
    Reflect          string    `json:"reflect"`
    Mood             string    `json:"mood"`
    Tag              string    `json:"tag"`
    ProgressScore    string    `json:"progress_score"`
    ChallengeScore   string    `json:"challenge_score"`
    CreatedAt        time.Time `json:"created_at"`
    TreeNodeID       string    `json:"tree_node_id"`
}

type CreateReflectionRequest struct {
    Learned        string `json:"learned"`
    FeelScore      string `json:"feel_score"`
    Reflect        string `json:"reflect"`
    ProgressScore  string `json:"progress_score"`
    ChallengeScore string `json:"challenge_score"`
    Mood           string `json:"mood"`
    Tag            string `json:"tag"`
    TreeNodeID     string `json:"tree_node_id"`
}


type ReflectionResponse struct {
    ReflectID       string                 `json:"reflect_id"`
    Sentiment       string                 `json:"sentiment"`
    ReflectionScore float64                `json:"reflection_score"`
    Summary         string                 `json:"summary"`
    Advanced        *AdvancedMetrics       `json:"advanced,omitempty"`
    DevelopmentPlan *DevelopmentPlan       `json:"development_plan,omitempty"`
    RerankedResults []string               `json:"reranked_results,omitempty"`
}

type AdvancedMetrics struct {
    PrimaryEmotion      string  `json:"primary_emotion"`
    ConfidenceScore     float64 `json:"confidence_score"`
    StrugglePoint       string  `json:"struggle_point"`
    LearningDisposition string  `json:"learning_disposition"`
    ConsistencyCheck    string  `json:"consistency_check"`
}

type DevelopmentPlan struct {
    NextSteps []string `json:"next_steps"`
}

type UpdateReflectionRequest struct {
    Learned        string `json:"learned"`
    FeelScore      string `json:"feel_score"`
    Reflect        string `json:"reflect"`
    ProgressScore  string `json:"progress_score"`
    ChallengeScore string `json:"challenge_score"`
    Mood           string `json:"mood"`
    Tag            string `json:"tag"`
}
