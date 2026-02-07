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
	CleanupTimeout         = 10 * time.Minute
	MaxConcurrentDeletions = 10
)

type CleanupWorker struct {
	db      connection.Database
	storage *storage.BlobService
}

func NewCleanupWorker(db connection.Database, storage *storage.BlobService) *CleanupWorker {
	return &CleanupWorker{
		db:      db,
		storage: storage,
	}
}

func (w *CleanupWorker) RunCleanup() {
	log.Println("[Cleanup Worker] Starting cleanup task...")
	ctx, cancel := context.WithTimeout(context.Background(), CleanupTimeout)
	defer cancel()

	query := "SELECT cover_img_url FROM learning_paths WHERE cover_img_url IS NOT NULL AND cover_img_url != ''"
	rows, err := w.db.GetDB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[Cleanup Worker] Error querying DB: %v\n", err)
		return
	}
	defer rows.Close()

	validFiles := make(map[string]struct{})

	for rows.Next() {
		var fullURL string
		if err := rows.Scan(&fullURL); err != nil {
			log.Printf("[Cleanup Worker] Warning: Failed to scan row: %v\n", err)
			continue
		}
		fileName := extractFilenameFromURL(fullURL)
		if fileName != "" {
			validFiles[fileName] = struct{}{}
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("[Cleanup Worker] Error iterating DB rows: %v\n", err)
		return
	}

	blobs, err := w.storage.ListBlobsOlderThan(ctx, "learning-path", 24*time.Hour)
	if err != nil {
		log.Printf("[Cleanup Worker] Error listing blobs from Azure: %v\n", err)
		return
	}

	var (
		deletedFiles []string
		errorCount   int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	sem := make(chan struct{}, MaxConcurrentDeletions)

	for _, blobName := range blobs {
		if ctx.Err() != nil {
			log.Println("[Cleanup Worker] Timeout reached, stopping cleanup.")
			break
		}

		if _, exists := validFiles[blobName]; !exists {

			wg.Add(1)

			go func(name string) {
				defer wg.Done()

				// Acquire Token: ถ้า Channel เต็ม จะ Block รอตรงนี้ (จำกัดความเร็ว)
				sem <- struct{}{}
				defer func() { <-sem }() // Release Token เมื่อทำเสร็จ

				// Double check context ใน Goroutine เผื่อโดน cancel ระหว่างรอคิว
				if ctx.Err() != nil {
					return
				}

				// ลบไฟล์ (ส่ง ctx ลงไปด้วย ถ้า timeout มันจะยกเลิก request เอง)
				err := w.storage.DeleteBlob(ctx, name, "learning-path")

				// Update ผลลัพธ์ (ต้อง Lock เพราะเขียนตัวแปร shared)
				mu.Lock()
				if err != nil {
					log.Printf("[Cleanup Worker] Failed to delete blob %s: %v\n", name, err)
					errorCount++
				} else {
					deletedFiles = append(deletedFiles, name)
				}
				mu.Unlock()

			}(blobName)
		}
	}

	// รอให้ Goroutines ทั้งหมดทำงานเสร็จ
	wg.Wait()

	if len(deletedFiles) > 0 {
		log.Printf("[Cleanup Worker] Cleanup finished. Deleted %d files. Errors: %d. Files: [%s]\n",
			len(deletedFiles),
			errorCount,
			strings.Join(deletedFiles, ", "),
		)
	} else {
		log.Println("[Cleanup Worker] Cleanup finished. No orphaned files found.")
	}
}

func extractFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
