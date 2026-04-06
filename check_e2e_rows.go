package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"passiontree/internal/config"
	"passiontree/internal/connection"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	backendEnv := filepath.Join(wd, ".env")
	infraEnv := filepath.Join(wd, "..", "Passion-Tree-Infrastructure", ".env")
	if err := godotenv.Load(backendEnv); err != nil {
		_ = godotenv.Load(infraEnv)
	}

	cfg, err := config.LoadDBConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load database config: %v\n", err)
		os.Exit(1)
	}

	db, err := connection.NewDatabase(cfg.DBConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()

	query := `
	SELECT TOP 10
		CONVERT(VARCHAR(36), lp.path_id) AS path_id,
		lp.title,
		CONVERT(VARCHAR(36), lp.creator_id) AS creator_id,
		ISNULL(node_count.total_nodes, 0) AS total_nodes,
		ISNULL(mat_count.total_materials, 0) AS total_materials,
		ISNULL(q_count.total_questions, 0) AS total_questions,
		ISNULL(c_count.total_choices, 0) AS total_choices,
		lp.create_at
	FROM learning_path lp
	LEFT JOIN (
		SELECT path_id, COUNT(*) AS total_nodes
		FROM node
		GROUP BY path_id
	) node_count ON node_count.path_id = lp.path_id
	LEFT JOIN (
		SELECT n.path_id, COUNT(*) AS total_materials
		FROM node_material nm
		JOIN node n ON n.node_id = nm.node_id
		GROUP BY n.path_id
	) mat_count ON mat_count.path_id = lp.path_id
	LEFT JOIN (
		SELECT n.path_id, COUNT(*) AS total_questions
		FROM node_question q
		JOIN node n ON n.node_id = q.node_id
		GROUP BY n.path_id
	) q_count ON q_count.path_id = lp.path_id
	LEFT JOIN (
		SELECT n.path_id, COUNT(*) AS total_choices
		FROM question_choice c
		JOIN node_question q ON q.question_id = c.question_id
		JOIN node n ON n.node_id = q.node_id
		GROUP BY n.path_id
	) c_count ON c_count.path_id = lp.path_id
	WHERE lp.title LIKE 'E2E LP%'
	ORDER BY lp.create_at DESC`

	rows, err := db.GetDB().QueryContext(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Latest E2E learning paths:")
	count := 0
	for rows.Next() {
		var pathID, title, creatorID, createdAt string
		var totalNodes, totalMaterials, totalQuestions, totalChoices int
		if err := rows.Scan(&pathID, &title, &creatorID, &totalNodes, &totalMaterials, &totalQuestions, &totalChoices, &createdAt); err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
		count++
		fmt.Printf("%d) path_id=%s | nodes=%d | materials=%d | questions=%d | choices=%d | creator=%s | title=%s | created_at=%s\n", count, pathID, totalNodes, totalMaterials, totalQuestions, totalChoices, creatorID, title, createdAt)
	}

	if count == 0 {
		fmt.Println("No E2E LP rows found.")
	}
}
