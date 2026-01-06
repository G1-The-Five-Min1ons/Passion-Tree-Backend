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
    TreeNodeID     string `json:"tree_node_id"`
}


type ReflectionResponse struct {
    ID        string `json:"id"`
    Score     string `json:"score"`
    Mood      string `json:"mood"`
    Summary   string `json:"summary"`
    CreatedAt string `json:"created_at"`
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
