package worker

import (
	"context"
	"log"
	"passiontree/internal/database"
	"strings"
	"time"
)

type CleanupWorker struct {
	db      database.Database
	storage *database.StorageClient
}

func NewCleanupWorker(db database.Database, storage *database.StorageClient) *CleanupWorker {
	return &CleanupWorker{
		db:      db,
		storage: storage,
	}
}

func (w *CleanupWorker) RunCleanup() {
	log.Println("[Cleanup Worker] Starting cleanup task...")
	ctx := context.Background()

	query := "SELECT cover_img_url FROM learning_paths WHERE cover_img_url IS NOT NULL AND cover_img_url != ''"
	rows, err := w.db.GetDB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[Cleanup Worker] Error querying DB: %v\n", err)
		return
	}
	defer rows.Close()

	validFiles := make(map[string]bool)
	for rows.Next() {
		var fullURL string
		if err := rows.Scan(&fullURL); err != nil {
			continue
		}
		fileName := extractFilenameFromURL(fullURL)
		if fileName != "" {
			validFiles[fileName] = true
		}
	}

	blobs, err := w.storage.ListBlobsOlderThan(ctx, "learning-path", 24*time.Hour)
	if err != nil {
		log.Printf("[Cleanup Worker] Error listing blobs from Azure: %v\n", err)
		return
	}

	deletedCount := 0
	for _, blobName := range blobs {
		if !validFiles[blobName] {
			log.Printf("[Cleanup Worker] Found orphaned file: %s. Deleting...", blobName)
			
			err := w.storage.DeleteBlob(ctx, blobName, "learning-path")
			if err != nil {
				log.Printf("[Cleanup Worker] Failed to delete blob %s: %v\n", blobName, err)
			} else {
				deletedCount++
			}
		}
	}

	log.Printf("[Cleanup Worker] Cleanup finished. Deleted %d files.\n", deletedCount)
}

func extractFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}