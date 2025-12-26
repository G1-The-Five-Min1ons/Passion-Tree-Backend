package tree

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
	albums := r.Group("/tree-albums")
	{
		albums.POST("", h.CreateAlbum)
		albums.GET("", h.GetAlbums)
		albums.PUT("/:album_id", h.UpdateAlbum)
		albums.DELETE("/:album_id", h.DeleteAlbum)

		albums.POST("/:album_id/trees", h.PlantTree)
		albums.GET("/:album_id/trees", h.GetAlbumTrees)
	}

	trees := r.Group("/trees")
	{
		trees.GET("/:tree_id", h.GetTree)
		
		trees.POST("/:tree_id/nodes", h.AddNode)
		trees.GET("/:tree_id/nodes/:node_id", h.GetNode)
		trees.PUT("/:tree_id/nodes/:node_id", h.UpdateNode)
		trees.DELETE("/:tree_id/nodes/:node_id", h.DeleteNode)
	}
}

func (h *Handler) CreateAlbum(c *gin.Context) {
	var req CreateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.CreateAlbum(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "album created", "album_id": id})
}

func (h *Handler) GetAlbums(c *gin.Context) {
	albums, _ := h.svc.GetAlbums()
	c.JSON(http.StatusOK, gin.H{"data": albums})
}

func (h *Handler) UpdateAlbum(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("album_id"))
	var req UpdateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.svc.UpdateAlbum(id, req)
	c.JSON(http.StatusOK, gin.H{"message": "album updated"})
}

func (h *Handler) DeleteAlbum(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("album_id"))
	h.svc.DeleteAlbum(id)
	c.JSON(http.StatusOK, gin.H{"message": "album deleted"})
}

func (h *Handler) PlantTree(c *gin.Context) {
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	var req CreateTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.PlantTree(albumID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "tree planted", "tree_id": id})
}

func (h *Handler) GetAlbumTrees(c *gin.Context) {
	albumID, _ := strconv.Atoi(c.Param("album_id"))
	trees, _ := h.svc.GetAlbumTrees(albumID)
	c.JSON(http.StatusOK, gin.H{"data": trees})
}

func (h *Handler) GetTree(c *gin.Context) {
	treeID, _ := strconv.Atoi(c.Param("tree_id"))
	tree, _ := h.svc.GetTreeDetails(treeID)
	c.JSON(http.StatusOK, tree)
}

func (h *Handler) AddNode(c *gin.Context) {
	treeID, _ := strconv.Atoi(c.Param("tree_id"))
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := h.svc.AddNode(treeID, req)
	c.JSON(http.StatusCreated, gin.H{"message": "node created", "node_id": id})
}

func (h *Handler) GetNode(c *gin.Context) {
	nodeID, _ := strconv.Atoi(c.Param("node_id"))
	node, _ := h.svc.GetNodeDetails(nodeID)
	c.JSON(http.StatusOK, node)
}

func (h *Handler) UpdateNode(c *gin.Context) {
	nodeID, _ := strconv.Atoi(c.Param("node_id"))
	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.svc.UpdateNodeReflection(nodeID, req)
	c.JSON(http.StatusOK, gin.H{"message": "node reflection updated"})
}

func (h *Handler) DeleteNode(c *gin.Context) {
	nodeID, _ := strconv.Atoi(c.Param("node_id"))
	h.svc.RemoveNode(nodeID)
	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}