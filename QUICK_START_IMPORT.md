# Quick Start: Import & Deploy Any API

This guide shows how to use the new import workflow to deploy any existing API project.

## Overview

The new `apidirect import` command allows you to deploy **any** API project without restructuring your code. It works by:

1. **Auto-detecting** your project configuration
2. **Generating** a simple manifest file
3. **Confirming** with you before proceeding
4. **Deploying** your API to the cloud

## Step-by-Step Example

### 1. Import Your Existing API

Navigate to your API project directory and run:

```bash
$ cd my-fastapi-project
$ apidirect import

🔍 Scanning project structure...
📦 Detected: Python 3.11 project
🚀 Found: FastAPI framework
📄 Located: requirements.txt
🔧 Discovered: 12 API endpoints

✅ Generated apidirect.yaml based on analysis:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# apidirect.yaml
# Auto-generated on 2024-03-15 14:23:00
# PLEASE REVIEW: These are our best guesses!

name: my-fastapi-project
runtime: python3.11

# How to start your server (PLEASE VERIFY!)
start_command: "uvicorn main:app --host 0.0.0.0 --port 8080"

# Where your server listens
port: 8080

# Your application files
files:
  main: ./main.py
  requirements: ./requirements.txt
  env_example: ./.env.example

# Detected endpoints
endpoints:
  - GET /
  - GET /health
  - POST /api/users
  - GET /api/users/{user_id}
  - PUT /api/users/{user_id}
  # ... (showing first 5 of 12 detected)

# Environment variables (found in .env.example)
env:
  required: [DATABASE_URL, SECRET_KEY]
  optional:
    LOG_LEVEL: info
    DEBUG: false

# Health check for monitoring
health_check: /health
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 Does this look correct? [Y/n/e]: 
```

### 2. Review and Correct if Needed

**Option A: Accept (if correct)**
```bash
📝 Does this look correct? [Y/n/e]: y

✅ Great! Manifest saved to ./apidirect.yaml
🚀 Ready to deploy! Run: apidirect deploy
```

**Option B: Quick Edit (for small changes)**
```bash
📝 Does this look correct? [Y/n/e]: e

🔧 Quick edit mode. What needs fixing?
  1) start_command (currently: uvicorn main:app --host 0.0.0.0 --port 8080)
  2) port (currently: 8080)
  3) main file (currently: ./main.py)
  4) Other field...

Choose field to edit [1-4]: 1

Enter new start_command: uvicorn app:application --host 0.0.0.0 --port 8000

✅ Updated! Here's the change:
- start_command: "uvicorn main:app --host 0.0.0.0 --port 8080"
+ start_command: "uvicorn app:application --host 0.0.0.0 --port 8000"

Anything else to fix? [y/N]: n

✅ Manifest saved! Run 'apidirect validate' to verify.
```

### 3. Validate Before Deploying

```bash
$ apidirect validate

🔍 Validating apidirect.yaml...

✅ YAML syntax: Valid
✅ Required fields: All present
✅ Port number: Valid (8000)
✅ Start command: Looks good
✅ File references: All files exist
⚠️  Warning: No Dockerfile found (will auto-generate during deploy)

📋 Validation Summary: READY TO DEPLOY
💡 Tip: Run 'apidirect deploy --dry-run' to preview deployment
```

### 4. Deploy Your API

```bash
$ apidirect deploy

🚀 Deploying 'my-fastapi-project' to hosted infrastructure
☁️  Using API-Direct hosted infrastructure...
📋 Configuration: python3.11 runtime, port 8000
🐳 Building container image...
   Generated Dockerfile from manifest
⬆️  Uploading code and building image...
🚀 Deploying to platform...
⏳ Waiting for deployment to be ready... ✓

✅ Deployment successful!
🌐 API URL: https://my-fastapi-project-abc123.api-direct.io
🆔 Deployment ID: dep_xyz789
📊 Dashboard: https://console.api-direct.io/apis/dep_xyz789

📍 Available endpoints:
   https://my-fastapi-project-abc123.api-direct.io/
   https://my-fastapi-project-abc123.api-direct.io/health
   https://my-fastapi-project-abc123.api-direct.io/api/users
   https://my-fastapi-project-abc123.api-direct.io/api/users/{user_id}
   ... and 8 more

🧪 Test your API:
   curl https://my-fastapi-project-abc123.api-direct.io/health

📝 Next steps:
   View logs:  apidirect logs my-fastapi-project
   Update:     apidirect deploy
   Scale:      apidirect scale my-fastapi-project --replicas 3

⚠️  Required environment variables:
   Set these in the dashboard: DATABASE_URL, SECRET_KEY
```

## Common Scenarios

### Express.js API
```yaml
name: my-express-api
runtime: node18
start_command: "node server.js"
port: 3000
files:
  main: ./server.js
  requirements: ./package.json
```

### Django API
```yaml
name: my-django-api  
runtime: python3.11
start_command: "gunicorn myproject.wsgi:application --bind 0.0.0.0:8000"
port: 8000
files:
  main: ./manage.py
  requirements: ./requirements.txt
```

### Custom Dockerfile
```yaml
name: my-custom-api
runtime: docker
start_command: "specified in Dockerfile"
port: 8080
files:
  dockerfile: ./Dockerfile
```

## AI Agent Usage

For AI coding assistants, the manifest is simple to generate:

```python
# AI agent can analyze code and generate:
manifest = {
    "name": project_name,
    "runtime": detect_runtime(),
    "start_command": find_start_command(),
    "port": find_port(),
    "files": {
        "main": find_main_file(),
        "requirements": find_deps_file()
    },
    "endpoints": extract_endpoints(),
    "env": parse_env_vars()
}

# Write to apidirect.yaml
write_yaml("apidirect.yaml", manifest)

# Then deploy
run_command("apidirect deploy")
```

## Troubleshooting

### Detection Failed?
```bash
⚠️  Couldn't auto-detect framework, but that's OK!

Let's set this up together:
1) What command starts your server?
   Examples: python app.py, npm start, go run main.go
   
   Your command: python api.py
```

### Validation Errors?
```bash
❌ Validation failed, but easy to fix:

1. Missing 'start_command' field
   ↳ Add: start_command: "python app.py"

📝 Would you like me to fix these? [Y/n]: y
```

### Need to Update After Deploy?
```bash
# Edit manifest
$ vim apidirect.yaml

# Validate changes  
$ apidirect validate

# Deploy update (automatic rolling update)
$ apidirect deploy
```

## Summary

The new import workflow makes it trivial to deploy any API:

1. **`apidirect import`** - Analyzes your project
2. **Review & tweak** - Ensure configuration is correct  
3. **`apidirect validate`** - Verify before deploying
4. **`apidirect deploy`** - Go live in seconds

No code changes required. No complex configuration. Just describe how to run your API, and we handle the rest!