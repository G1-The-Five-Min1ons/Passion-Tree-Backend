package recommendation

type Service interface {
	GetRecommendations(userID int) ([]RecommendedItem, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetRecommendations(userID int) ([]RecommendedItem, error) {
	if userID == 0 {
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