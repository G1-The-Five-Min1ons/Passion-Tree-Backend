package handler

import (
	"net/http"
	"strings"
	"database/sql"
	"Passion-Tree-Backend/internal/reflection/model"
	"Passion-Tree-Backend/internal/reflection/service"
)

type ReflectionHandler struct {
	service service.ReflectionService
}

func NewReflectionHandler(s service.ReflectionService) *ReflectionHandler {
	return &ReflectionHandler{service: s}
}

func (h *ReflectionHandler) Create(c *gin.Context) {
	var req model.CreateReflectionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_json",
			"message": "Request body is not valid JSON",
		})
		return
	}

	res, err := h.service.CreateReflection(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "create_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "reflection created successfully",
		"data":    res,
	})
}

func (h *ReflectionHandler) Update(c *gin.Context) {

	id := c.Param("id")

	var req model.UpdateReflectionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_json",
		})
		return
	}

	if err := h.service.UpdateReflection(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reflection updated successfully",
	})
}

func (h *ReflectionHandler) Delete(c *gin.Context) {

	id := c.Param("reflect_id")

	err := h.service.DeleteReflection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "delete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reflection deleted successfully",
	})
}
