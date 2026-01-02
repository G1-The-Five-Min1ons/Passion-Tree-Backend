package learningpath

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
	paths := r.Group("/learningpaths")
	{
		paths.Get("", h.GetAll)
		paths.Post("", h.Create)
		paths.Get("/:path_id", h.GetOne)
		paths.Put("/:path_id", h.Update)
		paths.Delete("/:path_id", h.Delete)
		paths.Post("/:path_id/start", h.Start)
		paths.Post("/:path_id/nodes", h.CreateNode)
	}

	nodes := r.Group("/learningpaths/nodes")
	{
		nodes.Put("/:node_id", h.UpdateNode)
		nodes.Delete("/:node_id", h.DeleteNode)
		nodes.Post("/:node_id/materials", h.CreateMaterial)
		nodes.Get("/:node_id/comments", h.GetComments)
		nodes.Post("/:node_id/comments", h.CreateComment)
		nodes.Get("/:node_id/questions", h.GetQuestions)
		nodes.Post("/:node_id/questions", h.CreateQuestion)
	}

	questions := r.Group("/learningpaths/questions")
	{
		questions.Delete("/:question_id", h.DeleteQuestion)
		questions.Post("/:question_id/choices", h.CreateChoice)
	}

	userPaths := r.Group("/user/learningpaths")
	{
		userPaths.Get("/:path_id/status", h.GetEnrollmentStatus)
	}

	r.Post("/learningpaths/comments/:comment_id/mentions", h.CreateMention)
	r.Post("/learningpaths/comments/:comment_id/reactions", h.CreateReaction)
	r.Delete("/learningpaths/comments/:comment_id", h.DeleteComment)
	r.Delete("/learningpaths/choices/:choice_id", h.DeleteChoice)
	r.Delete("/learningpaths/materials/:material_id", h.DeleteMaterial)
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	paths, err := h.svc.GetPaths()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": paths})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	id, err := h.svc.CreatePath(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "created", "path_id": id})
}

func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	path, err := h.svc.GetPathDetails(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "path not found or error fetching data"})
	}
	return c.Status(fiber.StatusOK).JSON(path)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("path_id")
	var req UpdatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.svc.UpdatePath(id, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "updated"})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("path_id")
	if err := h.svc.DeletePath(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "deleted"})
}

func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req StartPathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.StartPath(pathID, req.UserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "enrolled successfully"})
}

func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	status, err := h.svc.GetEnrollmentStatus(pathID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": status})
}

func (h *Handler) CreateNode(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	var req CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.PathID = pathID

	id, err := h.svc.AddNode(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"node_id": id})
}

func (h *Handler) UpdateNode(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.svc.EditNode(nodeID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "node updated"})
}

func (h *Handler) DeleteNode(c *fiber.Ctx) error {
	if err := h.svc.RemoveNode(c.Params("node_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "node deleted"})
}

func (h *Handler) CreateMaterial(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req CreateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.svc.AddMaterial(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"material_id": id})
}

func (h *Handler) DeleteMaterial(c *fiber.Ctx) error {
	if err := h.svc.RemoveMaterial(c.Params("material_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "material deleted"})
}

func (h *Handler) GetComments(c *fiber.Ctx) error {
	comments, err := h.svc.GetNodeComments(c.Params("node_id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": comments})
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.svc.AddComment(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"comment_id": id})
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	if err := h.svc.RemoveComment(c.Params("comment_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "comment deleted"})
}

func (h *Handler) CreateReaction(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.CommentID = commentID

	if err := h.svc.AddReaction(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reaction added"})
}

func (h *Handler) CreateMention(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req CreateMentionRequest
	c.BodyParser(&req)
	req.CommentID = commentID

	id, err := h.svc.AddMention(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"mention_id": id})
}

func (h *Handler) GetQuestions(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	questions, err := h.svc.GetQuestions(nodeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": questions})
}

func (h *Handler) CreateQuestion(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req CreateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.svc.AddQuestion(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"question_id": id})
}

func (h *Handler) DeleteQuestion(c *fiber.Ctx) error {
	if err := h.svc.RemoveQuestion(c.Params("question_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "question deleted"})
}

func (h *Handler) CreateChoice(c *fiber.Ctx) error {
	questionID := c.Params("question_id")
	var req CreateChoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.QuestionID = questionID

	id, err := h.svc.AddChoice(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"choice_id": id})
}

func (h *Handler) DeleteChoice(c *fiber.Ctx) error {
	if err := h.svc.RemoveChoice(c.Params("choice_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "choice deleted"})
}