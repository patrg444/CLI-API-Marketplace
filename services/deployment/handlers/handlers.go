package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/api-direct/services/deployment/k8s"
)

// DeployRequest represents a deployment request
type DeployRequest struct {
	APIId        string            `json:"api_id" binding:"required"`
	Version      string            `json:"version" binding:"required"`
	Runtime      string            `json:"runtime" binding:"required"`
	CodeURL      string            `json:"code_url" binding:"required"`
	Environment  map[string]string `json:"environment"`
	Replicas     int32             `json:"replicas"`
	Resources    ResourceSpec      `json:"resources"`
}

// ResourceSpec defines resource requirements
type ResourceSpec struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

// DeployAPI handles API deployment requests
func DeployAPI(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		var req DeployRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Override API ID from path
		req.APIId = apiId

		// Get user info from context
		userId, _ := c.Get("user_id")

		// Build container image based on runtime
		image := buildContainerImage(req.Runtime, req.Version)

		// Default replicas if not specified
		if req.Replicas == 0 {
			req.Replicas = 1
		}

		// Create deployment configuration
		config := k8s.DeploymentConfig{
			APIId:       req.APIId,
			Version:     req.Version,
			Image:       image,
			Port:        8080, // Standard port for API containers
			Replicas:    req.Replicas,
			Environment: req.Environment,
			ResourceLimits: k8s.ResourceRequirements{
				CPURequest:    req.Resources.CPURequest,
				CPULimit:      req.Resources.CPULimit,
				MemoryRequest: req.Resources.MemoryRequest,
				MemoryLimit:   req.Resources.MemoryLimit,
			},
		}

		// Add default environment variables
		if config.Environment == nil {
			config.Environment = make(map[string]string)
		}
		config.Environment["API_ID"] = req.APIId
		config.Environment["API_VERSION"] = req.Version
		config.Environment["USER_ID"] = userId.(string)

		// Deploy to Kubernetes
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
		defer cancel()

		if err := client.DeployAPI(ctx, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to deploy API: %v", err)})
			return
		}

		// Generate API endpoint URL
		endpoint := fmt.Sprintf("https://api.api-direct.io/apis/%s", req.APIId)

		c.JSON(http.StatusOK, gin.H{
			"message":  "API deployed successfully",
			"api_id":   req.APIId,
			"version":  req.Version,
			"endpoint": endpoint,
			"status":   "deploying",
		})
	}
}

// GetDeploymentStatus returns the current deployment status
func GetDeploymentStatus(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		status, err := client.GetDeploymentStatus(c.Request.Context(), apiId)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
			return
		}

		// Determine overall status
		var overallStatus string
		if status.ReadyReplicas == status.Replicas {
			overallStatus = "running"
		} else if status.ReadyReplicas > 0 {
			overallStatus = "partial"
		} else {
			overallStatus = "pending"
		}

		c.JSON(http.StatusOK, gin.H{
			"api_id":          apiId,
			"status":          overallStatus,
			"replicas":        status.Replicas,
			"ready_replicas":  status.ReadyReplicas,
			"updated_replicas": status.UpdatedReplicas,
			"conditions":      status.Conditions,
		})
	}
}

// UndeployAPI removes a deployed API
func UndeployAPI(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		// TODO: Verify user owns this API

		if err := client.DeleteDeployment(c.Request.Context(), apiId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to undeploy API: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "API undeployed successfully",
			"api_id":  apiId,
		})
	}
}

// ScaleDeployment adjusts the number of replicas
func ScaleDeployment(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		var req struct {
			Replicas int32 `json:"replicas" binding:"required,min=0,max=10"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// TODO: Verify user owns this API

		if err := client.ScaleDeployment(c.Request.Context(), apiId, req.Replicas); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scale deployment: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Deployment scaled successfully",
			"api_id":   apiId,
			"replicas": req.Replicas,
		})
	}
}

// GetEnvironmentVars returns environment variables for a deployment
func GetEnvironmentVars(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		// TODO: Implement environment variable retrieval from deployment

		c.JSON(http.StatusOK, gin.H{
			"api_id": apiId,
			"environment": map[string]string{
				"API_ID":      apiId,
				"LOG_LEVEL":   "info",
				"MAX_WORKERS": "4",
			},
		})
	}
}

// UpdateEnvironmentVars updates environment variables for a deployment
func UpdateEnvironmentVars(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		var env map[string]string
		if err := c.ShouldBindJSON(&env); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid environment variables"})
			return
		}

		// TODO: Implement environment variable update
		// This would require updating the deployment and restarting pods

		c.JSON(http.StatusOK, gin.H{
			"message":     "Environment variables updated",
			"api_id":      apiId,
			"environment": env,
		})
	}
}

// StreamLogs streams logs from a deployment
func StreamLogs(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		// Set up SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// TODO: Implement actual log streaming from Kubernetes pods
		// For now, send mock logs

		// Create a channel for sending logs
		logChan := make(chan string)
		done := make(chan bool)

		// Mock log generator
		go func() {
			for i := 0; i < 10; i++ {
				select {
				case <-done:
					return
				default:
					logChan <- fmt.Sprintf("[%s] Log line %d: API request processed", time.Now().Format(time.RFC3339), i)
					time.Sleep(1 * time.Second)
				}
			}
			close(logChan)
		}()

		// Stream logs to client
		c.Stream(func(w io.Writer) bool {
			select {
			case log, ok := <-logChan:
				if !ok {
					return false
				}
				c.SSEvent("log", log)
				return true
			case <-c.Request.Context().Done():
				close(done)
				return false
			}
		})
	}
}

// GetMetrics returns metrics for a deployment
func GetMetrics(client *k8s.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiId := c.Param("apiId")
		if apiId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API ID is required"})
			return
		}

		// TODO: Implement actual metrics collection from Kubernetes/Prometheus
		// For now, return mock metrics

		metrics := gin.H{
			"api_id": apiId,
			"period": "1h",
			"metrics": gin.H{
				"requests_total":      1234,
				"requests_per_second": 2.5,
				"error_rate":          0.02,
				"avg_response_time":   145, // milliseconds
				"cpu_usage":           0.35,
				"memory_usage":        0.67,
			},
			"timestamp": time.Now().UTC(),
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// Helper functions

func buildContainerImage(runtime, version string) string {
	// In production, this would return the actual container image URL
	// from ECR based on the runtime and version
	switch runtime {
	case "python3.9":
		return fmt.Sprintf("api-direct/python:3.9-%s", version)
	case "nodejs18":
		return fmt.Sprintf("api-direct/nodejs:18-%s", version)
	default:
		return fmt.Sprintf("api-direct/base:%s", version)
	}
}
