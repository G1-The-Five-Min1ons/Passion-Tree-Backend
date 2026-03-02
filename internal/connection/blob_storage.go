package connection

import (
	"fmt"
	"passiontree/internal/config"
	"passiontree/internal/pkg/storage"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// NewStorageClient สร้าง Azure Blob Storage client
func InitBlobStorage(cfg *config.Config) (*storage.BlobService, error) {
	if cfg.AzureStorageConnString == "" {
		return nil, fmt.Errorf("Azure Storage connection string is not configured")
	}

	// 1. สร้าง Low-level Client
	client, err := azblob.NewClientFromConnectionString(cfg.AzureStorageConnString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Storage client: %w", err)
	}

	// 2. Extract ข้อมูลที่จำเป็น
	accountName := extractAccountName(cfg.AzureStorageConnString)
	accountKey := extractAccountKey(cfg.AzureStorageConnString)
	containerMap := map[string]string{
		"learning-path":   cfg.ContainerLearningPath,
		"profile":         cfg.ContainerProfile,
		"reflect":         cfg.ContainerReflect,
		"materials-nodes": cfg.ContainerMaterialsNodes,
	}

	// 3. ส่งต่อ Client และ Config ไปให้ package storage จัดการต่อ
	return storage.NewBlobService(
		client,
		accountName,
		accountKey,
		containerMap,
	), nil
}

// extractAccountName ดึงชื่อ storage account จาก connection string
func extractAccountName(connString string) string {
	parts := strings.Split(connString, ";")

	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName=")
		}
	}

	return ""
}

func extractAccountKey(connString string) string {
	parts := strings.Split(connString, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountKey=") {
			return strings.TrimPrefix(part, "AccountKey=")
		}
	}
	return ""
}
