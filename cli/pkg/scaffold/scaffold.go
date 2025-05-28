package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

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
