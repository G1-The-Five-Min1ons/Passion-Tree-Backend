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
	r.POST("/recommendation", h.PostGeneral) 
	r.POST("/trees/:tree_id/recommendation", h.PostTree)
}

func (h *Handler) PostGeneral(c *gin.Context) {
	var req GeneralRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {}

	items, err := h.svc.GetGeneralRecommendations(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.UserID,
		"data":    items,
	})
}

func (h *Handler) PostTree(c *gin.Context) {
	treeIDStr := c.Param("tree_id")
	treeID, _ := strconv.Atoi(treeIDStr)

	var req TreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	items, err := h.svc.GetTreeRecommendations(treeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tree_id": treeID,
		"data":    items,
	})
}