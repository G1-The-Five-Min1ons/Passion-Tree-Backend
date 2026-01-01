package learningpath

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
	paths := r.Group("/paths")
	{
		paths.GET("", h.GetAll)
		paths.POST("", h.Create)
		paths.GET("/:path_id", h.GetOne)
		paths.PUT("/:path_id", h.Update)
		paths.DELETE("/:path_id", h.Delete)
		paths.POST("/:path_id/start", h.Start)
	}

	userPaths := r.Group("/user/paths")
	{
		userPaths.GET("/:path_id/status", h.GetEnrollmentStatus)
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	paths, err := h.svc.GetPaths()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": paths})
}

func (h *Handler) Create(c *gin.Context) {
	var req CreatePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.CreatePath(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "created", "path_id": id})
}

func (h *Handler) GetOne(c *gin.Context) {
	id := c.Param("path_id")
	
	path, err := h.svc.GetPathDetails(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "path not found or error fetching data"})
		return
	}
	c.JSON(http.StatusOK, path)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("path_id")
	
	var req UpdatePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.svc.UpdatePath(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("path_id")
	
	if err := h.svc.DeletePath(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) Start(c *gin.Context) {
	pathID := c.Param("path_id")
	
	var req StartPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.StartPath(pathID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "enrolled successfully"})
}

func (h *Handler) GetEnrollmentStatus(c *gin.Context) {
	pathID := c.Param("path_id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	status, err := h.svc.GetEnrollmentStatus(pathID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": status})
}