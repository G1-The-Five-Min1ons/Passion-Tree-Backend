package learningpath

import (
	"database/sql"
	"passiontree/internal/database"
	"time"
	"github.com/google/uuid"
)

type Repository interface {
	GetAllLearnningPath() ([]LearningPath, error)
	GetLearnningPathByID(id string) (*LearningPath, error)
	CreateLearnningPath(req CreatePathRequest) (string, error)
	UpdateLearnningPath(id string, req UpdatePathRequest) error
	DeleteLearnningPath(id string) error

	EnrollLearnningPathUser(pathID string, userID string) error
	GetLearnningPathEnrollmentStatus(pathID string, userID string) (*PathEnroll, error)

	CreateNode(req CreateNodeRequest) (string, error)
	GetNodesByPathID(pathID string) ([]Node, error)
	UpdateNode(nodeID string, req UpdateNodeRequest) error
	DeleteNode(nodeID string) error
	CreateMaterial(req CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(nodeID string) ([]NodeMaterial, error)
	DeleteMaterial(materialID string) error

	CreateComment(req CreateCommentRequest) (string, error)
	GetCommentsByNodeID(nodeID string) ([]NodeComment, error)
	DeleteComment(commentID string) error
	CreateReaction(req CreateReactionRequest) error
	GetReactionsByCommentID(commentID string) ([]CommentReaction, error)
	CreateMention(req CreateMentionRequest) (string, error)

	CreateQuestion(req CreateQuestionRequest) (string, error)
	GetQuestionsByNodeID(nodeID string) ([]NodeQuestion, error)
	DeleteQuestion(questionID string) error
	CreateChoice(req CreateChoiceRequest) (string, error)
	GetChoicesByQuestionID(questionID string) ([]QuestionChoice, error)
	DeleteChoice(choiceID string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(ds database.Database) Repository {
	return &repository{
		db: ds.GetDB(),
	}
}

func (r *repository) GetAllLearnningPath() ([]LearningPath, error) {
	query := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []LearningPath
	for rows.Next() {
		var p LearningPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repository) GetLearnningPathByID(id string) (*LearningPath, error) {
	pathQuery := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path 
		WHERE path_id = ?`

	var p LearningPath
	err := r.db.QueryRow(pathQuery, id).Scan(
		&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID,
	)
	if err != nil {
		return nil, err
	}

	nodes, err := r.GetNodesByPathID(id)
	if err != nil {
		return &p, nil
	}

	for i := range nodes {
		materials, err := r.GetMaterialsByNodeID(nodes[i].NodeID)
		if err == nil {
			nodes[i].Materials = materials
		}
	}

	p.Nodes = nodes
	return &p, nil
}

func (r *repository) CreateLearnningPath(req CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, status, creator_ID, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, 0.0, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, req.CreatorID, now, now)
	if err != nil {
		return "", err
	}
	return newID, nil
}

func (r *repository) UpdateLearnningPath(id string, req UpdatePathRequest) error {
	query := `
		UPDATE learning_path 
		SET title=?, objective=?, description=?, cover_img_url=?, status=?, update_at=? 
		WHERE path_id=?`

	_, err := r.db.Exec(query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, time.Now(), id)
	return err
}

func (r *repository) DeleteLearnningPath(id string) error {
	_, err := r.db.Exec("DELETE FROM learning_path WHERE path_id = ?", id)
	return err
}

func (r *repository) EnrollLearnningPathUser(pathID string, userID string) error {
	enrollID := uuid.New().String()
	now := time.Now()
	query := `
		INSERT INTO path_enroll (enroll_id, user_id, path_id, status, enroll_at)
		VALUES (?, ?, ?, 'active', ?)`

	_, err := r.db.Exec(query, enrollID, userID, pathID, now)
	return err
}

func (r *repository) GetLearnningPathEnrollmentStatus(pathID string, userID string) (*PathEnroll, error) {
	query := `
		SELECT enroll_id, status, enroll_at, complete_at 
		FROM path_enroll 
		WHERE user_id = ? AND path_id = ?`

	var pe PathEnroll
	err := r.db.QueryRow(query, userID, pathID).Scan(&pe.EnrollID, &pe.Status, &pe.EnrollAt, &pe.CompleteAt)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}

func (r *repository) CreateNode(req CreateNodeRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node (node_id, title, description, path_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Title, req.Description, req.PathID)
	return id, err
}

func (r *repository) GetNodesByPathID(pathID string) ([]Node, error) {
	query := `SELECT node_id, title, description, path_id FROM node WHERE path_id = ?`
	rows, err := r.db.Query(query, pathID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID); err != nil {
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (r *repository) UpdateNode(nodeID string, req UpdateNodeRequest) error {
	query := `UPDATE node SET title=?, description=? WHERE node_id=?`
	_, err := r.db.Exec(query, req.Title, req.Description, nodeID)
	return err
}

func (r *repository) DeleteNode(nodeID string) error {
	_, err := r.db.Exec(`DELETE FROM node WHERE node_id = ?`, nodeID)
	return err
}

func (r *repository) CreateMaterial(req CreateMaterialRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_material (material_id, type, url, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Type, req.URL, req.NodeID)
	return id, err
}

func (r *repository) GetMaterialsByNodeID(nodeID string) ([]NodeMaterial, error) {
	query := `SELECT material_id, type, url, node_id FROM node_material WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []NodeMaterial
	for rows.Next() {
		var m NodeMaterial
		if err := rows.Scan(&m.MaterialID, &m.Type, &m.URL, &m.NodeID); err != nil {
			continue
		}
		materials = append(materials, m)
	}
	return materials, nil
}

func (r *repository) DeleteMaterial(materialID string) error {
	_, err := r.db.Exec(`DELETE FROM node_material WHERE material_id = ?`, materialID)
	return err
}

func (r *repository) CreateComment(req CreateCommentRequest) (string, error) {
	id := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO node_comment (comment_id, parent_id, content, create_at, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ParentID, req.Content, now, req.NodeID)
	return id, err
}

func (r *repository) GetCommentsByNodeID(nodeID string) ([]NodeComment, error) {
	query := `SELECT comment_id, parent_id, content, create_at, edit_at FROM node_comment WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []NodeComment
	for rows.Next() {
		var c NodeComment
		if err := rows.Scan(&c.CommentID, &c.ParentID, &c.Content, &c.CreatedAt, &c.EditAt); err != nil {
			continue
		}
		reactions, _ := r.GetReactionsByCommentID(c.CommentID)
		c.Reactions = reactions

		comments = append(comments, c)
	}
	return comments, nil
}

func (r *repository) DeleteComment(commentID string) error {
	_, err := r.db.Exec(`DELETE FROM node_comment WHERE comment_id = ?`, commentID)
	return err
}

func (r *repository) CreateReaction(req CreateReactionRequest) error {
	id := uuid.New().String()
	query := `INSERT INTO comment_reaction (reaction_id, reaction_type, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ReactionType, req.CommentID)
	return err
}

func (r *repository) GetReactionsByCommentID(commentID string) ([]CommentReaction, error) {
	query := `SELECT reaction_id, reaction_type, comment_id FROM comment_reaction WHERE comment_id = ?`
	rows, err := r.db.Query(query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []CommentReaction
	for rows.Next() {
		var rc CommentReaction
		if err := rows.Scan(&rc.ReactionID, &rc.ReactionType, &rc.CommentID); err != nil {
			continue
		}
		reactions = append(reactions, rc)
	}
	return reactions, nil
}

func (r *repository) CreateMention(req CreateMentionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO comment_mention (reaction_id, create_at, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, id, time.Now(), req.CommentID)
	return id, err
}

func (r *repository) CreateQuestion(req CreateQuestionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_question (question_id, question_text, type, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.QuestionText, req.Type, req.NodeID)
	return id, err
}

func (r *repository) GetQuestionsByNodeID(nodeID string) ([]NodeQuestion, error) {
	query := `SELECT question_id, question_text, type, node_id FROM node_question WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []NodeQuestion
	for rows.Next() {
		var q NodeQuestion
		if err := rows.Scan(&q.QuestionID, &q.QuestionText, &q.Type, &q.NodeID); err != nil {
			continue
		}
		choices, _ := r.GetChoicesByQuestionID(q.QuestionID)
		q.Choices = choices
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *repository) DeleteQuestion(questionID string) error {
	_, err := r.db.Exec(`DELETE FROM node_question WHERE question_id = ?`, questionID)
	return err
}

func (r *repository) CreateChoice(req CreateChoiceRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO question_choice (choice_id, choice_text, is_correct, reasoning, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ChoiceText, req.IsCorrect, req.Reasoning, req.QuestionID)
	return id, err
}

func (r *repository) GetChoicesByQuestionID(questionID string) ([]QuestionChoice, error) {
	query := `SELECT choice_id, choice_text, is_correct, reasoning, node_id FROM question_choice WHERE node_id = ?`
	rows, err := r.db.Query(query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []QuestionChoice
	for rows.Next() {
		var c QuestionChoice
		if err := rows.Scan(&c.ChoiceID, &c.ChoiceText, &c.IsCorrect, &c.Reasoning, &c.QuestionID); err != nil {
			continue
		}
		choices = append(choices, c)
	}
	return choices, nil
}

func (r *repository) DeleteChoice(choiceID string) error {
	_, err := r.db.Exec(`DELETE FROM question_choice WHERE choice_id = ?`, choiceID)
	return err
}