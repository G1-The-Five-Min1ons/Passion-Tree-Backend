package quiz

type Quiz struct {
	ID            int      `json:"quiz_id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
}

type CreateQuizRequest struct {
	Question      string   `json:"question" binding:"required"`
	Options       []string `json:"options" binding:"required,min=2"`
	CorrectAnswer int      `json:"correct_answer" binding:"gte=0"`
	Explanation   string   `json:"explanation"`
}

type UpdateQuizRequest struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
}