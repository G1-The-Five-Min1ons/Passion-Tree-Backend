package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/google/uuid"
)

const (
	MaxFileSizeBytes = 5 * 1024 * 1024 // 5MB
	MaxFileSizeHuman = "5MB"

	ContentTypeJPEG = "image/jpeg"
	ContentTypeJPG  = "image/jpg"
	ContentTypePNG  = "image/png"
	ContentTypePDF  = "application/pdf"

	SASStartTimeBuffer = -1 * time.Minute // Buffer เวลาเผื่อนาฬิกา server ไม่ตรงกัน
)

type BlobService struct {
	client      *azblob.Client
	accountName string
	accountKey  string
	containers  map[string]string
}

func NewBlobService(client *azblob.Client, accName, accKey string, containerMap map[string]string) *BlobService {
	return &BlobService{
		client:      client,
		accountName: accName,
		accountKey:  accKey,
		containers:  containerMap,
	}
}

// GeneratePresignedURL สร้าง SAS URL สำหรับอัปโหลด
func (s *BlobService) GeneratePresignedURL(filename string, containerType string, expiresIn time.Duration) (string, string, error) {
	containerName, err := s.getContainerName(containerType)
	if err != nil {
		return "", "", err
	}
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
		StartTime:     time.Now().Add(SASStartTimeBuffer),
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

// ValidateUploadedFile ตรวจสอบไฟล์หลังอัปโหลด (Size, Content-Type)
func (s *BlobService) ValidateUploadedFile(ctx context.Context, blobURL string, containerType string) error {
	parts := strings.Split(blobURL, "/")
	if len(parts) < 1 {
		return fmt.Errorf("invalid blob URL")
	}
	blobName := parts[len(parts)-1]
	containerName, err := s.getContainerName(containerType)
	if err != nil {
		return err
	}

	blobClient := s.client.ServiceClient().NewContainerClient(containerName).NewBlobClient(blobName)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get blob properties: %w", err)
	}

	if props.ContentLength != nil && *props.ContentLength > MaxFileSizeBytes {
		_, _ = blobClient.Delete(ctx, nil)
		return fmt.Errorf("file size %d exceeds limit of %s", *props.ContentLength, MaxFileSizeHuman)
	}

	if props.ContentType != nil {
		ct := *props.ContentType
		isValidContentType := ct == ContentTypeJPEG || ct == ContentTypePNG || ct == ContentTypeJPG || ct == ContentTypePDF
		if !isValidContentType {
			_, _ = blobClient.Delete(ctx, nil)
			return fmt.Errorf("invalid content type: %s", ct)
		}
	}

	return nil
}

func (s *BlobService) ListBlobsOlderThan(ctx context.Context, containerType string, duration time.Duration) ([]string, error) {
	containerName, err := s.getContainerName(containerType)
	if err != nil {
		return nil, err
	}

	pager := s.client.NewListBlobsFlatPager(containerName, nil)

	var blobNames []string
	cutoffTime := time.Now().Add(-duration)

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, item := range resp.Segment.BlobItems {
			if item.Properties.CreationTime != nil && item.Properties.CreationTime.Before(cutoffTime) {
				blobNames = append(blobNames, *item.Name)
			}
		}
	}

	return blobNames, nil
}

func (s *BlobService) DeleteBlob(ctx context.Context, blobName string, containerType string) error {
	containerName, err := s.getContainerName(containerType)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteBlob(ctx, containerName, blobName, nil)
	return err
}

// getContainerName เลือก container name ตาม type
func (s *BlobService) getContainerName(containerType string) (string, error) {
	name, exists := s.containers[containerType]
	if !exists {
		return "", fmt.Errorf("invalid container type (folder): %s", containerType)
	}
	return name, nil
}

// generateBlobName สร้างชื่อ blob ที่ unique
func (s *BlobService) generateBlobName(filename string) string {
	cleanName := filepath.Base(filename)

	ext := filepath.Ext(cleanName)
	return uuid.New().String() + ext
}
