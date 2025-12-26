package user

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
	userGroup := r.Group("/user")
	{
		userGroup.GET("", h.GetUserOverview)
		userGroup.POST("/sign-out-all", h.SignOutAll)
		userGroup.POST("/export", h.ExportData)

		userGroup.PUT("/:id", h.UpdateProfile)
		userGroup.GET("/:id/settings", h.GetSettings) 
		userGroup.PUT("/:id/preferences", h.UpdatePrefs)
		userGroup.POST("/:id/deactivate", h.Deactivate)
		userGroup.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) GetUserOverview(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, _ := strconv.Atoi(userIDStr)

	user, err := h.svc.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateBasicProfile(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *Handler) GetSettings(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	settings, err := h.svc.GetAccountSettings(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdatePrefs(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateUserPreferences(id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "preferences updated"})
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.DeactivateUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deactivated"})
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted permanently"})
}

func (h *Handler) SignOutAll(c *gin.Context) {
	var req SignOutAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required in body"})
		return
	}

	if err := h.svc.SignOutAll(req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "signed out from all devices"})
}

func (h *Handler) ExportData(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required in body"})
		return
	}

	jobID, err := h.svc.ExportUserData(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "export started", "job_id": jobID})
}