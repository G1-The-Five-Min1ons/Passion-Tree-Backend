package recommendation

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
	routes := r.Group("/recommendations")
	{
		routes.GET("", h.Get) // GET /api/v1/recommendations
	}
}

func (h *Handler) Get(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, _ := strconv.Atoi(userIDStr)

	items, err := h.svc.GetRecommendations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"data":    items,
	})
}