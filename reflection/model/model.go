package reflection
package model

import (
	"time"
	"github.com/google/uuid"
)

type Reflection struct {
    ReflectID        uuid.UUID `json:"reflect_id"`
    ReflectScore     int       `json:"reflect_score"`
    ReflectDescription string  `json:"reflect_description"`
    Reflect          string    `json:"reflect"`
    Mood             string    `json:"mood"`
    Tag              string    `json:"tag"`
    ProgressScore    int       `json:"progress_score"`
    ChallengeScore   int       `json:"challenge_score"`
    CreatedAt        time.Time `json:"created_at"`
    TreeNodeID       uuid.UUID `json:"tree_node_id"`
}

type CreateReflectionRequest struct {
    Learned        string `json:"learned"`
    FeelScore      int    `json:"feel_score"`
    Reflect        string `json:"reflect"`
    ProgressScore  int    `json:"progress_score"`
    ChallengeScore int    `json:"challenge_score"`
    TreeNodeID     string `json:"tree_node_id"`
}


type ReflectionResponse struct {
    ID        string `json:"id"`
    Score     int    `json:"score"`
    Mood      string `json:"mood"`
    Summary   string `json:"summary"`
    CreatedAt string `json:"created_at"`
}

type UpdateReflectionRequest struct {
    Learned        string `json:"learned"`
    FeelScore      int    `json:"feel_score"`
    Reflect        string `json:"reflect"`
    ProgressScore  int    `json:"progress_score"`
    ChallengeScore int    `json:"challenge_score"`
    Mood           string `json:"mood"`
    Tag            string `json:"tag"`
}
