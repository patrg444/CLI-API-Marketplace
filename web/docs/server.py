#!/usr/bin/env python3
"""
Simple server for API-Direct documentation site
"""

from flask import Flask, send_from_directory
import os

app = Flask(__name__)

@app.route('/')
def docs_home():
    """Serve the documentation homepage"""
    return send_from_directory('.', 'index.html')

@app.route('/health')
def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "api-direct-docs"}

if __name__ == '__main__':
    print("📚 API-Direct Documentation Server")
    print("📍 Running at: http://localhost:3001")
    print("🔗 Links to landing page at: http://localhost:3000")
    app.run(host='0.0.0.0', port=3001, debug=True)