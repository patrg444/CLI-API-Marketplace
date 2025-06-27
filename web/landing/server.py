#!/usr/bin/env python3
"""
Simple server for API-Direct landing page
"""

from flask import Flask, send_from_directory, render_template_string
import os

app = Flask(__name__)

@app.route('/')
def landing_page():
    """Serve the landing page"""
    return send_from_directory('.', 'index.html')

@app.route('/health')
def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "api-direct-landing"}

if __name__ == '__main__':
    print("🌐 API-Direct Landing Page Server")
    print("📍 Running at: http://localhost:3000")
    print("🚀 Ready for viral launch!")
    app.run(host='0.0.0.0', port=3000, debug=True)