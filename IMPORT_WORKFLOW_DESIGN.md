# API-Direct Import Workflow Design

## Philosophy
**Auto-detection gets you 90% there. Human expertise handles the final 10%.**

## The Import Command Flow

### 1. Initial Scan & Best Guess
```bash
$ apidirect import ./my-api-project

🔍 Scanning project structure...
📦 Detected: Python 3.11 project
🚀 Found: FastAPI framework
📄 Located: requirements.txt
🔧 Discovered: 15 API endpoints
```

### 2. Interactive Confirmation (The Safety Net)
```bash
✅ Generated apidirect.yaml based on analysis:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# apidirect.yaml
# Auto-generated on 2024-03-15 14:23:00
# PLEASE REVIEW: These are our best guesses!

name: my-api-project
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

# Detected endpoints (found via @app.get decorators)
endpoints:
  - GET /
  - GET /health
  - POST /users
  - GET /users/{user_id}
  - PUT /users/{user_id}
  # ... (showing first 5 of 15 detected)

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

### 3. User Response Options

#### Option Y - Proceed
```bash
📝 Does this look correct? [Y/n/e]: y

✅ Great! Manifest saved to ./apidirect.yaml
🚀 Ready to deploy! Run: apidirect deploy
```

#### Option N - Reject & Edit
```bash
📝 Does this look correct? [Y/n/e]: n

📝 No problem! Let me help you fix it.

What would you like to do?
  1) Edit in your default editor
  2) Fix specific field
  3) Start over with different settings
  4) View documentation

Choose [1-4]: 1

🔧 Opening apidirect.yaml in VS Code...
💡 Tip: After editing, run 'apidirect validate' to check syntax
```

#### Option E - Quick Edit
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

### 4. The Validate Command
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

### 5. Dry Run Option (Extra Safety)
```bash
$ apidirect deploy --dry-run

🔍 Deployment Preview (DRY RUN - nothing will be deployed)

📦 Would create Docker container with:
   - Base image: python:3.11-slim
   - Install command: pip install -r requirements.txt
   - Start command: uvicorn app:application --host 0.0.0.0 --port 8000
   - Exposed port: 8000

🌐 Would deploy to:
   - URL: https://my-api-project-abc123.api-direct.io
   - Region: us-east-1
   - Auto-scaling: 1-10 instances

💰 Estimated cost: $0.20-0.50/month for 10K requests

Proceed with actual deployment? [y/N]: 
```

## Manifest Design Principles

### 1. Comments That Guide
```yaml
# apidirect.yaml
# Generated: 2024-03-15 14:23:00
# Docs: https://docs.api-direct.io/manifest

name: my-api-project
runtime: python3.11

# IMPORTANT: This is how we'll start your server in production
# Common examples:
#   FastAPI: uvicorn main:app --host 0.0.0.0 --port 8080
#   Flask: python app.py
#   Django: gunicorn myproject.wsgi:application
start_command: "uvicorn main:app --host 0.0.0.0 --port 8080"

# The port your application binds to (inside the container)
port: 8080
```

### 2. Progressive Disclosure
Start simple, allow complexity:
```yaml
# Minimal viable manifest
name: my-api
runtime: python3.11
start_command: "python app.py"
port: 8080

# Can grow to include:
scaling:
  min: 2
  max: 20
  
monitoring:
  alerts: true
  logs: "error"
```

### 3. Smart Defaults
```yaml
# User doesn't need to specify these unless different:
health_check: /health  # We'll try common endpoints
timeout: 30  # Sensible default
memory: 512Mi  # Good starting point
```

## Error Messages That Teach

### When Detection Fails
```bash
$ apidirect import ./my-project

⚠️  Couldn't auto-detect framework, but that's OK!

I found:
  - Python files (*.py)
  - A requirements.txt
  - No obvious web framework

Let's set this up together:

1) What command starts your server?
   Examples: 
   - python app.py
   - uvicorn main:app
   - flask run
   
   Your command: _
```

### When Validation Fails
```bash
$ apidirect validate

❌ Validation failed, but easy to fix:

1. Missing 'start_command' field
   ↳ Add: start_command: "python app.py"

2. Invalid port: "eight thousand"
   ↳ Change to number: port: 8000

📝 Would you like me to fix these? [Y/n]: 
```

## The Magic: It's Not Perfect, But It's Helpful

The system succeeds because:
1. **It shows its work** - Users see exactly what was detected
2. **It's easy to correct** - One command to edit, clear options
3. **It validates before deploy** - Catches errors early
4. **It teaches as it goes** - Examples and tips throughout
5. **It defaults to safety** - Always asks before proceeding

This creates a workflow where auto-detection + human verification = confidence in deployment.