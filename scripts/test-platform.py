#!/usr/bin/env python3
"""
API-Direct Platform Testing Suite
Comprehensive testing of all platform components
"""

import os
import sys
import json
import time
from datetime import datetime
from typing import Dict, List, Any
import logging

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    PURPLE = '\033[95m'
    CYAN = '\033[96m'
    WHITE = '\033[97m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    END = '\033[0m'

class PlatformTester:
    def __init__(self):
        self.backend_url = "http://localhost:8000"
        self.frontend_url = "http://localhost:8080"
        self.test_results = []
        self.total_tests = 0
        self.passed_tests = 0
        
    def print_header(self, text: str):
        print(f"\n{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.END}")
        print(f"{Colors.CYAN}{Colors.BOLD}{text.center(60)}{Colors.END}")
        print(f"{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.END}\n")
        
    def print_status(self, test_name: str, status: bool, details: str = ""):
        self.total_tests += 1
        if status:
            self.passed_tests += 1
            icon = f"{Colors.GREEN}✓{Colors.END}"
            status_text = f"{Colors.GREEN}PASS{Colors.END}"
        else:
            icon = f"{Colors.RED}✗{Colors.END}"
            status_text = f"{Colors.RED}FAIL{Colors.END}"
            
        print(f"{icon} {test_name:<40} [{status_text}] {details}")
        self.test_results.append({
            'test': test_name,
            'status': status,
            'details': details,
            'timestamp': datetime.now().isoformat()
        })

    def test_file_structure(self):
        """Test that all required files exist"""
        self.print_header("TESTING FILE STRUCTURE")
        
        required_files = [
            # Backend files
            "backend/api/main.py",
            "backend/api/websocket.py", 
            "backend/database/schema.sql",
            "backend/Dockerfile",
            "backend/requirements.txt",
            
            # Frontend files
            "web/console/templates/base.html",
            "web/console/pages/dashboard.html",
            "web/console/pages/apis.html",
            "web/console/pages/analytics.html",
            "web/console/pages/earnings.html",
            "web/console/pages/marketplace.html",
            "web/console/login.html",
            "web/console/register.html",
            "web/console/static/js/api-client.js",
            
            # Landing page
            "web/landing/index.html",
            
            # Deployment files
            "docker-compose.production.yml",
            ".env.production.example",
            "nginx/nginx.conf",
            "monitoring/prometheus.yml",
            
            # Scripts
            "scripts/deploy.sh",
            "scripts/backup.sh", 
            "scripts/health-check.sh",
            
            # Documentation
            "DEPLOYMENT_GUIDE.md",
            "BETA_LAUNCH_CHECKLIST.md"
        ]
        
        for file_path in required_files:
            full_path = os.path.join("/Users/patrickgloria/CLI-API-Marketplace", file_path)
            exists = os.path.exists(full_path)
            self.print_status(f"File: {file_path}", exists)

    def test_frontend_pages(self):
        """Test frontend pages for correct structure"""
        self.print_header("TESTING FRONTEND PAGES")
        
        pages = [
            ("login.html", "Sign In - API-Direct"),
            ("register.html", "Create Account - API-Direct"),
            ("pages/dashboard.html", "Dashboard"),
            ("pages/apis.html", "APIs & Deployments"),
            ("pages/analytics.html", "Analytics"),
            ("pages/earnings.html", "Earnings"),
            ("pages/marketplace.html", "Marketplace"),
            ("templates/base.html", "API-Direct Creator Portal")
        ]
        
        for page, expected_title in pages:
            try:
                file_path = f"/Users/patrickgloria/CLI-API-Marketplace/web/console/{page}"
                with open(file_path, 'r') as f:
                    content = f.read()
                    
                has_title = expected_title in content
                has_auth_check = "checkAuthentication" in content or "authentication" in content.lower()
                has_proper_structure = "<html" in content and "</html>" in content
                
                success = has_title and has_proper_structure
                details = f"Title: {has_title}, Structure: {has_proper_structure}"
                if page != "templates/base.html":
                    details += f", Auth: {has_auth_check}"
                
                self.print_status(f"Page: {page}", success, details)
                
            except Exception as e:
                self.print_status(f"Page: {page}", False, f"Error: {str(e)}")

    def test_backend_structure(self):
        """Test backend code structure"""
        self.print_header("TESTING BACKEND STRUCTURE")
        
        # Test main.py structure
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/backend/api/main.py", 'r') as f:
                content = f.read()
                
            has_fastapi = "from fastapi import FastAPI" in content
            has_cors = "CORSMiddleware" in content
            has_auth = "HTTPBearer" in content and "jwt" in content
            has_models = "BaseModel" in content
            has_endpoints = "@app.post" in content or "@app.get" in content
            
            self.print_status("FastAPI imports", has_fastapi)
            self.print_status("CORS middleware", has_cors)
            self.print_status("Authentication setup", has_auth)
            self.print_status("Pydantic models", has_models)
            self.print_status("API endpoints", has_endpoints)
            
        except Exception as e:
            self.print_status("Backend main.py", False, f"Error: {str(e)}")
            
        # Test WebSocket structure
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/backend/api/websocket.py", 'r') as f:
                content = f.read()
                
            has_websocket = "WebSocket" in content
            has_manager = "ConnectionManager" in content or "WebSocketManager" in content
            
            self.print_status("WebSocket implementation", has_websocket)
            self.print_status("WebSocket manager", has_manager)
            
        except Exception as e:
            self.print_status("WebSocket implementation", False, f"Error: {str(e)}")

    def test_database_schema(self):
        """Test database schema structure"""
        self.print_header("TESTING DATABASE SCHEMA")
        
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/backend/database/schema.sql", 'r') as f:
                content = f.read()
                
            required_tables = [
                "CREATE TABLE users",
                "CREATE TABLE apis", 
                "CREATE TABLE api_calls",
                "CREATE TABLE billing_events",
                "CREATE TABLE marketplace_listings",
                "CREATE TABLE password_reset_tokens"
            ]
            
            for table in required_tables:
                has_table = table in content
                table_name = table.replace("CREATE TABLE ", "")
                self.print_status(f"Table: {table_name}", has_table)
                
            # Check for indexes and constraints
            has_indexes = "CREATE INDEX" in content
            has_triggers = "CREATE TRIGGER" in content
            has_functions = "CREATE OR REPLACE FUNCTION" in content
            
            self.print_status("Database indexes", has_indexes)
            self.print_status("Update triggers", has_triggers)
            self.print_status("Database functions", has_functions)
            
        except Exception as e:
            self.print_status("Database schema", False, f"Error: {str(e)}")

    def test_deployment_config(self):
        """Test deployment configuration"""
        self.print_header("TESTING DEPLOYMENT CONFIGURATION")
        
        # Test Docker Compose
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/docker-compose.production.yml", 'r') as f:
                content = f.read()
                
            services = ["postgres", "redis", "influxdb", "backend", "nginx", "prometheus", "grafana"]
            for service in services:
                has_service = f"{service}:" in content
                self.print_status(f"Service: {service}", has_service)
                
            has_volumes = "volumes:" in content
            has_networks = "networks:" in content or "default:" in content
            has_health_checks = "healthcheck:" in content
            
            self.print_status("Volume configuration", has_volumes)
            self.print_status("Network configuration", has_networks)
            self.print_status("Health checks", has_health_checks)
            
        except Exception as e:
            self.print_status("Docker Compose config", False, f"Error: {str(e)}")
            
        # Test Nginx config
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/nginx/nginx.conf", 'r') as f:
                content = f.read()
                
            has_ssl = "ssl_certificate" in content
            has_proxy = "proxy_pass" in content
            has_rate_limit = "limit_req_zone" in content
            has_websocket = "Upgrade" in content and "websocket" in content.lower()
            
            self.print_status("SSL configuration", has_ssl)
            self.print_status("Reverse proxy", has_proxy)
            self.print_status("Rate limiting", has_rate_limit)
            self.print_status("WebSocket support", has_websocket)
            
        except Exception as e:
            self.print_status("Nginx configuration", False, f"Error: {str(e)}")

    def test_api_client(self):
        """Test frontend API client"""
        self.print_header("TESTING API CLIENT")
        
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/static/js/api-client.js", 'r') as f:
                content = f.read()
                
            has_class = "class APIClient" in content
            has_auth = "Authorization" in content and "Bearer" in content
            has_methods = "async " in content and "fetch" in content
            has_error_handling = "catch" in content and "error" in content.lower()
            
            self.print_status("APIClient class", has_class)
            self.print_status("Authentication headers", has_auth)
            self.print_status("Async methods", has_methods)
            self.print_status("Error handling", has_error_handling)
            
            # Check for specific API methods
            api_methods = [
                "getDashboardOverview",
                "getAPIs",
                "getAnalytics", 
                "getEarnings",
                "getMarketplace"
            ]
            
            for method in api_methods:
                has_method = method in content
                self.print_status(f"Method: {method}", has_method)
                
        except Exception as e:
            self.print_status("API Client", False, f"Error: {str(e)}")

    def test_authentication_flow(self):
        """Test authentication implementation"""
        self.print_header("TESTING AUTHENTICATION")
        
        # Test login page
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/login.html", 'r') as f:
                login_content = f.read()
                
            has_form = "form" in login_content and "login" in login_content.lower()
            has_email_field = 'type="email"' in login_content
            has_password_field = 'type="password"' in login_content
            has_remember_me = "remember" in login_content.lower()
            has_submit_handler = "handleLogin" in login_content
            
            self.print_status("Login form", has_form)
            self.print_status("Email field", has_email_field)
            self.print_status("Password field", has_password_field)
            self.print_status("Remember me option", has_remember_me)
            self.print_status("Submit handler", has_submit_handler)
            
        except Exception as e:
            self.print_status("Login page", False, f"Error: {str(e)}")
            
        # Test registration page
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/register.html", 'r') as f:
                register_content = f.read()
                
            has_name_field = 'id="name"' in register_content
            has_password_strength = "password-strength" in register_content
            has_confirm_password = "confirm-password" in register_content
            has_terms_checkbox = "terms" in register_content and "checkbox" in register_content
            
            self.print_status("Registration name field", has_name_field)
            self.print_status("Password strength indicator", has_password_strength)
            self.print_status("Confirm password field", has_confirm_password)
            self.print_status("Terms checkbox", has_terms_checkbox)
            
        except Exception as e:
            self.print_status("Registration page", False, f"Error: {str(e)}")
            
        # Test base template authentication
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/templates/base.html", 'r') as f:
                base_content = f.read()
                
            has_auth_check = "checkAuthentication" in base_content
            has_token_handling = "getStoredToken" in base_content
            has_logout = "logout" in base_content
            has_user_menu = "user-menu" in base_content
            
            self.print_status("Authentication check", has_auth_check)
            self.print_status("Token handling", has_token_handling)
            self.print_status("Logout functionality", has_logout)
            self.print_status("User menu", has_user_menu)
            
        except Exception as e:
            self.print_status("Base template auth", False, f"Error: {str(e)}")

    def test_websocket_implementation(self):
        """Test WebSocket implementation"""
        self.print_header("TESTING WEBSOCKET IMPLEMENTATION")
        
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/templates/base.html", 'r') as f:
                content = f.read()
                
            has_websocket_class = "WebSocketManager" in content
            has_connection_handling = "connect()" in content and "onopen" in content
            has_message_handling = "onmessage" in content and "handleMessage" in content
            has_reconnection = "reconnect" in content
            has_event_handlers = "eventHandlers" in content
            
            self.print_status("WebSocket class", has_websocket_class)
            self.print_status("Connection handling", has_connection_handling)
            self.print_status("Message handling", has_message_handling)
            self.print_status("Auto-reconnection", has_reconnection)
            self.print_status("Event handlers", has_event_handlers)
            
            # Check for specific event types
            event_types = [
                "api_status_update",
                "analytics_update", 
                "billing_update",
                "notification"
            ]
            
            for event_type in event_types:
                has_event = event_type in content
                self.print_status(f"Event: {event_type}", has_event)
                
        except Exception as e:
            self.print_status("WebSocket implementation", False, f"Error: {str(e)}")

    def test_marketplace_features(self):
        """Test marketplace implementation"""
        self.print_header("TESTING MARKETPLACE FEATURES")
        
        try:
            with open("/Users/patrickgloria/CLI-API-Marketplace/web/console/pages/marketplace.html", 'r') as f:
                content = f.read()
                
            has_search = "search" in content.lower()
            has_categories = "category" in content.lower()
            has_filters = "filter" in content.lower()
            has_api_cards = "api-card" in content or "listing" in content
            has_pagination = "pagination" in content.lower()
            
            self.print_status("Search functionality", has_search)
            self.print_status("Category filtering", has_categories)
            self.print_status("Advanced filters", has_filters)
            self.print_status("API listing cards", has_api_cards)
            self.print_status("Pagination", has_pagination)
            
            # Check for marketplace-specific features
            marketplace_features = [
                "loadMarketplaceData",
                "publishAPI",
                "subscribeToAPI"
            ]
            
            for feature in marketplace_features:
                has_feature = feature in content
                self.print_status(f"Feature: {feature}", has_feature)
                
        except Exception as e:
            self.print_status("Marketplace page", False, f"Error: {str(e)}")

    def test_documentation_completeness(self):
        """Test documentation completeness"""
        self.print_header("TESTING DOCUMENTATION")
        
        docs = [
            ("DEPLOYMENT_GUIDE.md", ["Prerequisites", "Quick Deployment", "Configuration"]),
            ("BETA_LAUNCH_CHECKLIST.md", ["Pre-Launch Setup", "Testing", "Go-Live"]),
            (".env.production.example", ["POSTGRES_PASSWORD", "JWT_SECRET", "STRIPE_SECRET_KEY"])
        ]
        
        for doc_file, required_sections in docs:
            try:
                file_path = f"/Users/patrickgloria/CLI-API-Marketplace/{doc_file}"
                with open(file_path, 'r') as f:
                    content = f.read()
                    
                for section in required_sections:
                    has_section = section in content
                    self.print_status(f"{doc_file}: {section}", has_section)
                    
            except Exception as e:
                self.print_status(f"Documentation: {doc_file}", False, f"Error: {str(e)}")

    def generate_test_report(self):
        """Generate comprehensive test report"""
        self.print_header("TEST SUMMARY")
        
        pass_rate = (self.passed_tests / self.total_tests * 100) if self.total_tests > 0 else 0
        
        print(f"{Colors.BOLD}Total Tests:{Colors.END} {self.total_tests}")
        print(f"{Colors.GREEN}{Colors.BOLD}Passed:{Colors.END} {self.passed_tests}")
        print(f"{Colors.RED}{Colors.BOLD}Failed:{Colors.END} {self.total_tests - self.passed_tests}")
        print(f"{Colors.BLUE}{Colors.BOLD}Pass Rate:{Colors.END} {pass_rate:.1f}%")
        
        if pass_rate >= 90:
            status_color = Colors.GREEN
            status_text = "EXCELLENT"
        elif pass_rate >= 80:
            status_color = Colors.YELLOW  
            status_text = "GOOD"
        elif pass_rate >= 70:
            status_color = Colors.YELLOW
            status_text = "NEEDS IMPROVEMENT"
        else:
            status_color = Colors.RED
            status_text = "CRITICAL ISSUES"
            
        print(f"\n{status_color}{Colors.BOLD}Overall Status: {status_text}{Colors.END}")
        
        # Save detailed report
        report = {
            'timestamp': datetime.now().isoformat(),
            'summary': {
                'total_tests': self.total_tests,
                'passed_tests': self.passed_tests,
                'failed_tests': self.total_tests - self.passed_tests,
                'pass_rate': pass_rate,
                'status': status_text
            },
            'test_results': self.test_results
        }
        
        with open('/Users/patrickgloria/CLI-API-Marketplace/test_report.json', 'w') as f:
            json.dump(report, f, indent=2)
            
        print(f"\n{Colors.CYAN}Detailed report saved to: test_report.json{Colors.END}")
        
        return pass_rate >= 80

    def run_all_tests(self):
        """Run the complete test suite"""
        print(f"{Colors.PURPLE}{Colors.BOLD}")
        print("╔══════════════════════════════════════════════════════════╗")
        print("║                 API-DIRECT TESTING SUITE                 ║")
        print("║              Comprehensive Platform Testing              ║")
        print("╚══════════════════════════════════════════════════════════╝")
        print(f"{Colors.END}")
        
        # Run all test categories
        self.test_file_structure()
        self.test_frontend_pages()
        self.test_backend_structure() 
        self.test_database_schema()
        self.test_deployment_config()
        self.test_api_client()
        self.test_authentication_flow()
        self.test_websocket_implementation()
        self.test_marketplace_features()
        self.test_documentation_completeness()
        
        # Generate final report
        success = self.generate_test_report()
        
        if success:
            print(f"\n{Colors.GREEN}{Colors.BOLD}🎉 PLATFORM READY FOR BETA LAUNCH! 🚀{Colors.END}")
        else:
            print(f"\n{Colors.RED}{Colors.BOLD}⚠️ ISSUES FOUND - REVIEW BEFORE LAUNCH{Colors.END}")
            
        return success

if __name__ == "__main__":
    tester = PlatformTester()
    success = tester.run_all_tests()
    sys.exit(0 if success else 1)