package detector

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeProject(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T) string
		wantErr     bool
		validate    func(*testing.T, *ProjectDetection)
	}{
		{
			name: "Python FastAPI project",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				
				// Create requirements.txt
				err := ioutil.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(`fastapi==0.95.0
uvicorn==0.21.0
pydantic==1.10.7`), 0644)
				require.NoError(t, err)
				
				// Create main.py
				mainContent := `from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "World"}

@app.get("/health")
def health_check():
    return {"status": "healthy"}

@app.post("/items/{item_id}")
def create_item(item_id: int):
    return {"item_id": item_id}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
`
				err = ioutil.WriteFile(filepath.Join(dir, "main.py"), []byte(mainContent), 0644)
				require.NoError(t, err)
				
				// Create .env.example
				envContent := `DATABASE_URL=
API_KEY=REQUIRED
DEBUG=false
PORT=8000`
				err = ioutil.WriteFile(filepath.Join(dir, ".env.example"), []byte(envContent), 0644)
				require.NoError(t, err)
				
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, d *ProjectDetection) {
				assert.Equal(t, "Python", d.Language)
				assert.Equal(t, "python3.11", d.Runtime)
				assert.Equal(t, "FastAPI", d.Framework)
				assert.Equal(t, "main.py", d.MainFile)
				assert.Equal(t, "requirements.txt", d.RequirementsFile)
				assert.Equal(t, 8000, d.Port)
				assert.Contains(t, d.StartCommand, "uvicorn")
				assert.Equal(t, "/health", d.HealthCheck)
				
				// Check endpoints
				assert.Len(t, d.Endpoints, 3)
				assert.Contains(t, d.Endpoints, Endpoint{Method: "GET", Path: "/"})
				assert.Contains(t, d.Endpoints, Endpoint{Method: "GET", Path: "/health"})
				assert.Contains(t, d.Endpoints, Endpoint{Method: "POST", Path: "/items/{item_id}"})
				
				// Check environment
				assert.Contains(t, d.Environment.Required, "DATABASE_URL")
				assert.Contains(t, d.Environment.Required, "API_KEY")
				assert.Equal(t, "false", d.Environment.Optional["DEBUG"])
			},
		},
		{
			name: "Node.js Express project",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				
				// Create package.json
				packageJSON := `{
  "name": "express-api",
  "version": "1.0.0",
  "main": "server.js",
  "dependencies": {
    "express": "^4.18.2",
    "dotenv": "^16.0.3"
  }
}`
				err := ioutil.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
				require.NoError(t, err)
				
				// Create server.js
				serverContent := `const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello World' });
});

app.get('/healthz', (req, res) => {
  res.json({ status: 'ok' });
});

app.post('/api/users', (req, res) => {
  res.json({ user: 'created' });
});

app.listen(PORT, () => {
  console.log('Server running on port ' + PORT);
});
`
				err = ioutil.WriteFile(filepath.Join(dir, "server.js"), []byte(serverContent), 0644)
				require.NoError(t, err)
				
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, d *ProjectDetection) {
				assert.Equal(t, "Node.js", d.Language)
				assert.Equal(t, "node18", d.Runtime)
				assert.Equal(t, "Express", d.Framework)
				assert.Equal(t, "server.js", d.MainFile)
				assert.Equal(t, "package.json", d.RequirementsFile)
				assert.Equal(t, 8080, d.Port) // Default port when not detected
				assert.Equal(t, "node server.js", d.StartCommand)
				assert.Equal(t, "/healthz", d.HealthCheck)
				
				// Check endpoints
				assert.Len(t, d.Endpoints, 3)
				assert.Contains(t, d.Endpoints, Endpoint{Method: "GET", Path: "/"})
				assert.Contains(t, d.Endpoints, Endpoint{Method: "GET", Path: "/healthz"})
				assert.Contains(t, d.Endpoints, Endpoint{Method: "POST", Path: "/api/users"})
			},
		},
		{
			name: "Go Gin project",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				
				// Create go.mod
				goMod := `module example.com/api

go 1.21

require github.com/gin-gonic/gin v1.9.0
`
				err := ioutil.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
				require.NoError(t, err)
				
				// Create main.go
				mainContent := `package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello"})
    })
    
    r.Run(":8080")
}
`
				err = ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(mainContent), 0644)
				require.NoError(t, err)
				
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, d *ProjectDetection) {
				assert.Equal(t, "Go", d.Language)
				assert.Equal(t, "go1.21", d.Runtime)
				assert.Equal(t, "Gin", d.Framework)
				assert.Equal(t, "main.go", d.MainFile)
				assert.Equal(t, "go.mod", d.RequirementsFile)
				assert.Equal(t, 8080, d.Port)
				assert.Equal(t, "./app", d.StartCommand)
			},
		},
		{
			name: "Ruby Rails project",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				
				// Create Gemfile
				gemfile := `source 'https://rubygems.org'

gem 'rails', '~> 7.0.0'
gem 'puma'
`
				err := ioutil.WriteFile(filepath.Join(dir, "Gemfile"), []byte(gemfile), 0644)
				require.NoError(t, err)
				
				// Create config.ru (standard Rack file for Rails)
				configRu := `require_relative "config/environment"
run Rails.application`
				err = ioutil.WriteFile(filepath.Join(dir, "config.ru"), []byte(configRu), 0644)
				require.NoError(t, err)
				
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, d *ProjectDetection) {
				assert.Equal(t, "Ruby", d.Language)
				assert.Equal(t, "ruby3.0", d.Runtime)
				assert.Equal(t, "Rails", d.Framework)
				assert.Equal(t, "config.ru", d.MainFile)
				assert.Equal(t, "Gemfile", d.RequirementsFile)
				assert.Contains(t, d.StartCommand, "rails server")
			},
		},
		{
			name: "unknown project type",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				// Create some random file
				err := ioutil.WriteFile(filepath.Join(dir, "random.txt"), []byte("content"), 0644)
				require.NoError(t, err)
				return dir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := tt.setupFunc(t)
			
			detection, err := AnalyzeProject(projectPath)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, detection)
				if tt.validate != nil {
					tt.validate(t, detection)
				}
			}
		})
	}
}

func TestDetectLanguageAndFramework(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		want      string
		framework string
	}{
		{
			name: "Python with requirements.txt",
			files: map[string]string{
				"requirements.txt": "flask==2.0.0",
			},
			want:      "Python",
			framework: "",
		},
		{
			name: "Python with Pipfile",
			files: map[string]string{
				"Pipfile": "[packages]\ndjango = '*'",
			},
			want:      "Python",
			framework: "",
		},
		{
			name: "Node.js project",
			files: map[string]string{
				"package.json": `{"name": "test"}`,
			},
			want:      "Node.js",
			framework: "",
		},
		{
			name: "Go project",
			files: map[string]string{
				"go.mod": "module test",
			},
			want:      "Go",
			framework: "",
		},
		{
			name: "Ruby project",
			files: map[string]string{
				"Gemfile": "gem 'sinatra'",
			},
			want:      "Ruby",
			framework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			
			// Create files
			for filename, content := range tt.files {
				err := ioutil.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
				require.NoError(t, err)
			}
			
			d := &ProjectDetection{}
			err := detectLanguageAndFramework(dir, d)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.want, d.Language)
		})
	}
}

func TestDetectPythonFramework(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		framework string
	}{
		{
			name: "FastAPI",
			content: `from fastapi import FastAPI
app = FastAPI()`,
			framework: "FastAPI",
		},
		{
			name: "Flask",
			content: `from flask import Flask
app = Flask(__name__)`,
			framework: "Flask",
		},
		{
			name: "Django",
			content: `import django
INSTALLED_APPS = []`,
			framework: "Django",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			err := ioutil.WriteFile(filepath.Join(dir, "app.py"), []byte(tt.content), 0644)
			require.NoError(t, err)
			
			d := &ProjectDetection{}
			detectPythonFramework(dir, d)
			
			assert.Equal(t, tt.framework, d.Framework)
		})
	}
}

func TestDetectEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		content   string
		expected  []Endpoint
	}{
		{
			name:      "FastAPI endpoints",
			framework: "FastAPI",
			content: `@app.get("/")
def root():
    pass

@app.post("/items")
def create():
    pass

@router.put('/users/{id}')
def update():
    pass`,
			expected: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "POST", Path: "/items"},
				{Method: "PUT", Path: "/users/{id}"},
			},
		},
		{
			name:      "Flask endpoints",
			framework: "Flask",
			content: `@app.route("/")
def home():
    pass

@app.route("/api", methods=["POST", "PUT"])
def api():
    pass`,
			expected: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "POST", Path: "/api"},
				{Method: "PUT", Path: "/api"},
			},
		},
		{
			name:      "Express endpoints",
			framework: "Express",
			content: `app.get('/', handler);
app.post("/api/users", handler);
router.delete('/items/:id', handler);`,
			expected: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "POST", Path: "/api/users"},
				{Method: "DELETE", Path: "/items/:id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var endpoints []Endpoint
			
			switch tt.framework {
			case "FastAPI":
				endpoints = detectFastAPIEndpoints(tt.content)
			case "Flask":
				endpoints = detectFlaskEndpoints(tt.content)
			case "Express":
				endpoints = detectExpressEndpoints(tt.content)
			}
			
			assert.Equal(t, tt.expected, endpoints)
		})
	}
}

func TestDetectPort(t *testing.T) {
	tests := []struct {
		name    string
		content string
		envFile string
		want    int
	}{
		{
			name:    "port in code",
			content: `app.listen(3000)`,
			want:    3000,
		},
		{
			name:    "port from environment",
			content: `const PORT = process.env.PORT || 5000`,
			envFile: "PORT=4000",
			want:    4000,
		},
		{
			name:    "Python uvicorn port",
			content: `uvicorn.run(app, host="0.0.0.0", port=8001)`,
			want:    8001,
		},
		{
			name:    "default port",
			content: `// no port specified`,
			want:    8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			
			d := &ProjectDetection{
				MainFile: "main.js",
				Port:     8080, // default
			}
			
			// Create main file
			err := ioutil.WriteFile(filepath.Join(dir, "main.js"), []byte(tt.content), 0644)
			require.NoError(t, err)
			
			// Create env file if specified
			if tt.envFile != "" {
				err = ioutil.WriteFile(filepath.Join(dir, ".env"), []byte(tt.envFile), 0644)
				require.NoError(t, err)
			}
			
			detectPort(dir, d)
			assert.Equal(t, tt.want, d.Port)
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantRequired []string
		wantOptional map[string]string
	}{
		{
			name: "mixed environment variables",
			content: `# Database configuration
DATABASE_URL=
DB_PASSWORD=REQUIRED

# API Keys
API_KEY=CHANGE_ME
SECRET_KEY=

# Optional settings
DEBUG=false
LOG_LEVEL=info
PORT=8080`,
			wantRequired: []string{"DATABASE_URL", "DB_PASSWORD", "API_KEY", "SECRET_KEY"},
			wantOptional: map[string]string{
				"DEBUG":     "false",
				"LOG_LEVEL": "info",
				"PORT":      "8080",
			},
		},
		{
			name: "all optional",
			content: `DEBUG=true
NODE_ENV=development
REDIS_URL=redis://localhost:6379`,
			wantRequired: []string{},
			wantOptional: map[string]string{
				"DEBUG":     "true",
				"NODE_ENV":  "development",
				"REDIS_URL": "redis://localhost:6379",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			envFile := filepath.Join(dir, ".env.example")
			
			err := ioutil.WriteFile(envFile, []byte(tt.content), 0644)
			require.NoError(t, err)
			
			d := &ProjectDetection{
				Environment: EnvironmentVars{
					Optional: make(map[string]string),
				},
			}
			
			parseEnvFile(envFile, d)
			
			assert.ElementsMatch(t, tt.wantRequired, d.Environment.Required)
			assert.Equal(t, tt.wantOptional, d.Environment.Optional)
		})
	}
}

func TestGenerateStartCommand(t *testing.T) {
	tests := []struct {
		name      string
		detection ProjectDetection
		want      string
	}{
		{
			name: "FastAPI",
			detection: ProjectDetection{
				Framework: "FastAPI",
				MainFile:  "main.py",
				Port:      8000,
			},
			want: "uvicorn main:app --host 0.0.0.0 --port 8000",
		},
		{
			name: "Flask with gunicorn",
			detection: ProjectDetection{
				Framework: "Flask",
				MainFile:  "wsgi.py",
				Port:      5000,
			},
			want: "gunicorn wsgi:app --bind 0.0.0.0:5000",
		},
		{
			name: "Express",
			detection: ProjectDetection{
				Framework: "Express",
				MainFile:  "server.js",
			},
			want: "node server.js",
		},
		{
			name: "Rails",
			detection: ProjectDetection{
				Framework: "Rails",
				Port:      3000,
			},
			want: "rails server -b 0.0.0.0 -p 3000",
		},
		{
			name: "Generic Python",
			detection: ProjectDetection{
				Language: "Python",
				MainFile: "app.py",
			},
			want: "python app.py",
		},
		{
			name: "Generic Go",
			detection: ProjectDetection{
				Language: "Go",
				MainFile: "main.go",
			},
			want: "./app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateStartCommand(&tt.detection)
			assert.Equal(t, tt.want, tt.detection.StartCommand)
		})
	}
}

func TestSetHealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []Endpoint
		want      string
	}{
		{
			name: "has /health endpoint",
			endpoints: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "GET", Path: "/health"},
			},
			want: "/health",
		},
		{
			name: "has /healthz endpoint",
			endpoints: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "GET", Path: "/healthz"},
			},
			want: "/healthz",
		},
		{
			name: "has /_health endpoint",
			endpoints: []Endpoint{
				{Method: "GET", Path: "/_health"},
			},
			want: "/_health",
		},
		{
			name: "no health endpoint",
			endpoints: []Endpoint{
				{Method: "GET", Path: "/"},
				{Method: "POST", Path: "/api"},
			},
			want: "/health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ProjectDetection{
				Endpoints: tt.endpoints,
			}
			setHealthCheck(d)
			assert.Equal(t, tt.want, d.HealthCheck)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		testFile := filepath.Join(dir, "test.txt")
		
		// File doesn't exist
		assert.False(t, exists(testFile))
		
		// Create file
		err := ioutil.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// File exists
		assert.True(t, exists(testFile))
	})
	
	t.Run("findFirst", func(t *testing.T) {
		dir := t.TempDir()
		
		// Create second file
		err := ioutil.WriteFile(filepath.Join(dir, "second.txt"), []byte("test"), 0644)
		require.NoError(t, err)
		
		result := findFirst(dir, "first.txt", "second.txt", "third.txt")
		assert.Equal(t, "second.txt", result)
		
		// No files exist
		result = findFirst(dir, "none1.txt", "none2.txt")
		assert.Equal(t, "", result)
	})
	
	t.Run("parseInt", func(t *testing.T) {
		assert.Equal(t, 8080, parseInt("8080"))
		assert.Equal(t, 0, parseInt("invalid"))
		assert.Equal(t, 3000, parseInt("3000"))
	})
	
	t.Run("getPackageMain", func(t *testing.T) {
		dir := t.TempDir()
		
		// Create package.json with main field
		packageJSON := `{
  "name": "test",
  "main": "index.js",
  "version": "1.0.0"
}`
		err := ioutil.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)
		
		main := getPackageMain(dir)
		assert.Equal(t, "index.js", main)
	})
}

func TestIsRequiredEnvVar(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		required bool
	}{
		{"DATABASE_URL", "", true},
		{"DB_HOST", "CHANGE_ME", true},
		{"API_KEY", "REQUIRED", true},
		{"SECRET_TOKEN", "", true},
		{"DEBUG", "false", false},
		{"PORT", "8080", false},
		{"LOG_LEVEL", "info", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := isRequiredEnvVar(tt.key, tt.value)
			assert.Equal(t, tt.required, result)
		})
	}
}

func TestComplexProjectStructure(t *testing.T) {
	// Test detection in nested directory structure
	dir := t.TempDir()
	
	// Create src directory
	srcDir := filepath.Join(dir, "src")
	err := os.MkdirAll(srcDir, 0755)
	require.NoError(t, err)
	
	// Create package.json
	packageJSON := `{"name": "complex-app", "dependencies": {"express": "^4.18.0"}}`
	err = ioutil.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
	require.NoError(t, err)
	
	// Create main file in src
	serverContent := `const express = require('express');
const app = express();
app.get('/', (req, res) => res.json({message: 'Hello'}));
app.listen(process.env.PORT || 4000);`
	err = ioutil.WriteFile(filepath.Join(srcDir, "server.js"), []byte(serverContent), 0644)
	require.NoError(t, err)
	
	detection, err := AnalyzeProject(dir)
	assert.NoError(t, err)
	assert.Equal(t, "Node.js", detection.Language)
	assert.Equal(t, "Express", detection.Framework)
	assert.Equal(t, "src/server.js", detection.MainFile)
	assert.Equal(t, 8080, detection.Port) // Default port when complex expression not parsed
}