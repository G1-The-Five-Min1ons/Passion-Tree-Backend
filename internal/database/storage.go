package database

import (
	"context"
	"fmt"
	"passiontree/internal/config"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/google/uuid"
)

type StorageClient struct {
	client                *azblob.Client
	accountName           string
	containerLearningPath string
	containerProfile      string
}

// NewStorageClient สร้าง Azure Blob Storage client
func NewStorageClient(cfg *config.Config) (*StorageClient, error) {
	if cfg.AzureStorageConnString == "" {
		return nil, fmt.Errorf("Azure Storage connection string is not configured")
	}

	client, err := azblob.NewClientFromConnectionString(cfg.AzureStorageConnString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Storage client: %w", err)
	}

	// Extract account name from connection string for URL generation
	accountName := extractAccountName(cfg.AzureStorageConnString)

	return &StorageClient{
		client:                client,
		accountName:           accountName,
		containerLearningPath: cfg.ContainerLearningPath,
		containerProfile:      cfg.ContainerProfile,
	}, nil
}

// GenerateBlobURL สร้าง blob URL string จากชื่อไฟล์
func (s *StorageClient) GenerateBlobURL(filename, containerType string) string {
	containerName := s.getContainerName(containerType)
	blobName := s.generateBlobName(filename)

	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.accountName,
		containerName,
		blobName,
	)
}

// GetBlobURL สร้าง URL สำหรับ blob ที่มีอยู่แล้ว
func (s *StorageClient) GetBlobURL(blobName, containerType string) string {
	containerName := s.getContainerName(containerType)

	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.accountName,
		containerName,
		blobName,
	)
}

// TestConnection ทดสอบการเชื่อมต่อกับ Azure Blob Storage
func (s *StorageClient) TestConnection(ctx context.Context) error {
	// ลองดึงรายการ containers เพื่อทดสอบการเชื่อมต่อ
	pager := s.client.NewListContainersPager(nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Azure Blob Storage: %w", err)
	}
	return nil
}

// getContainerName เลือก container name ตาม type
func (s *StorageClient) getContainerName(containerType string) string {
	switch containerType {
	case "learning-path":
		return s.containerLearningPath
	case "profile":
		return s.containerProfile
	default:
		return s.containerLearningPath
	}
}

// generateBlobName สร้างชื่อ blob ที่ unique
func (s *StorageClient) generateBlobName(filename string) string {
	// ดึง extension จากชื่อไฟล์เดิม
	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			break
		}
	}

	// สร้างชื่อใหม่ด้วย UUID
	return uuid.New().String() + ext
}

// extractAccountName ดึงชื่อ storage account จาก connection string
func extractAccountName(connString string) string {
	// Parse connection string to extract AccountName
	// Format: "...;AccountName=xxx;..."
	start := 0
	for i := 0; i < len(connString)-12; i++ {
		if connString[i:i+12] == "AccountName=" {
			start = i + 12
			break
		}
	}

	if start == 0 {
		return ""
	}

	end := start
	for end < len(connString) && connString[end] != ';' {
		end++
	}

	return connString[start:end]
}
