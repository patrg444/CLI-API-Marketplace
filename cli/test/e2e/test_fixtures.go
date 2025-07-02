package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFixtures contains sample projects for testing
type TestFixtures struct {
	FastAPISimple   *APIProject
	FastAPIComplex  *APIProject
	ExpressSimple   *APIProject
	GoGinAPI        *APIProject
	RubyRailsAPI    *APIProject
}

// APIProject represents a test API project
type APIProject struct {
	Name     string
	Runtime  string
	Port     int
	Files    map[string]string
	Manifest map[string]interface{}
}

// GetTestFixtures returns all test fixtures
func GetTestFixtures() *TestFixtures {
	return &TestFixtures{
		FastAPISimple:  getFastAPISimpleFixture(),
		FastAPIComplex: getFastAPIComplexFixture(),
		ExpressSimple:  getExpressSimpleFixture(),
		GoGinAPI:       getGoGinFixture(),
		RubyRailsAPI:   getRubyRailsFixture(),
	}
}

// getFastAPISimpleFixture returns a simple FastAPI project
func getFastAPISimpleFixture() *APIProject {
	return &APIProject{
		Name:    "fastapi-simple",
		Runtime: "python3.9",
		Port:    8000,
		Files: map[string]string{
			"main.py": `from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Hello World"}

@app.get("/health")
def health_check():
    return {"status": "healthy"}
`,
			"requirements.txt": `fastapi==0.104.1
uvicorn==0.24.0
`,
			"Dockerfile": `FROM python:3.9-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
`,
		},
		Manifest: map[string]interface{}{
			"version":       "1.0",
			"name":          "fastapi-simple",
			"runtime":       "python3.9",
			"port":          8000,
			"start_command": "uvicorn main:app --host 0.0.0.0 --port 8000",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
			},
		},
	}
}

// getFastAPIComplexFixture returns a complex FastAPI project
func getFastAPIComplexFixture() *APIProject {
	return &APIProject{
		Name:    "fastapi-complex",
		Runtime: "python3.9",
		Port:    8080,
		Files: map[string]string{
			"main.py": `from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.orm import Session
from typing import List
import os

from database import get_db
from models import User, Product
from schemas import UserCreate, UserResponse, ProductResponse

app = FastAPI(title="Complex API", version="2.0.0")
security = HTTPBearer()

@app.get("/")
def read_root():
    return {"message": "Complex FastAPI Example", "version": "2.0.0"}

@app.get("/health")
def health_check():
    return {
        "status": "healthy",
        "database": "connected",
        "cache": "connected"
    }

@app.post("/api/v1/users", response_model=UserResponse)
def create_user(user: UserCreate, db: Session = Depends(get_db)):
    db_user = User(**user.dict())
    db.add(db_user)
    db.commit()
    return db_user

@app.get("/api/v1/products", response_model=List[ProductResponse])
def list_products(
    skip: int = 0,
    limit: int = 100,
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
):
    products = db.query(Product).offset(skip).limit(limit).all()
    return products

@app.get("/api/v1/products/{product_id}")
def get_product(product_id: int, db: Session = Depends(get_db)):
    product = db.query(Product).filter(Product.id == product_id).first()
    if not product:
        raise HTTPException(status_code=404, detail="Product not found")
    return product
`,
			"database.py": `from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
import os

SQLALCHEMY_DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./test.db")

engine = create_engine(SQLALCHEMY_DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
`,
			"models.py": `from sqlalchemy import Column, Integer, String, Float, DateTime
from database import Base
import datetime

class User(Base):
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, index=True)
    username = Column(String, unique=True, index=True)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)

class Product(Base):
    __tablename__ = "products"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True)
    description = Column(String)
    price = Column(Float)
    stock = Column(Integer)
`,
			"schemas.py": `from pydantic import BaseModel
from datetime import datetime
from typing import Optional

class UserCreate(BaseModel):
    email: str
    username: str

class UserResponse(BaseModel):
    id: int
    email: str
    username: str
    created_at: datetime
    
    class Config:
        orm_mode = True

class ProductResponse(BaseModel):
    id: int
    name: str
    description: Optional[str]
    price: float
    stock: int
    
    class Config:
        orm_mode = True
`,
			"requirements.txt": `fastapi==0.104.1
uvicorn[standard]==0.24.0
sqlalchemy==2.0.23
pydantic==2.5.0
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
`,
			"Dockerfile": `FROM python:3.9-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
`,
		},
		Manifest: map[string]interface{}{
			"version":       "1.0",
			"name":          "fastapi-complex",
			"runtime":       "python3.9",
			"port":          8080,
			"start_command": "uvicorn main:app --host 0.0.0.0 --port 8080",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
				"POST /api/v1/users",
				"GET /api/v1/products",
				"GET /api/v1/products/{product_id}",
			},
			"env": map[string]interface{}{
				"required": []string{"DATABASE_URL", "JWT_SECRET"},
				"optional": []string{"REDIS_URL", "LOG_LEVEL"},
			},
			"scaling": map[string]interface{}{
				"min":        2,
				"max":        20,
				"target_cpu": 70,
			},
			"resources": map[string]interface{}{
				"memory": "1Gi",
				"cpu":    "500m",
			},
		},
	}
}

// getExpressSimpleFixture returns a simple Express.js project
func getExpressSimpleFixture() *APIProject {
	return &APIProject{
		Name:    "express-simple",
		Runtime: "node18",
		Port:    3000,
		Files: map[string]string{
			"index.js": `const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Express!' });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

app.get('/api/users', (req, res) => {
  res.json([
    { id: 1, name: 'John Doe' },
    { id: 2, name: 'Jane Smith' }
  ]);
});

app.post('/api/users', (req, res) => {
  const { name } = req.body;
  res.status(201).json({ id: 3, name });
});

app.listen(port, () => {
  console.log(`+"`Server running on port ${port}`"+`);
});
`,
			"package.json": `{
  "name": "express-simple",
  "version": "1.0.0",
  "description": "Simple Express API",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "devDependencies": {
    "nodemon": "^3.0.1"
  }
}
`,
			"Dockerfile": `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 3000
CMD ["node", "index.js"]
`,
		},
		Manifest: map[string]interface{}{
			"version":       "1.0",
			"name":          "express-simple",
			"runtime":       "node18",
			"port":          3000,
			"start_command": "node index.js",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
				"GET /api/users",
				"POST /api/users",
			},
		},
	}
}

// getGoGinFixture returns a Go Gin project
func getGoGinFixture() *APIProject {
	return &APIProject{
		Name:    "go-gin-api",
		Runtime: "go1.21",
		Port:    8080,
		Files: map[string]string{
			"main.go": `package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID    int    `+"`json:\"id\"`"+`
	Name  string `+"`json:\"name\"`"+`
	Email string `+"`json:\"email\"`"+`
}

func main() {
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from Go Gin!",
		})
	})
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	
	r.GET("/api/users", func(c *gin.Context) {
		users := []User{
			{ID: 1, Name: "John Doe", Email: "john@example.com"},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		}
		c.JSON(http.StatusOK, users)
	})
	
	r.Run(":8080")
}
`,
			"go.mod": `module go-gin-api

go 1.21

require github.com/gin-gonic/gin v1.9.1
`,
			"Dockerfile": `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
`,
		},
		Manifest: map[string]interface{}{
			"version":       "1.0",
			"name":          "go-gin-api",
			"runtime":       "go1.21",
			"port":          8080,
			"start_command": "./main",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
				"GET /api/users",
			},
		},
	}
}

// getRubyRailsFixture returns a Ruby on Rails API project
func getRubyRailsFixture() *APIProject {
	return &APIProject{
		Name:    "rails-api",
		Runtime: "ruby3.2",
		Port:    3000,
		Files: map[string]string{
			"config.ru": `require_relative "config/environment"
run Rails.application
Rails.application.load_server
`,
			"Gemfile": `source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "3.2.0"

gem "rails", "~> 7.0.0"
gem "puma", "~> 5.0"
gem "bootsnap", ">= 1.4.4", require: false
gem "rack-cors"

group :development, :test do
  gem "byebug", platforms: [:mri, :mingw, :x64_mingw]
end
`,
			"Dockerfile": `FROM ruby:3.2-slim
RUN apt-get update -qq && apt-get install -y build-essential libpq-dev nodejs
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle install
COPY . .
EXPOSE 3000
CMD ["rails", "server", "-b", "0.0.0.0"]
`,
		},
		Manifest: map[string]interface{}{
			"version":       "1.0",
			"name":          "rails-api",
			"runtime":       "ruby3.2",
			"port":          3000,
			"start_command": "rails server -b 0.0.0.0",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
			},
		},
	}
}

// CreateFixtureProject creates a test project from a fixture
func CreateFixtureProject(t *testing.T, dir string, fixture *APIProject) {
	projectDir := filepath.Join(dir, fixture.Name)
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	
	// Create all files
	for filename, content := range fixture.Files {
		filePath := filepath.Join(projectDir, filename)
		require.NoError(t, ioutil.WriteFile(filePath, []byte(content), 0644))
	}
	
	// Create manifest
	manifestBytes, err := json.MarshalIndent(fixture.Manifest, "", "  ")
	require.NoError(t, err)
	
	manifestPath := filepath.Join(projectDir, "apidirect.manifest.json")
	require.NoError(t, ioutil.WriteFile(manifestPath, manifestBytes, 0644))
}

// GetSampleManifests returns various sample manifests for testing
func GetSampleManifests() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"minimal": {
			"version":       "1.0",
			"name":          "minimal-api",
			"runtime":       "python3.9",
			"port":          8000,
			"start_command": "python app.py",
			"health_check":  "/",
		},
		"full-featured": {
			"version":       "1.0",
			"name":          "full-api",
			"runtime":       "python3.9",
			"port":          8080,
			"start_command": "uvicorn main:app --host 0.0.0.0 --port 8080",
			"health_check":  "/health",
			"endpoints": []string{
				"GET /",
				"GET /health",
				"GET /api/v1/users",
				"POST /api/v1/users",
				"GET /api/v1/products",
				"GET /api/v1/orders",
			},
			"env": map[string]interface{}{
				"required": []string{"DATABASE_URL", "REDIS_URL", "JWT_SECRET"},
				"optional": []string{"LOG_LEVEL", "DEBUG", "SENTRY_DSN"},
			},
			"scaling": map[string]interface{}{
				"min":        3,
				"max":        50,
				"target_cpu": 60,
			},
			"resources": map[string]interface{}{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
			"files": map[string]interface{}{
				"main":         "main.py",
				"requirements": "requirements.txt",
				"dockerfile":   "Dockerfile",
			},
			"metadata": map[string]interface{}{
				"description": "Full-featured production API",
				"version":     "2.1.0",
				"team":        "Platform Team",
			},
		},
		"ml-api": {
			"version":       "1.0",
			"name":          "ml-prediction-api",
			"runtime":       "python3.9",
			"port":          8000,
			"start_command": "uvicorn app:app --host 0.0.0.0 --port 8000",
			"health_check":  "/health",
			"endpoints": []string{
				"POST /predict",
				"GET /model/info",
				"POST /batch/predict",
			},
			"env": map[string]interface{}{
				"required": []string{"MODEL_PATH", "AWS_S3_BUCKET"},
				"optional": []string{"GPU_ENABLED", "BATCH_SIZE"},
			},
			"scaling": map[string]interface{}{
				"min":        1,
				"max":        10,
				"target_cpu": 80,
			},
			"resources": map[string]interface{}{
				"memory": "4Gi",
				"cpu":    "2000m",
			},
		},
	}
}

// TestManifestValidation tests manifest validation with various fixtures
func TestManifestValidation(t *testing.T) {
	manifests := GetSampleManifests()
	
	for name, manifest := range manifests {
		t.Run(fmt.Sprintf("Validate %s manifest", name), func(t *testing.T) {
			// Validate required fields
			assert.NotEmpty(t, manifest["version"])
			assert.NotEmpty(t, manifest["name"])
			assert.NotEmpty(t, manifest["runtime"])
			assert.NotEmpty(t, manifest["port"])
			assert.NotEmpty(t, manifest["start_command"])
			assert.NotEmpty(t, manifest["health_check"])
		})
	}
}