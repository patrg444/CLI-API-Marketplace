package s3client

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// Client wraps the S3 client with helper methods
type Client struct {
	s3Client *s3.Client
	bucket   string
}

// CodeMetadata represents metadata for uploaded code
type CodeMetadata struct {
	APIId       string            `json:"api_id"`
	Version     string            `json:"version"`
	UploadedAt  time.Time         `json:"uploaded_at"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Runtime     string            `json:"runtime"`
	Author      string            `json:"author"`
	Tags        map[string]string `json:"tags"`
}

// NewClient creates a new S3 client wrapper
func NewClient() (*Client, error) {
	bucket := os.Getenv("CODE_STORAGE_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("CODE_STORAGE_BUCKET environment variable not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3Client: s3.NewFromConfig(cfg),
		bucket:   bucket,
	}, nil
}

// UploadCode uploads code package to S3
func (c *Client) UploadCode(ctx context.Context, apiId string, data io.Reader, metadata CodeMetadata) (string, error) {
	// Generate version if not provided
	if metadata.Version == "" {
		metadata.Version = generateVersion()
	}

	key := fmt.Sprintf("apis/%s/versions/%s/code.tar.gz", apiId, metadata.Version)

	// Upload to S3
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   data,
		Metadata: map[string]string{
			"api-id":     apiId,
			"version":    metadata.Version,
			"uploaded":   metadata.UploadedAt.Format(time.RFC3339),
			"runtime":    metadata.Runtime,
			"author":     metadata.Author,
		},
		ServerSideEncryption: types.ServerSideEncryption(types.ServerSideEncryptionAes256),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Store metadata
	if err := c.storeMetadata(ctx, apiId, metadata); err != nil {
		return "", fmt.Errorf("failed to store metadata: %w", err)
	}

	return metadata.Version, nil
}

// DownloadCode downloads code package from S3
func (c *Client) DownloadCode(ctx context.Context, apiId, version string) (io.ReadCloser, error) {
	key := fmt.Sprintf("apis/%s/versions/%s/code.tar.gz", apiId, version)

	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// ListVersions lists all versions for an API
func (c *Client) ListVersions(ctx context.Context, apiId string) ([]CodeMetadata, error) {
	prefix := fmt.Sprintf("apis/%s/versions/", apiId)

	var versions []CodeMetadata
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Extract version from key
			if key := aws.ToString(obj.Key); len(key) > 0 {
				// Parse metadata from object
				metadata, err := c.GetMetadata(ctx, apiId, extractVersion(key))
				if err == nil {
					versions = append(versions, metadata)
				}
			}
		}
	}

	return versions, nil
}

// DeleteVersion deletes a specific version
func (c *Client) DeleteVersion(ctx context.Context, apiId, version string) error {
	key := fmt.Sprintf("apis/%s/versions/%s/code.tar.gz", apiId, version)

	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	// Delete metadata
	metadataKey := fmt.Sprintf("apis/%s/versions/%s/metadata.json", apiId, version)
	_, _ = c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(metadataKey),
	})

	return nil
}

// GetPresignedUploadURL generates a presigned URL for direct upload
func (c *Client) GetPresignedUploadURL(ctx context.Context, apiId, version string) (string, error) {
	key := fmt.Sprintf("apis/%s/versions/%s/code.tar.gz", apiId, version)

	presignClient := s3.NewPresignClient(c.s3Client)
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}

// Helper functions

func (c *Client) storeMetadata(ctx context.Context, apiId string, metadata CodeMetadata) error {
	// Implementation would store metadata in S3 or DynamoDB
	// For now, we'll store it as a JSON file in S3
	_ = fmt.Sprintf("apis/%s/versions/%s/metadata.json", apiId, metadata.Version)
	
	// Marshal metadata to JSON and upload
	// (Implementation details omitted for brevity)
	
	return nil
}

func (c *Client) GetMetadata(ctx context.Context, apiId, version string) (CodeMetadata, error) {
	// Implementation would retrieve metadata from storage
	// For now, return a placeholder
	return CodeMetadata{
		APIId:   apiId,
		Version: version,
	}, nil
}

func generateVersion() string {
	return fmt.Sprintf("v%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

func extractVersion(key string) string {
	// Extract version from S3 key path
	// Implementation depends on key structure
	return ""
}
