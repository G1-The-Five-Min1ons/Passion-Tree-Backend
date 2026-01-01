package recommendation

type Service interface {
	GetGeneralRecommendations(userID string) ([]RecommendedItem, error)
	GetTreeRecommendations(req TreeRequest) ([]RecommendedItem, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetGeneralRecommendations(userID string) ([]RecommendedItem, error) {
	if userID == "" {
		return s.repo.GetPopularItems()
	}

	items, err := s.repo.GetPersonalizedItems(userID)
	if err != nil {
		return nil, err
	}
	
	if len(items) == 0 {
		return s.repo.GetPopularItems()
	}

	return items, nil
}

func (s *service) GetTreeRecommendations(req TreeRequest) ([]RecommendedItem, error) {
	return s.repo.GetNextPathInTree(req.Tree_id, req.User_id)
}