package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// APITemplate represents a template for API creation (imported from wizard package)
type APITemplate struct {
	ID          string
	Name        string
	Description string
	Runtime     string
	Category    string
	Features    []string
}

// InitPythonProject initializes a new Python API project
func InitPythonProject(apiName, runtime string) error {
	projectPath := apiName

	// Create project structure
	dirs := []string{
		"",
		"tests",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	// Create files
	files := map[string]string{
		"apidirect.yaml":    getPythonConfigTemplate(apiName, runtime),
		"main.py":           getPythonMainTemplate(),
		"requirements.txt":  getPythonRequirementsTemplate(),
		".gitignore":        getPythonGitignoreTemplate(),
		"README.md":         getReadmeTemplate(apiName, "Python"),
		"tests/__init__.py": "",
		"tests/test_main.py": getPythonTestTemplate(),
	}

	for filename, content := range files {
		fullPath := filepath.Join(projectPath, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}

	return nil
}

// InitNodeProject initializes a new Node.js API project
func InitNodeProject(apiName, runtime string) error {
	projectPath := apiName

	// Create project structure
	dirs := []string{
		"",
		"tests",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	// Create files
	files := map[string]string{
		"apidirect.yaml":   getNodeConfigTemplate(apiName, runtime),
		"main.js":          getNodeMainTemplate(),
		"package.json":     getNodePackageTemplate(apiName),
		".gitignore":       getNodeGitignoreTemplate(),
		"README.md":        getReadmeTemplate(apiName, "Node.js"),
		"tests/main.test.js": getNodeTestTemplate(),
	}

	for filename, content := range files {
		fullPath := filepath.Join(projectPath, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}

	return nil
}

// Template functions

func getPythonConfigTemplate(apiName, runtime string) string {
	return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /hello
    method: GET
    handler: main.hello_world
  
  - path: /hello/{name}
    method: GET
    handler: main.hello_name
  
  - path: /data
    method: POST
    handler: main.process_data

# Environment Variables
environment:
  # Add your environment variables here
  # API_KEY: ${API_KEY}
  LOG_LEVEL: INFO
`, apiName, runtime)
}

func getPythonMainTemplate() string {
	return `"""
API-Direct Python API Template
This is a basic template for your serverless API.
"""
import json
import logging
import os
from typing import Dict, Any

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)


def hello_world(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Simple hello world endpoint
    """
    logger.info("Hello world endpoint called")
    
    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps({
            'message': 'Hello from API-Direct!',
            'timestamp': str(context.get_remaining_time_in_millis()) if hasattr(context, 'get_remaining_time_in_millis') else None
        })
    }


def hello_name(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Personalized greeting endpoint
    """
    path_params = event.get('pathParameters', {})
    name = path_params.get('name', 'Anonymous')
    
    logger.info(f"Hello name endpoint called with name: {name}")
    
    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps({
            'message': f'Hello, {name}!',
            'path': event.get('path', '')
        })
    }


def process_data(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Process POST data endpoint
    """
    try:
        # Parse request body
        body = json.loads(event.get('body', '{}'))
        logger.info(f"Processing data: {body}")
        
        # Example processing
        result = {
            'received': body,
            'processed': True,
            'item_count': len(body) if isinstance(body, (list, dict)) else 1
        }
        
        return {
            'statusCode': 200,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps(result)
        }
    except json.JSONDecodeError as e:
        logger.error(f"JSON decode error: {e}")
        return {
            'statusCode': 400,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': 'Invalid JSON in request body'
            })
        }
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        return {
            'statusCode': 500,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': 'Internal server error'
            })
        }
`
}

func getPythonRequirementsTemplate() string {
	return `# Add your Python dependencies here
# Example:
# requests==2.31.0
# pandas==2.0.3
# numpy==1.24.3
`
}

func getPythonGitignoreTemplate() string {
	return `# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
env/
venv/
ENV/
.venv
pip-log.txt
pip-delete-this-directory.txt
.pytest_cache/
.coverage
.coverage.*
coverage.xml
*.cover
.hypothesis/

# API-Direct
.apidirect/
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
}

func getPythonTestTemplate() string {
	return `"""
Tests for the API handlers
"""
import json
import unittest
from main import hello_world, hello_name, process_data


class TestAPIHandlers(unittest.TestCase):
    
    def test_hello_world(self):
        event = {}
        context = {}
        
        response = hello_world(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertIn('message', body)
        self.assertEqual(body['message'], 'Hello from API-Direct!')
    
    def test_hello_name(self):
        event = {
            'pathParameters': {
                'name': 'TestUser'
            }
        }
        context = {}
        
        response = hello_name(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertEqual(body['message'], 'Hello, TestUser!')
    
    def test_process_data(self):
        test_data = {'key': 'value', 'number': 42}
        event = {
            'body': json.dumps(test_data)
        }
        context = {}
        
        response = process_data(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertEqual(body['received'], test_data)
        self.assertTrue(body['processed'])


if __name__ == '__main__':
    unittest.main()
`
}

func getNodeConfigTemplate(apiName, runtime string) string {
	return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /hello
    method: GET
    handler: main.helloWorld
  
  - path: /hello/{name}
    method: GET
    handler: main.helloName
  
  - path: /data
    method: POST
    handler: main.processData

# Environment Variables
environment:
  # Add your environment variables here
  # API_KEY: ${API_KEY}
  LOG_LEVEL: info
`, apiName, runtime)
}

func getNodeMainTemplate() string {
	return `/**
 * API-Direct Node.js API Template
 * This is a basic template for your serverless API.
 */

const LOG_LEVEL = process.env.LOG_LEVEL || 'info';

/**
 * Simple hello world endpoint
 */
exports.helloWorld = async (event, context) => {
    console.log('Hello world endpoint called');
    
    return {
        statusCode: 200,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            message: 'Hello from API-Direct!',
            timestamp: new Date().toISOString()
        })
    };
};

/**
 * Personalized greeting endpoint
 */
exports.helloName = async (event, context) => {
    const name = event.pathParameters?.name || 'Anonymous';
    
    console.log(` + "`Hello name endpoint called with name: ${name}`" + `);
    
    return {
        statusCode: 200,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            message: ` + "`Hello, ${name}!`" + `,
            path: event.path || ''
        })
    };
};

/**
 * Process POST data endpoint
 */
exports.processData = async (event, context) => {
    try {
        // Parse request body
        const body = JSON.parse(event.body || '{}');
        console.log('Processing data:', body);
        
        // Example processing
        const result = {
            received: body,
            processed: true,
            itemCount: Array.isArray(body) ? body.length : Object.keys(body).length
        };
        
        return {
            statusCode: 200,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(result)
        };
    } catch (error) {
        console.error('Error processing request:', error);
        
        if (error instanceof SyntaxError) {
            return {
                statusCode: 400,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    error: 'Invalid JSON in request body'
                })
            };
        }
        
        return {
            statusCode: 500,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                error: 'Internal server error'
            })
        };
    }
};
`
}

func getNodePackageTemplate(apiName string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "API-Direct serverless API",
  "main": "main.js",
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch"
  },
  "keywords": ["api", "serverless", "api-direct"],
  "author": "",
  "license": "MIT",
  "dependencies": {
  },
  "devDependencies": {
    "jest": "^29.5.0"
  }
}
`, apiName)
}

func getNodeGitignoreTemplate() string {
	return `# Node
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.npm
.yarn-integrity

# API-Direct
.apidirect/
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Testing
coverage/
.nyc_output/
`
}

func getNodeTestTemplate() string {
	return `/**
 * Tests for the API handlers
 */
const { helloWorld, helloName, processData } = require('../main');

describe('API Handlers', () => {
    describe('helloWorld', () => {
        it('should return a hello message', async () => {
            const event = {};
            const context = {};
            
            const response = await helloWorld(event, context);
            
            expect(response.statusCode).toBe(200);
            const body = JSON.parse(response.body);
            expect(body.message).toBe('Hello from API-Direct!');
            expect(body.timestamp).toBeDefined();
        });
    });
    
    describe('helloName', () => {
        it('should return a personalized greeting', async () => {
            const event = {
                pathParameters: {
                    name: 'TestUser'
                }
            };
            const context = {};
            
            const response = await helloName(event, context);
            
            expect(response.statusCode).toBe(200);
            const body = JSON.parse(response.body);
            expect(body.message).toBe('Hello, TestUser!');
        });
        
        it('should handle missing name parameter', async () => {
            const event = {};
            const context = {};
            
            const response = await helloName(event, context);
            
            expect(response.statusCode).toBe(200);
            const body = JSON.parse(response.body);
            expect(body.message).toBe('Hello, Anonymous!');
        });
    });
    
    describe('processData', () => {
        it('should process valid JSON data', async () => {
            const testData = { key: 'value', number: 42 };
            const event = {
                body: JSON.stringify(testData)
            };
            const context = {};
            
            const response = await processData(event, context);
            
            expect(response.statusCode).toBe(200);
            const body = JSON.parse(response.body);
            expect(body.received).toEqual(testData);
            expect(body.processed).toBe(true);
            expect(body.itemCount).toBe(2);
        });
        
        it('should handle invalid JSON', async () => {
            const event = {
                body: 'invalid json'
            };
            const context = {};
            
            const response = await processData(event, context);
            
            expect(response.statusCode).toBe(400);
            const body = JSON.parse(response.body);
            expect(body.error).toBe('Invalid JSON in request body');
        });
    });
});
`
}

func getReadmeTemplate(apiName, language string) string {
	return fmt.Sprintf(`# %s

This is an API-Direct %s API project.

## Getting Started

1. **Install dependencies** (if using Node.js):
   ` + "```bash" + `
   npm install
   ` + "```" + `

2. **Configure your API**:
   Edit ` + "`apidirect.yaml`" + ` to define your endpoints and settings.

3. **Implement your logic**:
   Edit ` + "`main.py`" + ` or ` + "`main.js`" + ` to implement your API handlers.

4. **Test locally**:
   ` + "```bash" + `
   apidirect run
   ` + "```" + `

5. **Deploy to API-Direct**:
   ` + "```bash" + `
   apidirect deploy
   ` + "```" + `

## Project Structure

- ` + "`apidirect.yaml`" + ` - API configuration file
- ` + "`main.py/js`" + ` - Main API implementation
- ` + "`requirements.txt/package.json`" + ` - Dependencies
- ` + "`tests/`" + ` - Test files

## Available Endpoints

After deployment, your API will have the following endpoints:

- ` + "`GET /hello`" + ` - Returns a hello world message
- ` + "`GET /hello/{name}`" + ` - Returns a personalized greeting
- ` + "`POST /data`" + ` - Processes posted data

## Environment Variables

You can set environment variables in ` + "`apidirect.yaml`" + ` or using the CLI:

` + "```bash" + `
apidirect env set KEY=value
` + "```" + `

## Publishing to Marketplace

Once your API is deployed and tested, you can publish it to the API-Direct marketplace:

` + "```bash" + `
apidirect publish %s
` + "```" + `

## Need Help?

- Documentation: https://docs.api-direct.io
- Support: support@api-direct.io
`, apiName, language, apiName)
}

// InitPythonProjectWithTemplate initializes a Python project with a specific template
func InitPythonProjectWithTemplate(apiName, runtime string, template APITemplate, features []string) error {
	projectPath := apiName

	// Create project structure based on template
	dirs := getProjectDirs(template, features)
	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	// Create files based on template
	files := getPythonTemplateFiles(apiName, runtime, template, features)
	for filename, content := range files {
		fullPath := filepath.Join(projectPath, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}

	return nil
}

// InitNodeProjectWithTemplate initializes a Node.js project with a specific template
func InitNodeProjectWithTemplate(apiName, runtime string, template APITemplate, features []string) error {
	projectPath := apiName

	// Create project structure based on template
	dirs := getProjectDirs(template, features)
	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	// Create files based on template
	files := getNodeTemplateFiles(apiName, runtime, template, features)
	for filename, content := range files {
		fullPath := filepath.Join(projectPath, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}

	return nil
}

// getProjectDirs returns the directory structure based on template and features
func getProjectDirs(template APITemplate, features []string) []string {
	dirs := []string{"", "tests"}
	
	// Add directories based on template
	switch template.ID {
	case "crud-database":
		dirs = append(dirs, "models", "migrations")
	case "ml-model-serving":
		dirs = append(dirs, "models", "data")
	case "data-processing":
		dirs = append(dirs, "processors", "uploads")
	case "microservice":
		dirs = append(dirs, "health", "metrics")
	}
	
	// Add directories based on features
	for _, feature := range features {
		switch feature {
		case "Docker support":
			// Dockerfile will be created in root
		case "GitHub Actions CI/CD":
			dirs = append(dirs, ".github/workflows")
		case "API documentation generation":
			dirs = append(dirs, "docs")
		}
	}
	
	return dirs
}

// getPythonTemplateFiles returns the files to create for a Python template
func getPythonTemplateFiles(apiName, runtime string, template APITemplate, features []string) map[string]string {
	files := map[string]string{
		"apidirect.yaml":     getPythonTemplateConfig(apiName, runtime, template),
		"main.py":            getPythonTemplateMain(template),
		"requirements.txt":   getPythonTemplateRequirements(template, features),
		".gitignore":         getPythonGitignoreTemplate(),
		"README.md":          getTemplateReadme(apiName, "Python", template, features),
		"tests/__init__.py":  "",
		"tests/test_main.py": getPythonTemplateTests(template),
	}
	
	// Add feature-specific files
	for _, feature := range features {
		switch feature {
		case "Docker support":
			files["Dockerfile"] = getPythonDockerfile(runtime)
			files[".dockerignore"] = getDockerignore()
		case "GitHub Actions CI/CD":
			files[".github/workflows/deploy.yml"] = getGitHubActionsWorkflow()
		case "API documentation generation":
			files["docs/api.md"] = getAPIDocumentation(template)
		}
	}
	
	return files
}

// getNodeTemplateFiles returns the files to create for a Node.js template
func getNodeTemplateFiles(apiName, runtime string, template APITemplate, features []string) map[string]string {
	files := map[string]string{
		"apidirect.yaml":     getNodeTemplateConfig(apiName, runtime, template),
		"main.js":            getNodeTemplateMain(template),
		"package.json":       getNodeTemplatePackage(apiName, template, features),
		".gitignore":         getNodeGitignoreTemplate(),
		"README.md":          getTemplateReadme(apiName, "Node.js", template, features),
		"tests/main.test.js": getNodeTemplateTests(template),
	}
	
	// Add feature-specific files
	for _, feature := range features {
		switch feature {
		case "Docker support":
			files["Dockerfile"] = getNodeDockerfile(runtime)
			files[".dockerignore"] = getDockerignore()
		case "GitHub Actions CI/CD":
			files[".github/workflows/deploy.yml"] = getGitHubActionsWorkflow()
		case "API documentation generation":
			files["docs/api.md"] = getAPIDocumentation(template)
		}
	}
	
	return files
}

// Template-specific configuration generators
func getPythonTemplateConfig(apiName, runtime string, template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /items
    method: GET
    handler: main.list_items
  
  - path: /items
    method: POST
    handler: main.create_item
  
  - path: /items/{id}
    method: GET
    handler: main.get_item
  
  - path: /items/{id}
    method: PUT
    handler: main.update_item
  
  - path: /items/{id}
    method: DELETE
    handler: main.delete_item

# Environment Variables
environment:
  DATABASE_URL: ${DATABASE_URL}
  LOG_LEVEL: INFO
`, apiName, runtime)
	
	case "webhook-receiver":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /webhook
    method: POST
    handler: main.receive_webhook
  
  - path: /webhook/status
    method: GET
    handler: main.webhook_status

# Environment Variables
environment:
  WEBHOOK_SECRET: ${WEBHOOK_SECRET}
  LOG_LEVEL: INFO
`, apiName, runtime)
	
	default:
		return getPythonConfigTemplate(apiName, runtime)
	}
}

func getNodeTemplateConfig(apiName, runtime string, template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /items
    method: GET
    handler: main.listItems
  
  - path: /items
    method: POST
    handler: main.createItem
  
  - path: /items/{id}
    method: GET
    handler: main.getItem
  
  - path: /items/{id}
    method: PUT
    handler: main.updateItem
  
  - path: /items/{id}
    method: DELETE
    handler: main.deleteItem

# Environment Variables
environment:
  DATABASE_URL: ${DATABASE_URL}
  LOG_LEVEL: info
`, apiName, runtime)
	
	default:
		return getNodeConfigTemplate(apiName, runtime)
	}
}

// Template-specific main file generators
func getPythonTemplateMain(template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return `"""
CRUD Database API Template
A REST API with database operations using PostgreSQL.
"""
import json
import logging
import os
from typing import Dict, Any, List

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Mock database for demonstration
# In production, replace with actual database connection
ITEMS_DB = {}
NEXT_ID = 1

def list_items(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """List all items"""
    logger.info("Listing all items")
    
    items = list(ITEMS_DB.values())
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps({
            'items': items,
            'total': len(items)
        })
    }

def create_item(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Create a new item"""
    global NEXT_ID
    
    try:
        body = json.loads(event.get('body', '{}'))
        
        item = {
            'id': NEXT_ID,
            'name': body.get('name'),
            'description': body.get('description'),
            'created_at': '2024-01-01T00:00:00Z'
        }
        
        ITEMS_DB[NEXT_ID] = item
        NEXT_ID += 1
        
        logger.info(f"Created item: {item['id']}")
        
        return {
            'statusCode': 201,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(item)
        }
    except Exception as e:
        logger.error(f"Error creating item: {e}")
        return {
            'statusCode': 400,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Invalid request'})
        }

def get_item(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Get a specific item"""
    item_id = int(event.get('pathParameters', {}).get('id', 0))
    
    if item_id not in ITEMS_DB:
        return {
            'statusCode': 404,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Item not found'})
        }
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps(ITEMS_DB[item_id])
    }

def update_item(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Update an existing item"""
    item_id = int(event.get('pathParameters', {}).get('id', 0))
    
    if item_id not in ITEMS_DB:
        return {
            'statusCode': 404,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Item not found'})
        }
    
    try:
        body = json.loads(event.get('body', '{}'))
        item = ITEMS_DB[item_id]
        
        item.update({
            'name': body.get('name', item['name']),
            'description': body.get('description', item['description'])
        })
        
        logger.info(f"Updated item: {item_id}")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(item)
        }
    except Exception as e:
        logger.error(f"Error updating item: {e}")
        return {
            'statusCode': 400,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Invalid request'})
        }

def delete_item(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Delete an item"""
    item_id = int(event.get('pathParameters', {}).get('id', 0))
    
    if item_id not in ITEMS_DB:
        return {
            'statusCode': 404,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Item not found'})
        }
    
    del ITEMS_DB[item_id]
    logger.info(f"Deleted item: {item_id}")
    
    return {
        'statusCode': 204,
        'headers': {'Content-Type': 'application/json'},
        'body': ''
    }
`
	
	default:
		return getPythonMainTemplate()
	}
}

func getNodeTemplateMain(template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return `/**
 * CRUD Database API Template
 * A REST API with database operations using PostgreSQL.
 */

// Mock database for demonstration
// In production, replace with actual database connection
const itemsDB = {};
let nextId = 1;

exports.listItems = async (event, context) => {
    console.log('Listing all items');
    
    const items = Object.values(itemsDB);
    
    return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            items: items,
            total: items.length
        })
    };
};

exports.createItem = async (event, context) => {
    try {
        const body = JSON.parse(event.body || '{}');
        
        const item = {
            id: nextId,
            name: body.name,
            description: body.description,
            createdAt: new Date().toISOString()
        };
        
        itemsDB[nextId] = item;
        nextId++;
        
        console.log(` + "`Created item: ${item.id}`" + `);
        
        return {
            statusCode: 201,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item)
        };
    } catch (error) {
        console.error('Error creating item:', error);
        return {
            statusCode: 400,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ error: 'Invalid request' })
        };
    }
};

exports.getItem = async (event, context) => {
    const itemId = parseInt(event.pathParameters?.id || '0');
    
    if (!itemsDB[itemId]) {
        return {
            statusCode: 404,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ error: 'Item not found' })
        };
    }
    
    return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(itemsDB[itemId])
    };
};

exports.updateItem = async (event, context) => {
    const itemId = parseInt(event.pathParameters?.id || '0');
    
    if (!itemsDB[itemId]) {
        return {
            statusCode: 404,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ error: 'Item not found' })
        };
    }
    
    try {
        const body = JSON.parse(event.body || '{}');
        const item = itemsDB[itemId];
        
        Object.assign(item, {
            name: body.name || item.name,
            description: body.description || item.description
        });
        
        console.log(` + "`Updated item: ${itemId}`" + `);
        
        return {
            statusCode: 200,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item)
        };
    } catch (error) {
        console.error('Error updating item:', error);
        return {
            statusCode: 400,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ error: 'Invalid request' })
        };
    }
};

exports.deleteItem = async (event, context) => {
    const itemId = parseInt(event.pathParameters?.id || '0');
    
    if (!itemsDB[itemId]) {
        return {
            statusCode: 404,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ error: 'Item not found' })
        };
    }
    
    delete itemsDB[itemId];
    console.log(` + "`Deleted item: ${itemId}`" + `);
    
    return {
        statusCode: 204,
        headers: { 'Content-Type': 'application/json' },
        body: ''
    };
};
`
	
	default:
		return getNodeMainTemplate()
	}
}

// Feature-specific file generators
func getPythonDockerfile(runtime string) string {
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
`, strings.TrimPrefix(runtime, "python"))
}

func getNodeDockerfile(runtime string) string {
	return fmt.Sprintf(`FROM node:%s-slim

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 8000

CMD ["node", "main.js"]
`, strings.TrimPrefix(runtime, "nodejs"))
}

func getDockerignore() string {
	return `node_modules
npm-debug.log
.git
.gitignore
README.md
.env
.nyc_output
coverage
.pytest_cache
__pycache__
*.pyc
.venv
venv
`
}

func getGitHubActionsWorkflow() string {
	return `name: Deploy API

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.9'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
    
    - name: Run tests
      run: python -m pytest tests/

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to API-Direct
      run: |
        # Install API-Direct CLI
        # apidirect deploy
      env:
        API_DIRECT_TOKEN: ${{ secrets.API_DIRECT_TOKEN }}
`
}

func getAPIDocumentation(template APITemplate) string {
	return fmt.Sprintf(`# %s API Documentation

## Overview
%s

## Endpoints

### Base URL
` + "`https://api.yourdomain.com`" + `

### Authentication
All endpoints require an API key passed in the ` + "`X-API-Key`" + ` header.

## Template Features
%s

## Error Responses

All endpoints return errors in the following format:

` + "```json" + `
{
  "error": "Error description",
  "code": "ERROR_CODE"
}
` + "```" + `

## Rate Limiting
- 1000 requests per hour per API key
- Rate limit headers included in all responses

## Support
For API support, contact: support@yourdomain.com
`, template.Name, template.Description, strings.Join(template.Features, "\n- "))
}

func getPythonTemplateRequirements(template APITemplate, features []string) string {
	requirements := `# Core dependencies
`
	
	switch template.ID {
	case "crud-database":
		requirements += `psycopg2-binary==2.9.7
sqlalchemy==2.0.21
`
	case "ml-model-serving":
		requirements += `scikit-learn==1.3.0
numpy==1.24.3
pandas==2.0.3
`
	case "webhook-receiver":
		requirements += `cryptography==41.0.4
`
	}
	
	for _, feature := range features {
		switch feature {
		case "API documentation generation":
			requirements += `fastapi==0.103.1
uvicorn==0.23.2
`
		}
	}
	
	return requirements
}

func getNodeTemplatePackage(apiName string, template APITemplate, features []string) string {
	deps := `{}`
	devDeps := `{
    "jest": "^29.5.0"
  }`
	
	switch template.ID {
	case "crud-database":
		deps = `{
    "pg": "^8.11.3"
  }`
	}
	
	for _, feature := range features {
		switch feature {
		case "API documentation generation":
			// Add swagger dependencies
		}
	}
	
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "%s",
  "main": "main.js",
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "start": "node main.js"
  },
  "keywords": ["api", "serverless", "api-direct"],
  "author": "",
  "license": "MIT",
  "dependencies": %s,
  "devDependencies": %s
}`, apiName, template.Description, deps, devDeps)
}

func getPythonTemplateTests(template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return `"""
Tests for CRUD API handlers
"""
import json
import unittest
from main import list_items, create_item, get_item, update_item, delete_item


class TestCRUDAPI(unittest.TestCase):
    
    def test_list_items_empty(self):
        event = {}
        context = {}
        
        response = list_items(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertEqual(body['total'], 0)
    
    def test_create_item(self):
        event = {
            'body': json.dumps({
                'name': 'Test Item',
                'description': 'A test item'
            })
        }
        context = {}
        
        response = create_item(event, context)
        
        self.assertEqual(response['statusCode'], 201)
        body = json.loads(response['body'])
        self.assertEqual(body['name'], 'Test Item')
        self.assertIn('id', body)


if __name__ == '__main__':
    unittest.main()
`
	
	default:
		return getPythonTestTemplate()
	}
}

func getNodeTemplateTests(template APITemplate) string {
	switch template.ID {
	case "crud-database":
		return `/**
 * Tests for CRUD API handlers
 */
const { listItems, createItem, getItem, updateItem, deleteItem } = require('../main');

describe('CRUD API', () => {
    describe('listItems', () => {
        it('should return empty list initially', async () => {
            const event = {};
            const context = {};
            
            const response = await listItems(event, context);
            
            expect(response.statusCode).toBe(200);
            const body = JSON.parse(response.body);
            expect(body.total).toBe(0);
        });
    });
    
    describe('createItem', () => {
        it('should create a new item', async () => {
            const event = {
                body: JSON.stringify({
                    name: 'Test Item',
                    description: 'A test item'
                })
            };
            const context = {};
            
            const response = await createItem(event, context);
            
            expect(response.statusCode).toBe(201);
            const body = JSON.parse(response.body);
            expect(body.name).toBe('Test Item');
            expect(body.id).toBeDefined();
        });
    });
});
`
	
	default:
		return getNodeTestTemplate()
	}
}

func getTemplateReadme(apiName, language string, template APITemplate, features []string) string {
	featuresSection := ""
	if len(features) > 0 {
		featuresSection = fmt.Sprintf(`

## Features Included
%s`, strings.Join(features, "\n- "))
	}
	
	return fmt.Sprintf(`# %s

%s

**Template:** %s  
**Runtime:** %s  
**Category:** %s

%s%s

## Template Features
%s

## Getting Started

1. **Install dependencies**:
   ` + "```bash" + `
   %s
   ` + "```" + `

2. **Configure your API**:
   Edit ` + "`apidirect.yaml`" + ` to customize endpoints and settings.

3. **Implement your logic**:
   Edit the main implementation file to add your business logic.

4. **Test locally**:
   ` + "```bash" + `
   apidirect run
   ` + "```" + `

5. **Deploy to API-Direct**:
   ` + "```bash" + `
   apidirect deploy
   ` + "```" + `

6. **Publish to marketplace**:
   ` + "```bash" + `
   apidirect publish %s
   ` + "```" + `

## Need Help?

- Documentation: https://docs.api-direct.io
- Support: support@api-direct.io
`, 
		apiName, 
		template.Description,
		template.Name,
		language,
		template.Category,
		featuresSection,
		strings.Join(template.Features, "\n- "),
		getInstallCommand(language),
		apiName)
}

func getInstallCommand(language string) string {
	if strings.Contains(language, "Node") {
		return "npm install"
	}
	return "pip install -r requirements.txt"
}
