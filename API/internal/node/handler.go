package learningpath

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
	paths := r.Group("/paths")
	{
		paths.POST("/:path_id/nodes", h.CreateNode)
	}

	nodes := r.Group("/nodes")
	{
		nodes.GET("/:node_id", h.GetNode)
		nodes.PUT("/:node_id", h.UpdateNode)
		nodes.DELETE("/:node_id", h.DeleteNode)
	}
}

func (h *Handler) CreateNode(c *gin.Context) {
	pathIDStr := c.Param("path_id")
	pathID, _ := strconv.Atoi(pathIDStr)

	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodeID, err := h.svc.CreateNode(pathID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "node created successfully",
		"node_id": nodeID,
		"path_id": pathID,
	})
}

func (h *Handler) GetNode(c *gin.Context) {
	nodeIDStr := c.Param("node_id")
	nodeID, _ := strconv.Atoi(nodeIDStr)

	node, err := h.svc.GetNodeDetails(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

func (h *Handler) UpdateNode(c *gin.Context) {
	nodeIDStr := c.Param("node_id")
	nodeID, _ := strconv.Atoi(nodeIDStr)

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateNode(nodeID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node updated successfully"})
}

func (h *Handler) DeleteNode(c *gin.Context) {
	nodeIDStr := c.Param("node_id")
	nodeID, _ := strconv.Atoi(nodeIDStr)

	if err := h.svc.DeleteNode(nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node deleted successfully"})
}