package detector

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectDetection holds the results of project analysis
type ProjectDetection struct {
	Language         string
	Runtime          string
	Framework        string
	StartCommand     string
	Port             int
	MainFile         string
	RequirementsFile string
	EnvFile          string
	Endpoints        []Endpoint
	Environment      EnvironmentVars
	HealthCheck      string
}

// Endpoint represents a detected API endpoint
type Endpoint struct {
	Method string
	Path   string
}

// EnvironmentVars holds required and optional environment variables
type EnvironmentVars struct {
	Required []string               `yaml:"required,omitempty"`
	Optional map[string]string      `yaml:"optional,omitempty"`
}

// AnalyzeProject scans a project directory and detects its configuration
func AnalyzeProject(projectPath string) (*ProjectDetection, error) {
	detection := &ProjectDetection{
		Port: 8080, // default port
		Environment: EnvironmentVars{
			Optional: make(map[string]string),
		},
	}
	
	// Detect language and framework
	if err := detectLanguageAndFramework(projectPath, detection); err != nil {
		return nil, err
	}
	
	// Find main entry file
	if err := findMainFile(projectPath, detection); err != nil {
		return nil, err
	}
	
	// Detect endpoints
	if detection.MainFile != "" {
		detectEndpoints(filepath.Join(projectPath, detection.MainFile), detection)
	}
	
	// Find environment configuration
	findEnvConfig(projectPath, detection)
	
	// Generate start command based on framework
	generateStartCommand(detection)
	
	// Detect port from code
	detectPort(projectPath, detection)
	
	// Set health check endpoint
	setHealthCheck(detection)
	
	return detection, nil
}

func detectLanguageAndFramework(projectPath string, d *ProjectDetection) error {
	// Check for Python
	if exists(filepath.Join(projectPath, "requirements.txt")) || 
	   exists(filepath.Join(projectPath, "setup.py")) ||
	   exists(filepath.Join(projectPath, "Pipfile")) {
		d.Language = "Python"
		d.Runtime = "python3.11"
		d.RequirementsFile = findFirst(projectPath, "requirements.txt", "Pipfile")
		
		// Detect Python framework
		detectPythonFramework(projectPath, d)
		return nil
	}
	
	// Check for Node.js
	if exists(filepath.Join(projectPath, "package.json")) {
		d.Language = "Node.js"
		d.Runtime = "node18"
		d.RequirementsFile = "package.json"
		
		// Detect Node framework
		detectNodeFramework(projectPath, d)
		return nil
	}
	
	// Check for Go
	if exists(filepath.Join(projectPath, "go.mod")) {
		d.Language = "Go"
		d.Runtime = "go1.21"
		d.RequirementsFile = "go.mod"
		detectGoFramework(projectPath, d)
		return nil
	}
	
	// Check for Ruby
	if exists(filepath.Join(projectPath, "Gemfile")) {
		d.Language = "Ruby"
		d.Runtime = "ruby3.0"
		d.RequirementsFile = "Gemfile"
		detectRubyFramework(projectPath, d)
		return nil
	}
	
	return fmt.Errorf("could not detect project language")
}

func detectPythonFramework(projectPath string, d *ProjectDetection) {
	mainFiles := []string{"main.py", "app.py", "application.py", "server.py", "api.py"}
	
	for _, file := range mainFiles {
		fullPath := filepath.Join(projectPath, file)
		if exists(fullPath) {
			content, _ := ioutil.ReadFile(fullPath)
			contentStr := string(content)
			
			// Check for FastAPI
			if strings.Contains(contentStr, "from fastapi import") || 
			   strings.Contains(contentStr, "FastAPI()") {
				d.Framework = "FastAPI"
				return
			}
			
			// Check for Flask
			if strings.Contains(contentStr, "from flask import") || 
			   strings.Contains(contentStr, "Flask(__name__)") {
				d.Framework = "Flask"
				return
			}
			
			// Check for Django
			if strings.Contains(contentStr, "django") ||
			   exists(filepath.Join(projectPath, "manage.py")) {
				d.Framework = "Django"
				return
			}
		}
	}
}

func detectNodeFramework(projectPath string, d *ProjectDetection) {
	// Read package.json to check dependencies
	packagePath := filepath.Join(projectPath, "package.json")
	if content, err := ioutil.ReadFile(packagePath); err == nil {
		contentStr := string(content)
		
		if strings.Contains(contentStr, "\"express\"") {
			d.Framework = "Express"
		} else if strings.Contains(contentStr, "\"fastify\"") {
			d.Framework = "Fastify"
		} else if strings.Contains(contentStr, "\"koa\"") {
			d.Framework = "Koa"
		} else if strings.Contains(contentStr, "\"next\"") {
			d.Framework = "Next.js"
		}
	}
}

func detectGoFramework(projectPath string, d *ProjectDetection) {
	// Check go.mod for framework imports
	goModPath := filepath.Join(projectPath, "go.mod")
	if content, err := ioutil.ReadFile(goModPath); err == nil {
		contentStr := string(content)
		
		if strings.Contains(contentStr, "github.com/gin-gonic/gin") {
			d.Framework = "Gin"
		} else if strings.Contains(contentStr, "github.com/labstack/echo") {
			d.Framework = "Echo"
		} else if strings.Contains(contentStr, "github.com/gofiber/fiber") {
			d.Framework = "Fiber"
		}
	}
}

func detectRubyFramework(projectPath string, d *ProjectDetection) {
	// Check Gemfile for framework
	gemfilePath := filepath.Join(projectPath, "Gemfile")
	if content, err := ioutil.ReadFile(gemfilePath); err == nil {
		contentStr := string(content)
		
		if strings.Contains(contentStr, "gem 'rails'") ||
		   strings.Contains(contentStr, "gem \"rails\"") {
			d.Framework = "Rails"
		} else if strings.Contains(contentStr, "gem 'sinatra'") ||
		          strings.Contains(contentStr, "gem \"sinatra\"") {
			d.Framework = "Sinatra"
		}
	}
}

func findMainFile(projectPath string, d *ProjectDetection) error {
	var candidates []string
	
	switch d.Language {
	case "Python":
		candidates = []string{"main.py", "app.py", "application.py", "server.py", "api.py", "wsgi.py"}
	case "Node.js":
		// Check package.json for main field
		if main := getPackageMain(projectPath); main != "" {
			d.MainFile = main
			return nil
		}
		candidates = []string{"server.js", "app.js", "index.js", "main.js", "api.js"}
	case "Go":
		candidates = []string{"main.go", "server.go", "app.go"}
	case "Ruby":
		candidates = []string{"app.rb", "server.rb", "application.rb", "config.ru"}
	}
	
	// Look for candidates
	for _, candidate := range candidates {
		if exists(filepath.Join(projectPath, candidate)) {
			d.MainFile = candidate
			return nil
		}
	}
	
	// Look in common subdirectories
	subdirs := []string{"src", "app", "lib", "api"}
	for _, subdir := range subdirs {
		for _, candidate := range candidates {
			path := filepath.Join(subdir, candidate)
			if exists(filepath.Join(projectPath, path)) {
				d.MainFile = path
				return nil
			}
		}
	}
	
	return fmt.Errorf("could not find main application file")
}

func detectEndpoints(filePath string, d *ProjectDetection) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	
	contentStr := string(content)
	endpoints := []Endpoint{}
	
	switch d.Framework {
	case "FastAPI":
		endpoints = detectFastAPIEndpoints(contentStr)
	case "Flask":
		endpoints = detectFlaskEndpoints(contentStr)
	case "Express":
		endpoints = detectExpressEndpoints(contentStr)
	case "Django":
		endpoints = detectDjangoEndpoints(filePath)
	}
	
	d.Endpoints = endpoints
}

func detectFastAPIEndpoints(content string) []Endpoint {
	endpoints := []Endpoint{}
	
	// Patterns for FastAPI decorators
	patterns := []string{
		`@app\.(get|post|put|delete|patch)\("([^"]+)"\)`,
		`@app\.(get|post|put|delete|patch)\('([^']+)'\)`,
		`@router\.(get|post|put|delete|patch)\("([^"]+)"\)`,
		`@router\.(get|post|put|delete|patch)\('([^']+)'\)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				endpoints = append(endpoints, Endpoint{
					Method: strings.ToUpper(match[1]),
					Path:   match[2],
				})
			}
		}
	}
	
	return endpoints
}

func detectFlaskEndpoints(content string) []Endpoint {
	endpoints := []Endpoint{}
	
	// Pattern for @app.route decorators
	routePattern := regexp.MustCompile(`@app\.route\(['"]([^'"]+)['"](?:.*methods=\[([^\]]+)\])?`)
	matches := routePattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		path := match[1]
		methods := "GET" // default
		
		if len(match) > 2 && match[2] != "" {
			// Parse methods
			methodsStr := strings.ReplaceAll(match[2], "'", "")
			methodsStr = strings.ReplaceAll(methodsStr, "\"", "")
			methodsList := strings.Split(methodsStr, ",")
			for _, method := range methodsList {
				endpoints = append(endpoints, Endpoint{
					Method: strings.TrimSpace(method),
					Path:   path,
				})
			}
		} else {
			endpoints = append(endpoints, Endpoint{
				Method: methods,
				Path:   path,
			})
		}
	}
	
	return endpoints
}

func detectExpressEndpoints(content string) []Endpoint {
	endpoints := []Endpoint{}
	
	// Patterns for Express routes
	patterns := []string{
		`app\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`,
		`router\.(get|post|put|delete|patch)\(['"]([^'"]+)['"]`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				endpoints = append(endpoints, Endpoint{
					Method: strings.ToUpper(match[1]),
					Path:   match[2],
				})
			}
		}
	}
	
	return endpoints
}

func detectDjangoEndpoints(filePath string) []Endpoint {
	// For Django, we'd need to look at urls.py files
	// This is a simplified version
	return []Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/admin"},
	}
}

func generateStartCommand(d *ProjectDetection) {
	switch d.Framework {
	case "FastAPI":
		d.StartCommand = fmt.Sprintf("uvicorn %s:app --host 0.0.0.0 --port %d", 
			strings.TrimSuffix(d.MainFile, ".py"), d.Port)
	case "Flask":
		if d.MainFile == "wsgi.py" {
			d.StartCommand = fmt.Sprintf("gunicorn wsgi:app --bind 0.0.0.0:%d", d.Port)
		} else {
			d.StartCommand = fmt.Sprintf("python %s", d.MainFile)
		}
	case "Django":
		d.StartCommand = fmt.Sprintf("gunicorn myproject.wsgi:application --bind 0.0.0.0:%d", d.Port)
	case "Express":
		d.StartCommand = fmt.Sprintf("node %s", d.MainFile)
	case "Rails":
		d.StartCommand = fmt.Sprintf("rails server -b 0.0.0.0 -p %d", d.Port)
	default:
		// Generic commands
		switch d.Language {
		case "Python":
			d.StartCommand = fmt.Sprintf("python %s", d.MainFile)
		case "Node.js":
			d.StartCommand = fmt.Sprintf("node %s", d.MainFile)
		case "Go":
			d.StartCommand = "./app"
		case "Ruby":
			d.StartCommand = fmt.Sprintf("ruby %s", d.MainFile)
		}
	}
}

func detectPort(projectPath string, d *ProjectDetection) {
	// Try to find port configuration in common places
	portPatterns := []string{
		`(?:PORT|port)\s*=\s*(\d+)`,
		`\.listen\((\d+)`,
		`--port[= ](\d+)`,
		`:(\d{4})`, // Common port pattern
	}
	
	// Check main file
	if d.MainFile != "" {
		if content, err := ioutil.ReadFile(filepath.Join(projectPath, d.MainFile)); err == nil {
			contentStr := string(content)
			for _, pattern := range portPatterns {
				re := regexp.MustCompile(pattern)
				if match := re.FindStringSubmatch(contentStr); len(match) > 1 {
					if port := parseInt(match[1]); port > 0 && port < 65536 {
						d.Port = port
						return
					}
				}
			}
		}
	}
	
	// Check environment files
	envFiles := []string{".env", ".env.example", ".env.sample"}
	for _, envFile := range envFiles {
		if content, err := ioutil.ReadFile(filepath.Join(projectPath, envFile)); err == nil {
			contentStr := string(content)
			re := regexp.MustCompile(`PORT=(\d+)`)
			if match := re.FindStringSubmatch(contentStr); len(match) > 1 {
				if port := parseInt(match[1]); port > 0 {
					d.Port = port
					return
				}
			}
		}
	}
}

func findEnvConfig(projectPath string, d *ProjectDetection) {
	// Look for .env.example or similar files
	envFiles := []string{".env.example", ".env.sample", ".env.template", ".env"}
	
	for _, file := range envFiles {
		fullPath := filepath.Join(projectPath, file)
		if exists(fullPath) {
			d.EnvFile = file
			
			// Parse environment variables
			if file != ".env" { // Don't parse actual .env for security
				parseEnvFile(fullPath, d)
			}
			break
		}
	}
}

func parseEnvFile(filePath string, d *ProjectDetection) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Determine if required based on common patterns
			if isRequiredEnvVar(key, value) {
				d.Environment.Required = append(d.Environment.Required, key)
			} else {
				d.Environment.Optional[key] = value
			}
		}
	}
}

func isRequiredEnvVar(key, value string) bool {
	// Common required patterns
	requiredPatterns := []string{
		"DATABASE_URL",
		"DB_",
		"SECRET",
		"KEY",
		"TOKEN",
		"PASSWORD",
		"CREDENTIALS",
		"API_KEY",
	}
	
	for _, pattern := range requiredPatterns {
		if strings.Contains(key, pattern) && (value == "" || value == "CHANGE_ME" || value == "REQUIRED") {
			return true
		}
	}
	
	return false
}

func setHealthCheck(d *ProjectDetection) {
	// Check if we detected a health endpoint
	for _, endpoint := range d.Endpoints {
		if endpoint.Method == "GET" && 
		   (endpoint.Path == "/health" || endpoint.Path == "/healthz" || 
		    endpoint.Path == "/health-check" || endpoint.Path == "/_health") {
			d.HealthCheck = endpoint.Path
			return
		}
	}
	
	// Default health check
	d.HealthCheck = "/health"
}

// Helper functions

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func findFirst(basePath string, files ...string) string {
	for _, file := range files {
		if exists(filepath.Join(basePath, file)) {
			return file
		}
	}
	return ""
}

func getPackageMain(projectPath string) string {
	packagePath := filepath.Join(projectPath, "package.json")
	if content, err := ioutil.ReadFile(packagePath); err == nil {
		// Simple regex to find main field
		re := regexp.MustCompile(`"main"\s*:\s*"([^"]+)"`)
		if match := re.FindStringSubmatch(string(content)); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}