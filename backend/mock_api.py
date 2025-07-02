#!/usr/bin/env python3
"""
Mock API server for API-Direct demonstration
Provides minimal endpoints to test frontend functionality
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
from datetime import datetime, timedelta
import uuid

class MockAPIHandler(BaseHTTPRequestHandler):
    def do_OPTIONS(self):
        """Handle CORS preflight requests"""
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type, Authorization')
        self.end_headers()

    def do_GET(self):
        """Handle GET requests"""
        # Add CORS headers
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        
        # Route handling
        if self.path == '/health':
            self.end_headers()
            self.wfile.write(json.dumps({"status": "healthy", "timestamp": datetime.now().isoformat()}).encode())
            
        elif self.path == '/api/dashboard/overview':
            self.end_headers()
            data = {
                "metrics": {
                    "total_apis": 3,
                    "active_apis": 2,
                    "total_calls": 15420,
                    "revenue": 245.67,
                    "period": "last_30_days"
                },
                "recent_deployments": [
                    {
                        "id": str(uuid.uuid4()),
                        "name": "GPT Wrapper API",
                        "status": "running",
                        "deployed_at": datetime.now().isoformat(),
                        "calls_today": 523
                    },
                    {
                        "id": str(uuid.uuid4()),
                        "name": "Weather Info API",
                        "status": "building",
                        "deployed_at": datetime.now().isoformat(),
                        "calls_today": 0
                    }
                ],
                "alerts": []
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/api/apis':
            self.end_headers()
            data = {
                "apis": [
                    {
                        "id": str(uuid.uuid4()),
                        "name": "GPT Wrapper API",
                        "status": "running",
                        "endpoint": "https://api.api-direct.io/gpt-wrapper",
                        "created_at": datetime.now().isoformat(),
                        "total_calls": 10234,
                        "deployment_type": "hosted"
                    },
                    {
                        "id": str(uuid.uuid4()),
                        "name": "Weather Info API",
                        "status": "building",
                        "endpoint": None,
                        "created_at": datetime.now().isoformat(),
                        "total_calls": 0,
                        "deployment_type": "hosted"
                    }
                ]
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/api/marketplace/listings':
            self.end_headers()
            data = {
                "listings": [
                    {
                        "id": str(uuid.uuid4()),
                        "name": "OpenAI GPT-4 Wrapper",
                        "description": "Simple, cost-effective access to GPT-4",
                        "category": "AI/ML",
                        "pricing_model": "per-call",
                        "price_per_call": 0.002,
                        "rating": 4.8,
                        "user_count": 156
                    },
                    {
                        "id": str(uuid.uuid4()),
                        "name": "Real-time Weather API",
                        "description": "Global weather data with 5-minute updates",
                        "category": "Data",
                        "pricing_model": "freemium",
                        "price_per_call": 0,
                        "rating": 4.5,
                        "user_count": 89
                    }
                ],
                "total": 2
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/api/creator/apis':
            self.end_headers()
            data = {
                "apis": [
                    {
                        "id": str(uuid.uuid4()),
                        "name": "Weather Forecast API",
                        "status": "active",
                        "subscribers": 89,
                        "monthly_revenue": 2670.00,
                        "success_rate": 99.8,
                        "created_at": (datetime.now() - timedelta(days=90)).isoformat()
                    },
                    {
                        "id": str(uuid.uuid4()),
                        "name": "Currency Exchange API",
                        "status": "active",
                        "subscribers": 67,
                        "monthly_revenue": 780.00,
                        "success_rate": 99.5,
                        "created_at": (datetime.now() - timedelta(days=30)).isoformat()
                    }
                ],
                "stats": {
                    "total_apis": 2,
                    "total_subscribers": 156,
                    "monthly_revenue": 3450.00,
                    "lifetime_earnings": 12450.00
                }
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/api/creator/analytics':
            self.end_headers()
            data = {
                "daily_calls": [
                    {"date": (datetime.now() - timedelta(days=i)).isoformat(), "calls": 1000 + i * 100}
                    for i in range(7)
                ],
                "revenue_trend": [
                    {"month": f"2024-{i:02d}", "revenue": 2500 + i * 150}
                    for i in range(1, 7)
                ]
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/docs':
            self.end_headers()
            html = """
            <html>
            <head><title>API Documentation</title></head>
            <body>
                <h1>API-Direct Mock API</h1>
                <h2>Available Endpoints:</h2>
                <ul>
                    <li>GET /health - Health check</li>
                    <li>GET /api/dashboard/overview - Dashboard data</li>
                    <li>GET /api/apis - List user APIs</li>
                    <li>GET /api/marketplace/listings - Marketplace listings</li>
                    <li>POST /auth/login - Mock login (returns test token)</li>
                    <li>POST /auth/register - Mock registration</li>
                </ul>
            </body>
            </html>
            """
            self.wfile.write(html.encode())
            
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())

    def do_POST(self):
        """Handle POST requests"""
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length)
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        
        if self.path == '/auth/login':
            # Mock login - return a test token
            data = {
                "access_token": "mock_token_" + str(uuid.uuid4()),
                "token_type": "bearer",
                "expires_in": 3600,
                "user": {
                    "id": str(uuid.uuid4()),
                    "name": "Test User",
                    "email": "test@example.com",
                    "created_at": datetime.now().isoformat(),
                    "isCreator": False
                }
            }
            self.wfile.write(json.dumps(data).encode())
            
        elif self.path == '/auth/register':
            # Mock registration
            data = {
                "access_token": "mock_token_" + str(uuid.uuid4()),
                "token_type": "bearer", 
                "expires_in": 3600,
                "user": {
                    "id": str(uuid.uuid4()),
                    "name": "New User",
                    "email": "new@example.com",
                    "created_at": datetime.now().isoformat()
                }
            }
            self.wfile.write(json.dumps(data).encode())
            
        else:
            self.wfile.write(json.dumps({"success": True}).encode())

    def log_message(self, format, *args):
        """Override to reduce console spam"""
        if '/health' not in args[0]:
            print(f"{self.address_string()} - {format % args}")

def run_server(port=8000):
    """Run the mock API server"""
    server_address = ('', port)
    httpd = HTTPServer(server_address, MockAPIHandler)
    print(f"🚀 Mock API server running on http://localhost:{port}")
    print(f"📚 API docs: http://localhost:{port}/docs")
    print("Press Ctrl+C to stop")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        httpd.shutdown()

if __name__ == '__main__':
    run_server()