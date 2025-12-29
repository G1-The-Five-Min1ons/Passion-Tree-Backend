package recommendation

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/recommendation", h.PostGeneral)
	r.POST("/trees/recommendation", h.PostTree)
}

func (h *Handler) PostGeneral(c *gin.Context) {
	var req GeneralRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {}

	items, err := h.svc.GetGeneralRecommendations(req.User_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.User_id,
		"data":    items,
	})
}

func (h *Handler) PostTree(c *gin.Context) {
	var req TreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	items, err := h.svc.GetTreeRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.User_id,
		"tree_id": req.Tree_id,
		"data":    items,
	})
}