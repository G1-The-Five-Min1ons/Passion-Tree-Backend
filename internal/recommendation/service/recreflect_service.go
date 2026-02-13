package service

import (
	"context"
	"Strings"
	"passiontree/internal/recommendation/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) RecommendPathsForUser(ctx context.Context, user_id string) (*model.RecommendPathResponse, error) {
	s.logger.InfoContext(ctx, "start generating personalized recommendation", "user_id", user_id)

	// 1. ดึง Reflect ทุก Node ของ User
	// (ต้อง cast repo ให้รองรับ interface ใหม่ หรือใส่รวมใน repositoryImpl ของคุณ)
	reflections, err := s.recreflectRepo.(repository.RepositoryRecommendation).GetUserReflectionsAllNodes(ctx, user_id)
	if err != nil {
		return nil, apperror.NewInternal("failed to fetch user reflections: %w", err)
	}

	// 2. ถ้า User เป็นมือใหม่ (ไม่มี Reflect เลย) -> Cold Start Problem
	if len(reflections) == 0 {
		s.logger.InfoContext(ctx, "no reflections found, returning default recommendation", "user_id", user_id)
		// TODO: แนะนำ Path ยอดฮิต หรือ Path พื้นฐาน
		return &model.RecommendPathResponse{
			UserPersonaQuery: "New User",
		}, nil
	}

	// 3. Synthesize Data (สร้าง User Persona Vector Text)
	var personaBuilder strings.Builder
	var totalChallenge, totalProgress int
	var tags []string

	personaBuilder.WriteString("User is looking for a learning path. Historical feedback: ")

	// ลูปประกอบร่างข้อความจาก Reflect ย้อนหลัง (จำกัดแค่ 10 อันล่าสุดเพื่อไม่ให้ Noise เยอะไป)
	limit := len(reflections)
	if limit > 10 { limit = 10 }

	for i := 0; i < limit; i++ {
		ref := reflections[i]
		totalChallenge += ref.ChallengeScore
		totalProgress += ref.ProgressScore
		
		if ref.Tag != "" {
			tags = append(tags, ref.Tag)
		}
		if ref.ReflectDescription != "" {
			personaBuilder.WriteString(fmt.Sprintf("[%s], ", ref.ReflectDescription))
		}
	}

	avgChallenge := float64(totalChallenge) / float64(limit)
	
	// สรุปพฤติกรรมความยาก (Inferring Difficulty Preference)
	if avgChallenge > 180 {
		personaBuilder.WriteString(" User prefers basic, fundamental, and easy-to-understand explanations.")
	} else if avgChallenge < 80 {
		personaBuilder.WriteString(" User prefers advanced, challenging, and complex project-based topics.")
	}

	userPersonaQuery := personaBuilder.String()
	s.logger.InfoContext(ctx, "generated user persona", "query", userPersonaQuery, "avg_challenge", avgChallenge)

	// 4. เรียกใช้ AI/Qdrant Client เพื่อ Search (สมมติว่า aiclient มีฟังก์ชัน SemanticSearch)
	// ระบบจะแปลง userPersonaQuery เป็น Vector แล้วไปเทียบกับ Description ของ Learning Paths
	searchReq := model.SearchPathRequest{
		Query: userPersonaQuery,
		Limit: 10,
	}
	
	searchResult, err := s.aiClient.SearchLearningPaths(ctx, searchReq)
	if err != nil {
		return nil, apperror.NewInternal("vector search failed: %w", err)
	}

	// 5. Multi-Factor Ranking (ปรับคะแนนใหม่โดยนำ Behavior มา Weight)
	var finalRecommendations []model.RecommendedPath

	for _, path := range searchResult.Paths {
		// สมมติ SearchResult คืนค่า Cosine Score มาให้ในตัวแปร Score
		cosineScore := path.Score 
		
		// Weight 1: Cosine Similarity (ความหมายตรงกับ Reflect ไหม)
		finalScore := cosineScore * 0.7 

		// Weight 2: Behavior/Preference Boost (สมมติเทียบ Tag)
		// ถ้า Path มี keyword ตรงกับ tag ที่ user เคยเรียน ให้คะแนนเพิ่ม
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(path.Title), strings.ToLower(tag)) {
				finalScore += 0.1 
				break
			}
		}

		finalRecommendations = append(finalRecommendations, model.RecommendedPath{
			LearningPath:        path.LearningPath,
			RecommendationScore: finalScore,
			Reason:              fmt.Sprintf("Matches your learning style. Avg Challenge Score: %.1f", avgChallenge),
		})
	}

	// (Optional) Sort finalRecommendations ตาม finalScore ลงมาอีกที

	return &model.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}