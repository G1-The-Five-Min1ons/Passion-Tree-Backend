package database

import (
	"context"
	"fmt"
	"passiontree/internal/config"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/google/uuid"
)

type StorageClient struct {
	client                *azblob.Client
	accountName           string
	accountKey            string
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
	accountKey := extractAccountKey(cfg.AzureStorageConnString)

	return &StorageClient{
		client:                client,
		accountName:           accountName,
		accountKey:            accountKey,
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

func (s *StorageClient) GeneratePresignedURL(filename string, containerType string, expiresIn time.Duration) (string, string, error) {
	containerName := s.getContainerName(containerType)
	blobName := s.generateBlobName(filename)

	cred, err := azblob.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		return "", "", fmt.Errorf("invalid credentials for SAS: %w", err)
	}

	sasPermissions := sas.BlobPermissions{
		Create: true,
		Write:  true,
		Add:    true,
		Read:   true,
	}

	expiry := time.Now().Add(expiresIn)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().Add(-1 * time.Minute),
		ExpiryTime:    expiry,
		Permissions:   sasPermissions.String(),
		ContainerName: containerName,
		BlobName:      blobName,
	}.SignWithSharedKey(cred)

	if err != nil {
		return "", "", fmt.Errorf("failed to sign SAS: %w", err)
	}

	uploadURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		s.accountName,
		containerName,
		blobName,
		sasQueryParams.Encode(),
	)

	publicURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.accountName,
		containerName,
		blobName,
	)

	return uploadURL, publicURL, nil
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

func (s *StorageClient) ValidateUploadedFile(ctx context.Context, blobURL string, containerType string) error {
	parts := strings.Split(blobURL, "/")
	if len(parts) < 1 {
		return fmt.Errorf("invalid blob URL")
	}
	blobName := parts[len(parts)-1]
	containerName := s.getContainerName(containerType)
	
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", s.accountName)
	cred, err := azblob.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		return err
	}
	
	blobClient, err := blob.NewClientWithSharedKeyCredential(fmt.Sprintf("%s%s/%s", serviceURL, containerName, blobName), cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create blob client: %w", err)
	}

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get blob properties (file might not exist): %w", err)
	}

	const maxFileSize = 5 * 1024 * 1024

	if props.ContentLength != nil && *props.ContentLength > maxFileSize {
		_, _ = blobClient.Delete(ctx, nil) 
		return fmt.Errorf("file size %d exceeds limit of 5MB", *props.ContentLength)
	}

	if props.ContentType != nil {
		ct := *props.ContentType
		if ct != "image/jpeg" && ct != "image/png" && ct != "image/jpg" {
			_, _ = blobClient.Delete(ctx, nil)
			return fmt.Errorf("invalid content type: %s", ct)
		}
	}

	return nil
}