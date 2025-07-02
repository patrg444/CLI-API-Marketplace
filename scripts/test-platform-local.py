#!/usr/bin/env python3
"""
API-Direct Platform Local Testing
Tests the platform components that we can run locally
"""

import os
import sys
import time
import subprocess
import requests
import signal
from threading import Thread
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
    END = '\033[0m'

class PlatformTester:
    def __init__(self):
        self.base_dir = "/Users/patrickgloria/CLI-API-Marketplace"
        self.processes = []
        self.test_results = []

    def print_header(self, text):
        print(f"\n{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.END}")
        print(f"{Colors.CYAN}{Colors.BOLD}{text.center(60)}{Colors.END}")
        print(f"{Colors.CYAN}{Colors.BOLD}{'='*60}{Colors.END}\n")

    def print_status(self, test_name, status, details=""):
        if status:
            icon = f"{Colors.GREEN}✓{Colors.END}"
            status_text = f"{Colors.GREEN}PASS{Colors.END}"
        else:
            icon = f"{Colors.RED}✗{Colors.END}"
            status_text = f"{Colors.RED}FAIL{Colors.END}"
            
        print(f"{icon} {test_name:<40} [{status_text}] {details}")
        self.test_results.append({'test': test_name, 'status': status, 'details': details})

    def cleanup(self):
        """Kill all started processes"""
        print(f"\n{Colors.YELLOW}Cleaning up processes...{Colors.END}")
        for proc in self.processes:
            try:
                proc.terminate()
                proc.wait(timeout=5)
            except:
                try:
                    proc.kill()
                except:
                    pass

    def start_frontend_server(self, port, directory, name):
        """Start a Python HTTP server"""
        try:
            cmd = ["python3", "-m", "http.server", str(port)]
            proc = subprocess.Popen(
                cmd,
                cwd=os.path.join(self.base_dir, directory),
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )
            self.processes.append(proc)
            
            # Wait a moment for server to start
            time.sleep(2)
            
            # Test if server is responding
            try:
                response = requests.get(f"http://localhost:{port}", timeout=5)
                if response.status_code == 200:
                    self.print_status(f"{name} server", True, f"http://localhost:{port}")
                    return True
            except requests.exceptions.RequestException:
                pass
                
            self.print_status(f"{name} server", False, "Failed to start")
            return False
            
        except Exception as e:
            self.print_status(f"{name} server", False, f"Error: {str(e)}")
            return False

    def test_file_structure(self):
        """Test critical files exist"""
        self.print_header("TESTING FILE STRUCTURE")
        
        critical_files = [
            "backend/api/main.py",
            "backend/database/schema.sql",
            "web/console/login.html",
            "web/console/register.html",
            "web/console/pages/dashboard.html",
            "web/console/static/js/api-client.js",
            "web/landing/index.html",
            "docker-compose.production.yml",
            "scripts/deploy.sh"
        ]
        
        for file_path in critical_files:
            full_path = os.path.join(self.base_dir, file_path)
            exists = os.path.exists(full_path)
            self.print_status(f"File: {file_path}", exists)

    def test_frontend_pages(self):
        """Test frontend pages load correctly"""
        self.print_header("TESTING FRONTEND PAGES")
        
        pages = [
            ("login.html", "Sign In - API-Direct"),
            ("register.html", "Create Account - API-Direct"),
            ("pages/dashboard.html", "Dashboard"),
            ("pages/apis.html", "APIs"),
            ("pages/marketplace.html", "Marketplace")
        ]
        
        for page, expected_title in pages:
            try:
                response = requests.get(f"http://localhost:8080/{page}", timeout=5)
                
                if response.status_code == 200:
                    has_title = expected_title.lower() in response.text.lower()
                    has_html = "<html" in response.text and "</html>" in response.text
                    
                    if has_title and has_html:
                        self.print_status(f"Page: {page}", True, "Content loaded")
                    else:
                        self.print_status(f"Page: {page}", False, "Missing expected content")
                else:
                    self.print_status(f"Page: {page}", False, f"HTTP {response.status_code}")
                    
            except requests.exceptions.RequestException as e:
                self.print_status(f"Page: {page}", False, f"Request failed: {str(e)}")

    def test_api_client_js(self):
        """Test API client JavaScript file"""
        self.print_header("TESTING API CLIENT")
        
        try:
            api_client_path = os.path.join(self.base_dir, "web/console/static/js/api-client.js")
            with open(api_client_path, 'r') as f:
                content = f.read()
                
            tests = [
                ("APIClient class", "class APIClient" in content),
                ("Authentication headers", "Authorization" in content and "Bearer" in content),
                ("Dashboard method", "getDashboardOverview" in content),
                ("APIs method", "getAPIs" in content),
                ("Analytics method", "getAnalytics" in content),
                ("Marketplace method", "getMarketplace" in content),
                ("Error handling", "catch" in content)
            ]
            
            for test_name, condition in tests:
                self.print_status(test_name, condition)
                
        except Exception as e:
            self.print_status("API Client file", False, f"Error: {str(e)}")

    def test_landing_page(self):
        """Test landing page content"""
        self.print_header("TESTING LANDING PAGE")
        
        try:
            response = requests.get("http://localhost:3003", timeout=5)
            
            if response.status_code == 200:
                content = response.text.lower()
                
                tests = [
                    ("Page loads", True),
                    ("Has title", "api-direct" in content),
                    ("Has main heading", "infrastructure for the ai agent economy" in content),
                    ("Has pricing section", "pricing" in content),
                    ("Has hero section", "hero" in content or "deploy" in content)
                ]
                
                for test_name, condition in tests:
                    if test_name == "Page loads":
                        self.print_status(test_name, condition)
                    else:
                        self.print_status(test_name, condition)
            else:
                self.print_status("Landing page", False, f"HTTP {response.status_code}")
                
        except requests.exceptions.RequestException as e:
            self.print_status("Landing page", False, f"Request failed: {str(e)}")

    def test_authentication_flow(self):
        """Test authentication page structure"""
        self.print_header("TESTING AUTHENTICATION")
        
        # Test login page
        try:
            response = requests.get("http://localhost:8080/login.html", timeout=5)
            if response.status_code == 200:
                content = response.text
                
                login_tests = [
                    ("Login form exists", 'id="login-form"' in content),
                    ("Email field", 'type="email"' in content),
                    ("Password field", 'type="password"' in content),
                    ("Remember me", "remember" in content.lower()),
                    ("Submit button", 'type="submit"' in content)
                ]
                
                for test_name, condition in login_tests:
                    self.print_status(f"Login: {test_name}", condition)
            else:
                self.print_status("Login page", False, f"HTTP {response.status_code}")
                
        except Exception as e:
            self.print_status("Login page", False, f"Error: {str(e)}")
            
        # Test registration page
        try:
            response = requests.get("http://localhost:8080/register.html", timeout=5)
            if response.status_code == 200:
                content = response.text
                
                register_tests = [
                    ("Registration form", 'id="register-form"' in content),
                    ("Name field", 'id="name"' in content),
                    ("Password strength", "password-strength" in content),
                    ("Terms checkbox", "terms" in content and "checkbox" in content)
                ]
                
                for test_name, condition in register_tests:
                    self.print_status(f"Register: {test_name}", condition)
            else:
                self.print_status("Register page", False, f"HTTP {response.status_code}")
                
        except Exception as e:
            self.print_status("Register page", False, f"Error: {str(e)}")

    def run_tests(self):
        """Run all local tests"""
        print(f"{Colors.PURPLE}{Colors.BOLD}")
        print("╔══════════════════════════════════════════════════════════╗")
        print("║              API-DIRECT LOCAL TESTING SUITE              ║")
        print("║          Testing Platform Components Locally            ║")
        print("╚══════════════════════════════════════════════════════════╝")
        print(f"{Colors.END}")
        
        # Test file structure first
        self.test_file_structure()
        
        # Start frontend servers
        self.print_header("STARTING FRONTEND SERVICES")
        
        landing_running = self.start_frontend_server(3003, "web/landing", "Landing page")
        console_running = self.start_frontend_server(8080, "web/console", "Creator portal")
        
        if not (landing_running and console_running):
            print(f"\n{Colors.RED}Could not start frontend services. Stopping tests.{Colors.END}")
            self.cleanup()
            return False
            
        # Run tests
        self.test_api_client_js()
        
        if landing_running:
            self.test_landing_page()
            
        if console_running:
            self.test_frontend_pages()
            self.test_authentication_flow()
        
        # Summary
        self.print_header("TEST SUMMARY")
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result['status'])
        pass_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"{Colors.BOLD}Total Tests:{Colors.END} {total_tests}")
        print(f"{Colors.GREEN}{Colors.BOLD}Passed:{Colors.END} {passed_tests}")
        print(f"{Colors.RED}{Colors.BOLD}Failed:{Colors.END} {total_tests - passed_tests}")
        print(f"{Colors.BLUE}{Colors.BOLD}Pass Rate:{Colors.END} {pass_rate:.1f}%")
        
        if pass_rate >= 80:
            status_color = Colors.GREEN
            status_text = "EXCELLENT"
        elif pass_rate >= 60:
            status_color = Colors.YELLOW
            status_text = "GOOD"
        else:
            status_color = Colors.RED
            status_text = "NEEDS WORK"
            
        print(f"\n{status_color}{Colors.BOLD}Overall Status: {status_text}{Colors.END}")
        
        if landing_running and console_running:
            print(f"\n{Colors.CYAN}🌐 Platform URLs:{Colors.END}")
            print(f"  Landing Page:    http://localhost:3003")
            print(f"  Creator Portal:  http://localhost:8080/login.html")
            print(f"  Registration:    http://localhost:8080/register.html")
            print(f"  Dashboard:       http://localhost:8080/pages/dashboard.html")
            
        return pass_rate >= 80

def main():
    tester = PlatformTester()
    
    # Setup signal handler for cleanup
    def signal_handler(signum, frame):
        print(f"\n{Colors.YELLOW}Received interrupt signal. Cleaning up...{Colors.END}")
        tester.cleanup()
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        success = tester.run_tests()
        
        if success:
            print(f"\n{Colors.GREEN}{Colors.BOLD}🎉 PLATFORM COMPONENTS WORKING LOCALLY! 🚀{Colors.END}")
            print(f"\n{Colors.YELLOW}Press Ctrl+C to stop servers and exit{Colors.END}")
            
            # Keep servers running
            try:
                while True:
                    time.sleep(1)
            except KeyboardInterrupt:
                pass
        else:
            print(f"\n{Colors.RED}{Colors.BOLD}⚠️ SOME ISSUES FOUND{Colors.END}")
            
    finally:
        tester.cleanup()

if __name__ == "__main__":
    main()