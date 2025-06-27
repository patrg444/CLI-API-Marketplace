#!/usr/bin/env python3
"""
Flask application for API-Direct Creator Portal
Serves individual pages with proper routing and templating
"""

from flask import Flask, render_template_string, send_from_directory
import os

app = Flask(__name__)

# Configure template directory
template_dir = os.path.join(os.path.dirname(__file__), 'pages')

def load_template(template_name):
    """Load and render template with base template inheritance"""
    template_path = os.path.join(template_dir, template_name)
    base_path = os.path.join(os.path.dirname(__file__), 'templates', 'base.html')
    
    try:
        with open(template_path, 'r') as f:
            template_content = f.read()
        with open(base_path, 'r') as f:
            base_content = f.read()
        
        # Simple template inheritance simulation
        # In production, use proper Jinja2 template inheritance
        return render_template_string(template_content, current_page=template_name.replace('.html', ''))
    except FileNotFoundError:
        return f"Template {template_name} not found", 404

@app.route('/')
@app.route('/dashboard')
def dashboard():
    """Creator dashboard homepage"""
    return load_template('dashboard.html')

@app.route('/apis')
def apis():
    """APIs and deployments management"""
    return load_template('apis.html')

@app.route('/apis/<api_id>')
def api_detail(api_id):
    """Detailed view of a specific API"""
    # In a real app, fetch API details from database
    return f"<h1>API Detail: {api_id}</h1><p>Detailed API management interface would be here.</p>"

@app.route('/analytics')
def analytics():
    """Analytics and insights"""
    return load_template('analytics.html')

@app.route('/marketplace')
def marketplace():
    """Marketplace management"""
    return "<h1>Marketplace</h1><p>Marketplace management interface coming soon...</p>"

@app.route('/earnings')
def earnings():
    """Earnings and billing"""
    return load_template('earnings.html')

@app.route('/cli-setup')
def cli_setup():
    """CLI setup and configuration"""
    return """
    <h1>CLI Setup</h1>
    <div class="bg-white p-6 rounded-lg border">
        <h2>Install API-Direct CLI</h2>
        <pre class="bg-gray-900 text-green-400 p-4 rounded">
# macOS/Linux
curl -fsSL https://cli.apidirect.dev/install.sh | sh

# Windows  
iwr -useb https://cli.apidirect.dev/install.ps1 | iex

# npm
npm install -g @api-direct/cli
        </pre>
        
        <h3>Authenticate</h3>
        <pre class="bg-gray-900 text-green-400 p-4 rounded">
apidirect login
        </pre>
        
        <p>Your API Token: <code>apid_live_12345abcdef</code></p>
    </div>
    """

@app.route('/api-keys')
def api_keys():
    """API keys management"""
    return "<h1>API Keys</h1><p>API key management interface coming soon...</p>"

@app.route('/templates')
def templates():
    """API templates"""
    return "<h1>Templates</h1><p>Template management interface coming soon...</p>"

@app.route('/settings')
def settings():
    """Account settings"""
    return "<h1>Settings</h1><p>Account settings interface coming soon...</p>"

@app.route('/security')
def security():
    """Security settings"""
    return "<h1>Security</h1><p>Security settings interface coming soon...</p>"

@app.route('/help')
def help_center():
    """Help center"""
    return "<h1>Help Center</h1><p>Help and documentation coming soon...</p>"

@app.route('/community')
def community():
    """Community"""
    return "<h1>Community</h1><p>Community features coming soon...</p>"

@app.route('/health')
def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "api-direct-creator-portal", "version": "1.0.0"}

# Static file serving (in production, use a proper web server)
@app.route('/static/<path:filename>')
def static_files(filename):
    """Serve static files"""
    static_dir = os.path.join(os.path.dirname(__file__), 'static')
    return send_from_directory(static_dir, filename)

if __name__ == '__main__':
    print("🎨 API-Direct Creator Portal")
    print("📍 Running at: http://localhost:3003")
    print("")
    print("🔗 Available Routes:")
    print("   • http://localhost:3003/dashboard     - Creator Dashboard")
    print("   • http://localhost:3003/apis         - APIs & Deployments") 
    print("   • http://localhost:3003/analytics    - Analytics & Insights")
    print("   • http://localhost:3003/marketplace  - Marketplace Management")
    print("   • http://localhost:3003/earnings     - Earnings & Billing")
    print("   • http://localhost:3003/cli-setup    - CLI Setup Guide")
    print("   • http://localhost:3003/settings     - Account Settings")
    print("")
    print("💼 Multi-page Creator Portal ready for business!")
    
    app.run(host='0.0.0.0', port=3003, debug=True)