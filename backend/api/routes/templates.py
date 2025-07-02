from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Any
from datetime import datetime
from pydantic import BaseModel
import os
import shutil
import tempfile
import subprocess
import json
import zipfile
from pathlib import Path

from backend.api.auth.auth import get_current_user
from backend.api.database import get_db, User, API

router = APIRouter()

# Template definitions
TEMPLATES = {
    "express-api": {
        "name": "Express REST API",
        "description": "Node.js Express API with MongoDB",
        "category": "web",
        "language": "javascript",
        "framework": "express",
        "features": ["RESTful endpoints", "MongoDB integration", "JWT auth", "Rate limiting"],
        "icon": "fab fa-node-js",
        "color": "green"
    },
    "flask-api": {
        "name": "Flask REST API",
        "description": "Python Flask API with PostgreSQL",
        "category": "web",
        "language": "python",
        "framework": "flask",
        "features": ["RESTful endpoints", "PostgreSQL", "SQLAlchemy ORM", "JWT auth"],
        "icon": "fab fa-python",
        "color": "blue"
    },
    "fastapi": {
        "name": "FastAPI",
        "description": "Modern Python API with automatic docs",
        "category": "web",
        "language": "python",
        "framework": "fastapi",
        "features": ["Auto documentation", "Type hints", "Async support", "High performance"],
        "icon": "fas fa-bolt",
        "color": "teal"
    },
    "gpt-wrapper": {
        "name": "GPT API Wrapper",
        "description": "OpenAI GPT wrapper with rate limiting",
        "category": "ai",
        "language": "python",
        "framework": "fastapi",
        "features": ["GPT-3/4 integration", "Token management", "Rate limiting", "Caching"],
        "icon": "fas fa-brain",
        "color": "purple"
    },
    "image-processor": {
        "name": "Image Processing API",
        "description": "Image manipulation and analysis API",
        "category": "ai",
        "language": "python",
        "framework": "fastapi",
        "features": ["Image resizing", "Format conversion", "AI enhancement", "Batch processing"],
        "icon": "fas fa-image",
        "color": "indigo"
    },
    "sentiment-analyzer": {
        "name": "Sentiment Analysis API",
        "description": "Text sentiment analysis using ML",
        "category": "ai",
        "language": "python",
        "framework": "fastapi",
        "features": ["Multi-language support", "Emotion detection", "Batch processing", "Real-time analysis"],
        "icon": "fas fa-comment-dots",
        "color": "pink"
    },
    "webhook-relay": {
        "name": "Webhook Relay Service",
        "description": "Webhook forwarding and transformation",
        "category": "tools",
        "language": "javascript",
        "framework": "express",
        "features": ["Webhook forwarding", "Request transformation", "Retry logic", "Event filtering"],
        "icon": "fas fa-exchange-alt",
        "color": "orange"
    },
    "graphql-api": {
        "name": "GraphQL API",
        "description": "GraphQL API with Apollo Server",
        "category": "web",
        "language": "javascript",
        "framework": "apollo",
        "features": ["GraphQL schema", "Resolvers", "Subscriptions", "DataLoader"],
        "icon": "fas fa-project-diagram",
        "color": "red"
    }
}

# Pydantic models
class TemplateInfo(BaseModel):
    id: str
    name: str
    description: str
    category: str
    language: str
    framework: str
    features: List[str]
    icon: str
    color: str

class TemplateScaffold(BaseModel):
    template_id: str
    api_name: str
    configuration: Optional[Dict[str, Any]] = {}

class ScaffoldResponse(BaseModel):
    api_id: str
    api_name: str
    template_id: str
    status: str
    created_at: datetime
    repository_url: Optional[str] = None

@router.get("/templates", response_model=List[TemplateInfo])
async def list_templates(
    category: Optional[str] = None,
    language: Optional[str] = None,
    current_user: User = Depends(get_current_user)
):
    """List available API templates"""
    templates = []
    
    for template_id, template_data in TEMPLATES.items():
        # Filter by category if specified
        if category and template_data["category"] != category:
            continue
        
        # Filter by language if specified
        if language and template_data["language"] != language:
            continue
        
        templates.append(TemplateInfo(
            id=template_id,
            **template_data
        ))
    
    return templates

@router.get("/templates/{template_id}", response_model=TemplateInfo)
async def get_template_details(
    template_id: str,
    current_user: User = Depends(get_current_user)
):
    """Get detailed information about a specific template"""
    if template_id not in TEMPLATES:
        raise HTTPException(status_code=404, detail="Template not found")
    
    return TemplateInfo(
        id=template_id,
        **TEMPLATES[template_id]
    )

@router.post("/templates/scaffold", response_model=ScaffoldResponse)
async def scaffold_from_template(
    scaffold_data: TemplateScaffold,
    background_tasks: BackgroundTasks,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Create a new API from a template"""
    # Validate template
    if scaffold_data.template_id not in TEMPLATES:
        raise HTTPException(status_code=404, detail="Template not found")
    
    template = TEMPLATES[scaffold_data.template_id]
    
    # Create API record
    new_api = API(
        owner_id=current_user.id,
        user_id=current_user.id,
        name=scaffold_data.api_name,
        description=f"API created from {template['name']} template",
        deployment_type="hosted",
        status="building",
        template_id=scaffold_data.template_id,
        runtime_config={
            "language": template["language"],
            "framework": template["framework"],
            "template": scaffold_data.template_id
        }
    )
    
    db.add(new_api)
    db.commit()
    db.refresh(new_api)
    
    # Start scaffolding in background
    background_tasks.add_task(
        scaffold_template,
        api_id=str(new_api.id),
        template_id=scaffold_data.template_id,
        api_name=scaffold_data.api_name,
        configuration=scaffold_data.configuration,
        db=db
    )
    
    return ScaffoldResponse(
        api_id=str(new_api.id),
        api_name=scaffold_data.api_name,
        template_id=scaffold_data.template_id,
        status="building",
        created_at=new_api.created_at
    )

async def scaffold_template(
    api_id: str,
    template_id: str,
    api_name: str,
    configuration: Dict[str, Any],
    db: Session
):
    """Background task to scaffold an API from a template"""
    try:
        # Create temporary directory
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir) / api_name
            
            # Generate project structure based on template
            if template_id == "express-api":
                await generate_express_api(project_path, api_name, configuration)
            elif template_id == "flask-api":
                await generate_flask_api(project_path, api_name, configuration)
            elif template_id == "fastapi":
                await generate_fastapi_api(project_path, api_name, configuration)
            elif template_id == "gpt-wrapper":
                await generate_gpt_wrapper(project_path, api_name, configuration)
            elif template_id == "image-processor":
                await generate_image_processor(project_path, api_name, configuration)
            else:
                # Use generic template
                await generate_generic_template(project_path, api_name, template_id, configuration)
            
            # Create zip file
            zip_path = Path(temp_dir) / f"{api_name}.zip"
            shutil.make_archive(str(zip_path.with_suffix('')), 'zip', project_path)
            
            # Update API status
            api = db.query(API).filter(API.id == api_id).first()
            if api:
                api.status = "ready"
                # Store the generated code location (in real implementation, upload to S3)
                api.runtime_config["generated_code"] = f"/generated/{api_id}/{api_name}.zip"
                db.commit()
    
    except Exception as e:
        # Update API status to error
        api = db.query(API).filter(API.id == api_id).first()
        if api:
            api.status = "error"
            api.runtime_config["error"] = str(e)
            db.commit()

async def generate_express_api(path: Path, name: str, config: Dict[str, Any]):
    """Generate Express.js API template"""
    path.mkdir(parents=True, exist_ok=True)
    
    # Package.json
    package_json = {
        "name": name.lower().replace(" ", "-"),
        "version": "1.0.0",
        "description": f"{name} API",
        "main": "server.js",
        "scripts": {
            "start": "node server.js",
            "dev": "nodemon server.js"
        },
        "dependencies": {
            "express": "^4.18.2",
            "cors": "^2.8.5",
            "helmet": "^7.0.0",
            "express-rate-limit": "^6.7.0",
            "jsonwebtoken": "^9.0.0",
            "bcryptjs": "^2.4.3",
            "mongoose": "^7.0.3",
            "dotenv": "^16.0.3"
        },
        "devDependencies": {
            "nodemon": "^2.0.22"
        }
    }
    
    with open(path / "package.json", "w") as f:
        json.dump(package_json, f, indent=2)
    
    # Server.js
    server_code = '''const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json());

// Rate limiting
const limiter = rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100 // limit each IP to 100 requests per windowMs
});
app.use('/api/', limiter);

// Routes
app.get('/health', (req, res) => {
    res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

app.get('/api/hello', (req, res) => {
    res.json({ message: 'Hello from ' + process.env.API_NAME || 'API' });
});

// Error handling
app.use((err, req, res, next) => {
    console.error(err.stack);
    res.status(500).json({ error: 'Something went wrong!' });
});

app.listen(PORT, () => {
    console.log(`Server running on port ${PORT}`);
});
'''
    
    with open(path / "server.js", "w") as f:
        f.write(server_code)
    
    # .env file
    env_content = f'''API_NAME={name}
PORT=3000
MONGODB_URI=mongodb://localhost:27017/{name.lower().replace(" ", "_")}
JWT_SECRET=your-secret-key-change-in-production
'''
    
    with open(path / ".env", "w") as f:
        f.write(env_content)
    
    # README
    readme_content = f'''# {name}

Generated from Express REST API template.

## Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Configure environment variables in `.env`

3. Run the server:
   ```bash
   npm start
   ```

## Features

- Express.js framework
- MongoDB integration
- JWT authentication ready
- Rate limiting
- CORS enabled
- Helmet security
'''
    
    with open(path / "README.md", "w") as f:
        f.write(readme_content)

async def generate_fastapi_api(path: Path, name: str, config: Dict[str, Any]):
    """Generate FastAPI template"""
    path.mkdir(parents=True, exist_ok=True)
    
    # requirements.txt
    requirements = '''fastapi==0.104.1
uvicorn==0.24.0
pydantic==2.4.2
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
sqlalchemy==2.0.23
asyncpg==0.29.0
alembic==1.12.1
python-dotenv==1.0.0
'''
    
    with open(path / "requirements.txt", "w") as f:
        f.write(requirements)
    
    # main.py
    main_code = f'''from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from datetime import datetime
import os
from dotenv import load_dotenv

load_dotenv()

app = FastAPI(
    title="{name}",
    description="API generated from FastAPI template",
    version="1.0.0"
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Models
class HealthResponse(BaseModel):
    status: str
    timestamp: datetime
    version: str

class MessageResponse(BaseModel):
    message: str

# Routes
@app.get("/health", response_model=HealthResponse)
async def health_check():
    return HealthResponse(
        status="healthy",
        timestamp=datetime.now(),
        version="1.0.0"
    )

@app.get("/", response_model=MessageResponse)
async def root():
    return MessageResponse(message=f"Welcome to {name}")

@app.get("/api/hello", response_model=MessageResponse)
async def hello():
    return MessageResponse(message=f"Hello from {name}!")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
'''
    
    with open(path / "main.py", "w") as f:
        f.write(main_code)
    
    # Dockerfile
    dockerfile = '''FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
'''
    
    with open(path / "Dockerfile", "w") as f:
        f.write(dockerfile)

async def generate_gpt_wrapper(path: Path, name: str, config: Dict[str, Any]):
    """Generate GPT API wrapper template"""
    path.mkdir(parents=True, exist_ok=True)
    
    # Enhanced GPT wrapper with caching and rate limiting
    main_code = '''from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Optional, List
import openai
import os
from dotenv import load_dotenv
import redis
import hashlib
import json
from datetime import datetime, timedelta

load_dotenv()

app = FastAPI(title="GPT API Wrapper", version="1.0.0")

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Redis for caching (optional)
try:
    redis_client = redis.from_url(os.getenv("REDIS_URL", "redis://localhost:6379"))
    CACHE_ENABLED = True
except:
    CACHE_ENABLED = False

# OpenAI setup
openai.api_key = os.getenv("OPENAI_API_KEY")

class CompletionRequest(BaseModel):
    prompt: str
    model: str = Field(default="gpt-3.5-turbo", description="GPT model to use")
    max_tokens: int = Field(default=150, ge=1, le=4000)
    temperature: float = Field(default=0.7, ge=0, le=2)
    top_p: float = Field(default=1.0, ge=0, le=1)
    cache: bool = Field(default=True, description="Enable response caching")

class CompletionResponse(BaseModel):
    text: str
    model: str
    usage: dict
    cached: bool = False

def get_cache_key(request: CompletionRequest) -> str:
    """Generate cache key from request parameters"""
    key_data = f"{request.prompt}:{request.model}:{request.max_tokens}:{request.temperature}:{request.top_p}"
    return hashlib.md5(key_data.encode()).hexdigest()

@app.post("/api/completion", response_model=CompletionResponse)
async def create_completion(request: CompletionRequest):
    """Generate text completion using GPT"""
    
    # Check cache if enabled
    if CACHE_ENABLED and request.cache:
        cache_key = get_cache_key(request)
        cached_response = redis_client.get(cache_key)
        if cached_response:
            data = json.loads(cached_response)
            return CompletionResponse(**data, cached=True)
    
    try:
        # Create completion
        response = openai.ChatCompletion.create(
            model=request.model,
            messages=[{"role": "user", "content": request.prompt}],
            max_tokens=request.max_tokens,
            temperature=request.temperature,
            top_p=request.top_p
        )
        
        result = CompletionResponse(
            text=response.choices[0].message.content,
            model=response.model,
            usage=response.usage
        )
        
        # Cache response if enabled
        if CACHE_ENABLED and request.cache:
            cache_key = get_cache_key(request)
            redis_client.setex(
                cache_key,
                3600,  # 1 hour TTL
                json.dumps(result.dict())
            )
        
        return result
        
    except openai.error.RateLimitError:
        raise HTTPException(status_code=429, detail="OpenAI rate limit exceeded")
    except openai.error.InvalidRequestError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/models")
async def list_models():
    """List available GPT models"""
    return {
        "models": [
            {"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo"},
            {"id": "gpt-4", "name": "GPT-4"},
            {"id": "gpt-4-turbo-preview", "name": "GPT-4 Turbo"}
        ]
    }

@app.get("/health")
async def health_check():
    return {"status": "healthy", "cache_enabled": CACHE_ENABLED}
'''
    
    with open(path / "main.py", "w") as f:
        f.write(main_code)
    
    # requirements.txt
    requirements = '''fastapi==0.104.1
uvicorn==0.24.0
openai==0.28.1
redis==5.0.1
python-dotenv==1.0.0
'''
    
    with open(path / "requirements.txt", "w") as f:
        f.write(requirements)
    
    # .env template
    env_template = '''OPENAI_API_KEY=your-openai-api-key
REDIS_URL=redis://localhost:6379
'''
    
    with open(path / ".env.example", "w") as f:
        f.write(env_template)

async def generate_flask_api(path: Path, name: str, config: Dict[str, Any]):
    """Generate Flask API template"""
    path.mkdir(parents=True, exist_ok=True)
    
    # app.py
    app_code = f'''from flask import Flask, jsonify, request
from flask_cors import CORS
from flask_sqlalchemy import SQLAlchemy
from datetime import datetime
import os
from dotenv import load_dotenv

load_dotenv()

app = Flask(__name__)
CORS(app)

# Configuration
app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DATABASE_URL', 'postgresql://localhost/{name.lower()}')
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.config['SECRET_KEY'] = os.getenv('SECRET_KEY', 'dev-secret-key')

db = SQLAlchemy(app)

# Models
class Item(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100), nullable=False)
    description = db.Column(db.Text)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)

# Routes
@app.route('/health')
def health_check():
    return jsonify({{'status': 'healthy', 'timestamp': datetime.utcnow().isoformat()}})

@app.route('/api/items', methods=['GET'])
def get_items():
    items = Item.query.all()
    return jsonify([{{
        'id': item.id,
        'name': item.name,
        'description': item.description,
        'created_at': item.created_at.isoformat()
    }} for item in items])

@app.route('/api/items', methods=['POST'])
def create_item():
    data = request.get_json()
    item = Item(
        name=data.get('name'),
        description=data.get('description')
    )
    db.session.add(item)
    db.session.commit()
    return jsonify({{
        'id': item.id,
        'name': item.name,
        'description': item.description,
        'created_at': item.created_at.isoformat()
    }}), 201

if __name__ == '__main__':
    with app.app_context():
        db.create_all()
    app.run(debug=True)
'''
    
    with open(path / "app.py", "w") as f:
        f.write(app_code)
    
    # requirements.txt
    requirements = '''Flask==3.0.0
Flask-CORS==4.0.0
Flask-SQLAlchemy==3.1.1
psycopg2-binary==2.9.9
python-dotenv==1.0.0
gunicorn==21.2.0
'''
    
    with open(path / "requirements.txt", "w") as f:
        f.write(requirements)

async def generate_image_processor(path: Path, name: str, config: Dict[str, Any]):
    """Generate image processing API template"""
    path.mkdir(parents=True, exist_ok=True)
    
    # main.py with image processing endpoints
    main_code = '''from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.responses import StreamingResponse
from PIL import Image
import io
from typing import Optional
import os

app = FastAPI(title="Image Processing API", version="1.0.0")

ALLOWED_FORMATS = {"JPEG", "PNG", "GIF", "BMP", "WEBP"}
MAX_SIZE = 10 * 1024 * 1024  # 10MB

@app.post("/api/resize")
async def resize_image(
    file: UploadFile = File(...),
    width: int = None,
    height: int = None,
    maintain_aspect_ratio: bool = True
):
    """Resize an image to specified dimensions"""
    
    # Validate file size
    contents = await file.read()
    if len(contents) > MAX_SIZE:
        raise HTTPException(status_code=413, detail="File too large")
    
    # Open image
    try:
        image = Image.open(io.BytesIO(contents))
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid image file")
    
    # Validate format
    if image.format not in ALLOWED_FORMATS:
        raise HTTPException(status_code=400, detail=f"Unsupported format: {image.format}")
    
    # Calculate dimensions
    if maintain_aspect_ratio:
        image.thumbnail((width or image.width, height or image.height), Image.Resampling.LANCZOS)
    else:
        if width and height:
            image = image.resize((width, height), Image.Resampling.LANCZOS)
    
    # Save to buffer
    output = io.BytesIO()
    image.save(output, format=image.format)
    output.seek(0)
    
    return StreamingResponse(output, media_type=f"image/{image.format.lower()}")

@app.post("/api/convert")
async def convert_format(
    file: UploadFile = File(...),
    format: str = "PNG"
):
    """Convert image to different format"""
    format = format.upper()
    if format not in ALLOWED_FORMATS:
        raise HTTPException(status_code=400, detail=f"Unsupported format: {format}")
    
    contents = await file.read()
    image = Image.open(io.BytesIO(contents))
    
    output = io.BytesIO()
    # Handle transparency for JPEG
    if format == "JPEG" and image.mode in ("RGBA", "LA", "P"):
        rgb_image = Image.new("RGB", image.size, (255, 255, 255))
        rgb_image.paste(image, mask=image.split()[-1] if image.mode == "RGBA" else None)
        image = rgb_image
    
    image.save(output, format=format)
    output.seek(0)
    
    return StreamingResponse(output, media_type=f"image/{format.lower()}")

@app.post("/api/enhance")
async def enhance_image(
    file: UploadFile = File(...),
    brightness: float = 1.0,
    contrast: float = 1.0,
    sharpness: float = 1.0
):
    """Enhance image with adjustments"""
    from PIL import ImageEnhance
    
    contents = await file.read()
    image = Image.open(io.BytesIO(contents))
    
    # Apply enhancements
    if brightness != 1.0:
        enhancer = ImageEnhance.Brightness(image)
        image = enhancer.enhance(brightness)
    
    if contrast != 1.0:
        enhancer = ImageEnhance.Contrast(image)
        image = enhancer.enhance(contrast)
    
    if sharpness != 1.0:
        enhancer = ImageEnhance.Sharpness(image)
        image = enhancer.enhance(sharpness)
    
    output = io.BytesIO()
    image.save(output, format=image.format)
    output.seek(0)
    
    return StreamingResponse(output, media_type=f"image/{image.format.lower()}")

@app.get("/health")
async def health_check():
    return {"status": "healthy", "supported_formats": list(ALLOWED_FORMATS)}
'''
    
    with open(path / "main.py", "w") as f:
        f.write(main_code)
    
    # requirements.txt
    requirements = '''fastapi==0.104.1
uvicorn==0.24.0
pillow==10.1.0
python-multipart==0.0.6
'''
    
    with open(path / "requirements.txt", "w") as f:
        f.write(requirements)

async def generate_generic_template(path: Path, name: str, template_id: str, config: Dict[str, Any]):
    """Generate a generic template for other template types"""
    path.mkdir(parents=True, exist_ok=True)
    
    # Basic structure based on template
    template = TEMPLATES.get(template_id, {})
    
    if template.get("language") == "python":
        # Python-based template
        with open(path / "requirements.txt", "w") as f:
            f.write("fastapi==0.104.1\nuvicorn==0.24.0\n")
        
        with open(path / "main.py", "w") as f:
            f.write(f'''from fastapi import FastAPI

app = FastAPI(title="{name}")

@app.get("/")
async def root():
    return {{"message": "Hello from {name}"}}

@app.get("/health")
async def health():
    return {{"status": "healthy"}}
''')
    
    elif template.get("language") == "javascript":
        # JavaScript-based template
        package_json = {
            "name": name.lower().replace(" ", "-"),
            "version": "1.0.0",
            "main": "index.js",
            "scripts": {"start": "node index.js"},
            "dependencies": {"express": "^4.18.2"}
        }
        
        with open(path / "package.json", "w") as f:
            json.dump(package_json, f, indent=2)
        
        with open(path / "index.js", "w") as f:
            f.write(f'''const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {{
    res.json({{ message: 'Hello from {name}' }});
}});

app.get('/health', (req, res) => {{
    res.json({{ status: 'healthy' }});
}});

app.listen(PORT, () => {{
    console.log(`Server running on port ${{PORT}}`);
}});
''')

@router.get("/templates/preview/{template_id}")
async def preview_template_structure(
    template_id: str,
    current_user: User = Depends(get_current_user)
):
    """Preview the file structure that will be generated for a template"""
    if template_id not in TEMPLATES:
        raise HTTPException(status_code=404, detail="Template not found")
    
    # Return expected file structure
    structures = {
        "express-api": {
            "files": [
                {"path": "package.json", "type": "file"},
                {"path": "server.js", "type": "file"},
                {"path": ".env", "type": "file"},
                {"path": "README.md", "type": "file"},
                {"path": "routes/", "type": "directory"},
                {"path": "models/", "type": "directory"},
                {"path": "middleware/", "type": "directory"}
            ]
        },
        "fastapi": {
            "files": [
                {"path": "main.py", "type": "file"},
                {"path": "requirements.txt", "type": "file"},
                {"path": "Dockerfile", "type": "file"},
                {"path": ".env", "type": "file"},
                {"path": "README.md", "type": "file"},
                {"path": "app/", "type": "directory"},
                {"path": "tests/", "type": "directory"}
            ]
        },
        "gpt-wrapper": {
            "files": [
                {"path": "main.py", "type": "file"},
                {"path": "requirements.txt", "type": "file"},
                {"path": ".env.example", "type": "file"},
                {"path": "Dockerfile", "type": "file"},
                {"path": "README.md", "type": "file"},
                {"path": "cache/", "type": "directory"}
            ]
        }
    }
    
    return {
        "template_id": template_id,
        "structure": structures.get(template_id, {"files": []})
    }