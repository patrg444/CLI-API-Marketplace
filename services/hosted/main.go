package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BuildRequest struct {
	ImageTag string `json:"image_tag"`
}

type BuildResponse struct {
	ImageTag string `json:"image_tag"`
	BuildID  string `json:"build_id"`
	Status   string `json:"status"`
}

type DeployRequest struct {
	APIName       string                 `json:"api_name"`
	ImageTag      string                 `json:"image_tag"`
	Runtime       string                 `json:"runtime"`
	Endpoints     []Endpoint             `json:"endpoints"`
	Environment   map[string]string      `json:"environment"`
	ResourceLimits map[string]string     `json:"resource_limits"`
	AutoScaling   map[string]interface{} `json:"auto_scaling"`
}

type Endpoint struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
}

type DeployResponse struct {
	Endpoint     string `json:"endpoint"`
	DeploymentID string `json:"deployment_id"`
	Status       string `json:"status"`
	DatabaseURL  string `json:"database_url"`
	Subdomain    string `json:"subdomain"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

// In-memory store for demo purposes
var deployments = make(map[string]*DeployResponse)
var builds = make(map[string]*BuildResponse)

func main() {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "api-direct-hosted"})
	})

	// Hosted API endpoints
	v1 := r.Group("/hosted/v1")
	{
		v1.POST("/build", handleBuild)
		v1.POST("/deploy", handleDeploy)
		v1.GET("/status/:deployment_id", handleStatus)
		v1.GET("/deployments", handleListDeployments)
	}

	log.Println("🚀 API-Direct Hosted Service starting on :8084")
	log.Fatal(http.ListenAndServe(":8084", r))
}

func handleBuild(c *gin.Context) {
	var req BuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	buildID := uuid.New().String()
	
	// Simulate build process
	log.Printf("🐳 Building container image: %s", req.ImageTag)
	
	response := &BuildResponse{
		ImageTag: req.ImageTag,
		BuildID:  buildID,
		Status:   "success",
	}
	
	builds[buildID] = response
	
	// Simulate build time
	time.Sleep(2 * time.Second)
	
	c.JSON(200, response)
}

func handleDeploy(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	deploymentID := uuid.New().String()
	
	// Generate subdomain
	subdomain := fmt.Sprintf("%s-%s", req.APIName, deploymentID[:8])
	endpoint := fmt.Sprintf("https://%s.api-direct.io", subdomain)
	
	// Generate database URL
	databaseURL := fmt.Sprintf("postgresql://api_%s:%s@postgres-hosted:5432/api_%s_%s", 
		deploymentID[:8], "generated_password", deploymentID[:8], req.APIName)
	
	log.Printf("☁️  Deploying %s to hosted infrastructure", req.APIName)
	log.Printf("📍 Endpoint: %s", endpoint)
	log.Printf("🗄️  Database: %s", databaseURL)
	
	response := &DeployResponse{
		Endpoint:     endpoint,
		DeploymentID: deploymentID,
		Status:       "deploying",
		DatabaseURL:  databaseURL,
		Subdomain:    subdomain,
	}
	
	deployments[deploymentID] = response
	
	// Simulate deployment process
	go func() {
		time.Sleep(5 * time.Second)
		deployments[deploymentID].Status = "running"
		log.Printf("✅ Deployment %s is now running", deploymentID)
	}()
	
	c.JSON(200, response)
}

func handleStatus(c *gin.Context) {
	deploymentID := c.Param("deployment_id")
	
	deployment, exists := deployments[deploymentID]
	if !exists {
		c.JSON(404, gin.H{"error": "Deployment not found"})
		return
	}
	
	c.JSON(200, StatusResponse{Status: deployment.Status})
}

func handleListDeployments(c *gin.Context) {
	var deploymentList []DeployResponse
	for _, deployment := range deployments {
		deploymentList = append(deploymentList, *deployment)
	}
	
	c.JSON(200, gin.H{
		"deployments": deploymentList,
		"count":       len(deploymentList),
	})
}