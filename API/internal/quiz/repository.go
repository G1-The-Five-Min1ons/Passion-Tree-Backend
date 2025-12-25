package quiz

import "database/sql"

type Repository interface {
	CreateQuiz(req CreateQuizRequest) (int, error)
	GetQuizByID(quizID int) (*Quiz, error)
	UpdateQuiz(quizID int, req UpdateQuizRequest) error
	DeleteQuiz(quizID int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateQuiz(req CreateQuizRequest) (int, error) {
	return 501, nil
}

func (r *repository) GetQuizByID(quizID int) (*Quiz, error) {
	// return &Quiz{
	// 	ID:            quizID,
	// 	Question:      "What is the capital of Thailand?",
	// 	Options:       []string{"Chiang Mai", "Bangkok", "Phuket", "Pattaya"},
	// 	CorrectAnswer: 1, // Index 1 = Bangkok
	// 	Explanation:   "Bangkok is the capital city of Thailand.",
	// }, nil
	
	return nil, nil
}

func (r *repository) UpdateQuiz(quizID int, req UpdateQuizRequest) error {
	// จำลองว่า Update สำเร็จ
	return nil
}

func (r *repository) DeleteQuiz(quizID int) error {
	// จำลองว่า Delete สำเร็จ
	return nil
}