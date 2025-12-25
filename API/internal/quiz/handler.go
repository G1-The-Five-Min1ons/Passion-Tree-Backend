package quiz

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	quizGroup := r.Group("/quiz")
	{
		quizGroup.POST("", h.CreateQuiz)
		quizGroup.GET("/:quiz_id", h.GetQuiz)
		quizGroup.PUT("/:quiz_id", h.UpdateQuiz)
		quizGroup.DELETE("/:quiz_id", h.DeleteQuiz)
	}
}

func (h *Handler) CreateQuiz(c *gin.Context) {
	var req CreateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.CreateQuiz(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "quiz created successfully",
		"quiz_id": id,
	})
}

func (h *Handler) GetQuiz(c *gin.Context) {
	quizIDStr := c.Param("quiz_id")
	quizID, _ := strconv.Atoi(quizIDStr)

	quiz, err := h.svc.GetQuiz(quizID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quiz not found"})
		return
	}

	c.JSON(http.StatusOK, quiz)
}

func (h *Handler) UpdateQuiz(c *gin.Context) {
	quizIDStr := c.Param("quiz_id")
	quizID, _ := strconv.Atoi(quizIDStr)

	var req UpdateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateQuiz(quizID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quiz updated successfully"})
}

func (h *Handler) DeleteQuiz(c *gin.Context) {
	quizIDStr := c.Param("quiz_id")
	quizID, _ := strconv.Atoi(quizIDStr)

	if err := h.svc.DeleteQuiz(quizID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quiz deleted successfully"})
}