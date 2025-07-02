package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/api-direct/cli/pkg/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	docsFormat       string
	docsOutput       string
	docsIncludeTests bool
	docsTheme        string
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate API documentation",
	Long: `Generate comprehensive API documentation from your code and OpenAPI specifications.

This command can:
- Auto-generate OpenAPI/Swagger documentation
- Create markdown documentation
- Generate interactive API documentation
- Export Postman collections`,
}

// generateDocsCmd generates documentation
var generateDocsCmd = &cobra.Command{
	Use:   "generate [api-name]",
	Short: "Generate API documentation",
	Long: `Generate API documentation in various formats including OpenAPI, Markdown, and HTML.

Examples:
  apidirect docs generate my-api
  apidirect docs generate my-api --format openapi
  apidirect docs generate my-api --format markdown --output docs/
  apidirect docs generate my-api --format html --theme slate`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerateDocs,
}

// previewDocsCmd previews documentation
var previewDocsCmd = &cobra.Command{
	Use:   "preview [api-name]",
	Short: "Preview API documentation locally",
	Long: `Start a local server to preview your API documentation.

Examples:
  apidirect docs preview my-api
  apidirect docs preview my-api --port 8080`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPreviewDocs,
}

// publishDocsCmd publishes documentation
var publishDocsCmd = &cobra.Command{
	Use:   "publish [api-name]",
	Short: "Publish API documentation",
	Long: `Publish your API documentation to make it accessible online.

Examples:
  apidirect docs publish my-api
  apidirect docs publish my-api --custom-domain docs.myapi.com`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPublishDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.AddCommand(generateDocsCmd)
	docsCmd.AddCommand(previewDocsCmd)
	docsCmd.AddCommand(publishDocsCmd)
	
	// Generate flags
	generateDocsCmd.Flags().StringVarP(&docsFormat, "format", "f", "openapi", "Documentation format (openapi, markdown, html, postman)")
	generateDocsCmd.Flags().StringVarP(&docsOutput, "output", "o", "./docs", "Output directory")
	generateDocsCmd.Flags().BoolVar(&docsIncludeTests, "include-tests", false, "Include test examples in documentation")
	generateDocsCmd.Flags().StringVar(&docsTheme, "theme", "default", "Documentation theme (for HTML format)")
	
	// Preview flags
	previewDocsCmd.Flags().IntP("port", "p", 8080, "Port to run preview server")
	
	// Publish flags
	publishDocsCmd.Flags().String("custom-domain", "", "Custom domain for documentation")
	publishDocsCmd.Flags().Bool("private", false, "Make documentation private (requires authentication)")
}

func runGenerateDocs(cmd *cobra.Command, args []string) error {
	apiName := getCurrentAPIName(args)
	
	_, err := config.Load()
	if err != nil {
		return err
	}
	
	fmt.Printf("🔍 Analyzing API: %s\n", apiName)
	
	// Read manifest
	manifestPath := filepath.Join(".", "apidirect.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestPath = filepath.Join(".", "apidirect.yml")
	}
	
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var manifest map[string]interface{}
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	// Analyze API structure
	fmt.Println("📊 Detecting API endpoints...")
	endpoints, err := detectAPIEndpoints(manifest)
	if err != nil {
		return fmt.Errorf("failed to detect endpoints: %w", err)
	}
	
	fmt.Printf("Found %d endpoints\n", len(endpoints))
	
	// Generate documentation based on format
	switch docsFormat {
	case "openapi":
		if err := generateOpenAPIDoc(apiName, manifest, endpoints); err != nil {
			return err
		}
	case "markdown":
		if err := generateMarkdownDoc(apiName, manifest, endpoints); err != nil {
			return err
		}
	case "html":
		if err := generateHTMLDoc(apiName, manifest, endpoints); err != nil {
			return err
		}
	case "postman":
		if err := generatePostmanCollection(apiName, manifest, endpoints); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", docsFormat)
	}
	
	color.Green("✅ Documentation generated successfully!")
	fmt.Printf("Output: %s\n", docsOutput)
	
	return nil
}

func runPreviewDocs(cmd *cobra.Command, args []string) error {
	apiName := getCurrentAPIName(args)
	port, _ := cmd.Flags().GetInt("port")
	
	// Check if documentation exists
	docsPath := filepath.Join(docsOutput, "index.html")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return fmt.Errorf("documentation not found. Run 'apidirect docs generate' first")
	}
	
	fmt.Printf("🌐 Starting documentation preview server...\n")
	fmt.Printf("API: %s\n", apiName)
	fmt.Printf("URL: http://localhost:%d\n", port)
	fmt.Println("\nPress Ctrl+C to stop")
	
	// In a real implementation, start an HTTP server
	// For now, we'll simulate it
	select {}
}

func runPublishDocs(cmd *cobra.Command, args []string) error {
	apiName := getCurrentAPIName(args)
	customDomain, _ := cmd.Flags().GetString("custom-domain")
	isPrivate, _ := cmd.Flags().GetBool("private")
	
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	
	fmt.Printf("📤 Publishing documentation for: %s\n", apiName)
	
	// Package documentation
	fmt.Println("📦 Packaging documentation...")
	
	// Upload to platform
	publishData := struct {
		APIName      string `json:"api_name"`
		CustomDomain string `json:"custom_domain,omitempty"`
		Private      bool   `json:"private"`
	}{
		APIName:      apiName,
		CustomDomain: customDomain,
		Private:      isPrivate,
	}
	
	data, _ := json.Marshal(publishData)
	url := fmt.Sprintf("%s/api/v1/apis/%s/docs/publish", cfg.APIEndpoint, apiName)
	
	resp, err := makeAuthenticatedRequest("POST", url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return handleErrorResponse(resp)
	}
	
	var result struct {
		URL          string `json:"url"`
		CustomDomain string `json:"custom_domain,omitempty"`
		Status       string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	color.Green("✅ Documentation published successfully!")
	fmt.Printf("\n📚 Documentation URL: %s\n", color.BlueString(result.URL))
	if result.CustomDomain != "" {
		fmt.Printf("🌐 Custom domain: %s\n", color.BlueString(result.CustomDomain))
	}
	
	return nil
}

func detectAPIEndpoints(manifest map[string]interface{}) ([]map[string]interface{}, error) {
	// This is a simplified version - in reality, we'd analyze the code
	endpoints := []map[string]interface{}{
		{
			"path":        "/api/v1/users",
			"method":      "GET",
			"description": "List all users",
			"parameters": []map[string]interface{}{
				{"name": "page", "type": "integer", "required": false},
				{"name": "limit", "type": "integer", "required": false},
			},
		},
		{
			"path":        "/api/v1/users/{id}",
			"method":      "GET",
			"description": "Get user by ID",
			"parameters": []map[string]interface{}{
				{"name": "id", "type": "string", "required": true},
			},
		},
	}
	
	return endpoints, nil
}

func generateOpenAPIDoc(apiName string, manifest map[string]interface{}, endpoints []map[string]interface{}) error {
	openapi := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       apiName,
			"version":     "1.0.0",
			"description": manifest["description"],
		},
		"servers": []map[string]interface{}{
			{
				"url":         "https://api.apidirect.io/" + apiName,
				"description": "Production server",
			},
		},
		"paths": make(map[string]interface{}),
	}
	
	// Convert endpoints to OpenAPI paths
	paths := openapi["paths"].(map[string]interface{})
	for _, endpoint := range endpoints {
		path := endpoint["path"].(string)
		method := strings.ToLower(endpoint["method"].(string))
		
		if paths[path] == nil {
			paths[path] = make(map[string]interface{})
		}
		
		pathItem := paths[path].(map[string]interface{})
		pathItem[method] = map[string]interface{}{
			"summary":     endpoint["description"],
			"operationId": fmt.Sprintf("%s_%s", method, strings.ReplaceAll(path, "/", "_")),
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
				},
			},
		}
		
		// Add parameters
		if params, ok := endpoint["parameters"].([]map[string]interface{}); ok && len(params) > 0 {
			pathItem[method].(map[string]interface{})["parameters"] = params
		}
	}
	
	// Write OpenAPI spec
	outputPath := filepath.Join(docsOutput, "openapi.yaml")
	if err := os.MkdirAll(docsOutput, 0755); err != nil {
		return err
	}
	
	data, err := yaml.Marshal(openapi)
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputPath, data, 0644)
}

func generateMarkdownDoc(apiName string, manifest map[string]interface{}, endpoints []map[string]interface{}) error {
	var buf bytes.Buffer
	
	// Header
	fmt.Fprintf(&buf, "# %s API Documentation\n\n", apiName)
	if desc, ok := manifest["description"].(string); ok {
		fmt.Fprintf(&buf, "%s\n\n", desc)
	}
	
	// Table of contents
	fmt.Fprintf(&buf, "## Table of Contents\n\n")
	for i, endpoint := range endpoints {
		fmt.Fprintf(&buf, "%d. [%s %s](#endpoint-%d)\n", 
			i+1, endpoint["method"], endpoint["path"], i+1)
	}
	fmt.Fprintf(&buf, "\n")
	
	// Endpoints
	fmt.Fprintf(&buf, "## Endpoints\n\n")
	for i, endpoint := range endpoints {
		fmt.Fprintf(&buf, "### <a name=\"endpoint-%d\"></a>%s %s\n\n", 
			i+1, endpoint["method"], endpoint["path"])
		fmt.Fprintf(&buf, "%s\n\n", endpoint["description"])
		
		// Parameters
		if params, ok := endpoint["parameters"].([]map[string]interface{}); ok && len(params) > 0 {
			fmt.Fprintf(&buf, "**Parameters:**\n\n")
			fmt.Fprintf(&buf, "| Name | Type | Required | Description |\n")
			fmt.Fprintf(&buf, "|------|------|----------|-------------|\n")
			for _, param := range params {
				fmt.Fprintf(&buf, "| %s | %s | %v | %s |\n",
					param["name"], param["type"], param["required"], 
					param["description"])
			}
			fmt.Fprintf(&buf, "\n")
		}
		
		// Example
		fmt.Fprintf(&buf, "**Example Request:**\n\n```bash\n")
		fmt.Fprintf(&buf, "curl -X %s \\\n", endpoint["method"])
		fmt.Fprintf(&buf, "  https://api.apidirect.io/%s%s \\\n", apiName, endpoint["path"])
		fmt.Fprintf(&buf, "  -H 'X-API-Key: YOUR_API_KEY'\n")
		fmt.Fprintf(&buf, "```\n\n")
	}
	
	// Write markdown file
	outputPath := filepath.Join(docsOutput, "API_DOCUMENTATION.md")
	if err := os.MkdirAll(docsOutput, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func generateHTMLDoc(apiName string, manifest map[string]interface{}, endpoints []map[string]interface{}) error {
	// First generate OpenAPI spec
	if err := generateOpenAPIDoc(apiName, manifest, endpoints); err != nil {
		return err
	}
	
	// Generate HTML using Swagger UI
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s API Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.15.5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        #swagger-ui { padding: 20px; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "./openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`, apiName)
	
	// Write HTML file
	outputPath := filepath.Join(docsOutput, "index.html")
	return os.WriteFile(outputPath, []byte(html), 0644)
}

func generatePostmanCollection(apiName string, manifest map[string]interface{}, endpoints []map[string]interface{}) error {
	collection := map[string]interface{}{
		"info": map[string]interface{}{
			"name":        apiName + " API",
			"description": manifest["description"],
			"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item": []map[string]interface{}{},
		"variable": []map[string]interface{}{
			{
				"key":   "baseUrl",
				"value": "https://api.apidirect.io/" + apiName,
				"type":  "string",
			},
			{
				"key":   "apiKey",
				"value": "YOUR_API_KEY",
				"type":  "string",
			},
		},
	}
	
	// Convert endpoints to Postman items
	items := []map[string]interface{}{}
	for _, endpoint := range endpoints {
		item := map[string]interface{}{
			"name": fmt.Sprintf("%s %s", endpoint["method"], endpoint["path"]),
			"request": map[string]interface{}{
				"method": endpoint["method"],
				"header": []map[string]interface{}{
					{
						"key":   "X-API-Key",
						"value": "{{apiKey}}",
						"type":  "text",
					},
				},
				"url": map[string]interface{}{
					"raw":  "{{baseUrl}}" + endpoint["path"].(string),
					"host":  []string{"{{baseUrl}}"},
					"path":  strings.Split(strings.TrimPrefix(endpoint["path"].(string), "/"), "/"),
				},
			},
		}
		
		items = append(items, item)
	}
	collection["item"] = items
	
	// Write Postman collection
	outputPath := filepath.Join(docsOutput, apiName+"_postman_collection.json")
	if err := os.MkdirAll(docsOutput, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputPath, data, 0644)
}

func getCurrentAPIName(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	
	// Try to read from manifest
	if manifest, err := readManifest("."); err == nil {
		if name, ok := manifest["name"].(string); ok {
			return name
		}
	}
	
	// Default to current directory name
	dir, _ := os.Getwd()
	return filepath.Base(dir)
}