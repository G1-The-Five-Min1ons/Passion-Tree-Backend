package worker

import (
	"context"
	"log"
	"passiontree/internal/connection"
	"passiontree/internal/pkg/storage"
	"strings"
	"sync"
	"time"
)

const (
	CleanupTimeout         = 30 * time.Minute
	MaxConcurrentDeletions = 10
)

type CleanupWorker struct {
	db      connection.Database
	storage *storage.BlobService
}

type CleanupTarget struct {
	Folder string
	Table  string
	Column string
}

func NewCleanupWorker(db connection.Database, storage *storage.BlobService) *CleanupWorker {
	return &CleanupWorker{
		db:      db,
		storage: storage,
	}
}

func (w *CleanupWorker) RunCleanup() {
	log.Println("[Cleanup Worker] Starting global cleanup task...")
	ctx, cancel := context.WithTimeout(context.Background(), CleanupTimeout)
	defer cancel()

	targets := []CleanupTarget{
		{Folder: "learning-path", Table: "learning_paths", Column: "cover_img_url"},
		{Folder: "profile", Table: "users", Column: "profile_img_url"},
		{Folder: "reflect", Table: "tree_album", Column: "cover_image_url"},
		{Folder: "materials-nodes", Table: "tree_nodes", Column: "material_url"},
	}

	for _, target := range targets {
		if ctx.Err() != nil {
			log.Println("[Cleanup Worker] Timeout reached, stopping cleanup.")
			break
		}
		w.processFolderCleanup(ctx, target)
	}

	log.Println("[Cleanup Worker] Global cleanup task completed.")
}

func (w *CleanupWorker) processFolderCleanup(ctx context.Context, target CleanupTarget) {
	log.Printf("[Cleanup Worker] Processing folder: %s (Check DB: %s.%s)\n", target.Folder, target.Table, target.Column)

	query := "SELECT " + target.Column + " FROM " + target.Table +
		" WHERE " + target.Column + " IS NOT NULL AND " + target.Column + " != ''"

	rows, err := w.db.GetDB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[Cleanup Worker] Error querying DB for %s: %v\n", target.Folder, err)
		return
	}
	defer rows.Close()

	validFiles := make(map[string]struct{})
	for rows.Next() {
		var fullURL string
		if err := rows.Scan(&fullURL); err != nil {
			continue
		}
		fileName := extractFilenameFromURL(fullURL)
		if fileName != "" {
			validFiles[fileName] = struct{}{}
		}
	}
	blobs, err := w.storage.ListBlobsOlderThan(ctx, target.Folder, 24*time.Hour)
	if err != nil {
		log.Printf("[Cleanup Worker] Error listing blobs for %s: %v\n", target.Folder, err)
		return
	}

	var (
		deletedCount int
		errorCount   int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	sem := make(chan struct{}, MaxConcurrentDeletions)

	for _, blobName := range blobs {
		if ctx.Err() != nil {
			break
		}

		if _, exists := validFiles[blobName]; !exists {
			wg.Add(1)

			go func(name string) {
				defer wg.Done()

				sem <- struct{}{}
				defer func() { <-sem }()

				if ctx.Err() != nil {
					return
				}

				err := w.storage.DeleteBlob(ctx, name, target.Folder)

				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					log.Printf("[Cleanup Worker] Failed to delete %s from %s: %v\n", name, target.Folder, err)
					errorCount++
				} else {
					deletedCount++
				}
			}(blobName)
		}
	}

	wg.Wait()

	if deletedCount > 0 || errorCount > 0 {
		log.Printf("[Cleanup Worker] Folder '%s' summary: Deleted %d files, Errors %d.\n", target.Folder, deletedCount, errorCount)
	} else {
		log.Printf("[Cleanup Worker] Folder '%s' is clean.\n", target.Folder)
	}
}

func extractFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		return filename
	}
	return ""
}
