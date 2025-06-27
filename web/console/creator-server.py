#!/usr/bin/env python3
"""
Server for API-Direct Creator Portal
"""

from flask import Flask, send_from_directory
import os

app = Flask(__name__)

@app.route('/')
def creator_portal():
    """Serve the Creator Portal dashboard"""
    return send_from_directory('.', 'creator-portal.html')

@app.route('/health')
def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "api-direct-creator-portal"}

if __name__ == '__main__':
    print("🎨 API-Direct Creator Portal")
    print("📍 Running at: http://localhost:3003")
    print("🔗 Landing page: http://localhost:3000")
    print("📚 Docs: http://localhost:3001") 
    print("🖥️  Legacy console: http://localhost:3002")
    print("💼 Complete Creator Portal ready for business!")
    app.run(host='0.0.0.0', port=3003, debug=True)