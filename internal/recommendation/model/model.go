package model

type UserReflection struct {
	ReflectID          string `json:"reflect_id"`
	ReflectDescription string `json:"reflect_description"`
	Mood               string `json:"mood"`
	Tag                string `json:"tag"`
	ChallengeScore     int    `json:"challenge_score"` // tinyint ใน DB
	ProgressScore      int    `json:"progress_score"`  // tinyint ใน DB
	TreeNodeID         string `json:"tree_node_id"`
}

type RecommendPathResponse struct {
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
	UserPersonaQuery string            `json:"user_persona_query"` // บอกว่าระบบมอง User คนนี้เป็นยังไง (สำหรับ Debug/โชว์ UI)
}

type RecommendedPath struct {
	LearningPath
	RecommendationScore float64 `json:"recommendation_score"` // คะแนนที่เกิดจาก Cosine + Skill Gap
	Reason              string  `json:"reason"`               // เหตุผลที่แนะนำ
}