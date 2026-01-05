package handler

import (
	"passiontree/internal/reflection/service"
)

type ReflectionHandler struct {
	service service.ReflectionService
}

func NewReflectionHandler(s service.ReflectionService) *ReflectionHandler {
	return &ReflectionHandler{service: s}
}
