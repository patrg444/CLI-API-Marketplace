#!/usr/bin/env python3
"""
Simple server for API-Direct console application
"""

from flask import Flask, send_from_directory
import os

app = Flask(__name__)

@app.route('/')
def console_home():
    """Serve the console dashboard"""
    return send_from_directory('.', 'index.html')

@app.route('/health')
def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "api-direct-console"}

if __name__ == '__main__':
    print("🖥️  API-Direct Console Server")
    print("📍 Running at: http://localhost:3002")
    print("🔗 Landing page: http://localhost:3000")
    print("📚 Docs: http://localhost:3001")
    print("💻 Complete user journey ready!")
    app.run(host='0.0.0.0', port=3002, debug=True)