package recommendation

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r fiber.Router) {
	r.Post("/recommendation", h.PostGeneral)
	r.Post("/trees/recommendation", h.PostTree)
}

func (h *Handler) PostGeneral(c *fiber.Ctx) error {
	var req GeneralRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	items, err := h.svc.GetGeneralRecommendations(req.User_id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id": req.User_id,
		"data":    items,
	})
}

func (h *Handler) PostTree(c *fiber.Ctx) error {
	var req TreeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	items, err := h.svc.GetTreeRecommendations(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id": req.User_id,
		"tree_id": req.Tree_id,
		"data":    items,
	})
}