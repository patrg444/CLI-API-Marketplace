package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/storage/s3client"
)

// UploadCode handles code package uploads
func UploadCode(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		// Get user info from context (set by auth middleware)
		userId, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")

		// Parse multipart form
		file, header, err := c.Request.FormFile("code")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get uploaded file"})
			return
		}
		defer file.Close()

		// Calculate checksum
		hasher := sha256.New()
		tee := io.TeeReader(file, hasher)

		// Read runtime from form
		runtime := c.PostForm("runtime")
		if runtime == "" {
			runtime = "python3.9" // Default
		}

		// Create metadata
		metadata := s3client.CodeMetadata{
			APIId:      apiId,
			UploadedAt: time.Now(),
			Size:       header.Size,
			Runtime:    runtime,
			Author:     userEmail.(string),
			Tags: map[string]string{
				"user_id": userId.(string),
			},
		}

		// Upload to S3
		version, err := client.UploadCode(c.Request.Context(), apiId, tee, metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload code: %v", err)})
			return
		}

		// Update checksum
		metadata.Checksum = hex.EncodeToString(hasher.Sum(nil))
		metadata.Version = version

		c.JSON(http.StatusOK, gin.H{
			"message":  "Code uploaded successfully",
			"version":  version,
			"checksum": metadata.Checksum,
			"size":     metadata.Size,
		})
	}
}

// DownloadCode handles code package downloads
func DownloadCode(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		version := c.Param("version")

		if apiId == "" || version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID and version are required"})
			return
		}

		// Download from S3
		reader, err := client.DownloadCode(c.Request.Context(), apiId, version)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Code version not found"})
			return
		}
		defer reader.Close()

		// Set headers
		c.Header("Content-Type", "application/gzip")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.tar.gz", apiId, version))

		// Stream the file
		_, err = io.Copy(c.Writer, reader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download code"})
			return
		}
	}
}

// ListVersions lists all versions for an API
func ListVersions(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		versions, err := client.ListVersions(c.Request.Context(), apiId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list versions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"api_id":   apiId,
			"versions": versions,
			"count":    len(versions),
		})
	}
}

// DeleteVersion deletes a specific version
func DeleteVersion(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		version := c.Param("version")

		if apiId == "" || version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID and version are required"})
			return
		}

		// TODO: Check if user owns this API

		err := client.DeleteVersion(c.Request.Context(), apiId, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete version"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Version deleted successfully",
			"api_id":  apiId,
			"version": version,
		})
	}
}

// GetMetadata retrieves metadata for a version
func GetMetadata(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		version := c.Param("version")

		if apiId == "" || version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID and version are required"})
			return
		}

		// TODO: Implement metadata retrieval
		c.JSON(http.StatusOK, gin.H{
			"api_id":  apiId,
			"version": version,
			"metadata": gin.H{
				"runtime":    "python3.9",
				"uploaded":   time.Now(),
				"size":       1024,
				"checksum":   "placeholder",
			},
		})
	}
}

// UpdateMetadata updates metadata for a version
func UpdateMetadata(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		version := c.Param("version")

		if apiId == "" || version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID and version are required"})
			return
		}

		var metadata map[string]interface{}
		if err := c.ShouldBindJSON(&metadata); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata format"})
			return
		}

		// TODO: Implement metadata update

		c.JSON(http.StatusOK, gin.H{
			"message":  "Metadata updated successfully",
			"api_id":   apiId,
			"version":  version,
			"metadata": metadata,
		})
	}
}

// GetPresignedUploadURL generates a presigned URL for direct upload
func GetPresignedUploadURL(client *s3client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		version := c.Query("version")
		if version == "" {
			version = fmt.Sprintf("v%d", time.Now().Unix())
		}

		url, err := client.GetPresignedUploadURL(c.Request.Context(), apiId, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"upload_url": url,
			"version":    version,
			"expires_in": "15 minutes",
			"method":     "PUT",
		})
	}
}
