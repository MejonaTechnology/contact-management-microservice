#!/usr/bin/env python3
"""
Test All 20 Endpoints of Contact Management Microservice
"""

import requests
import json
from datetime import datetime

BASE_URL = "http://65.1.94.25:8081"

def test_endpoint(method, endpoint, data=None, headers=None):
    url = f"{BASE_URL}{endpoint}"
    try:
        if method == 'GET':
            response = requests.get(url, headers=headers, timeout=10)
        elif method == 'POST':
            response = requests.post(url, json=data, headers=headers, timeout=10)
        elif method == 'PUT':
            response = requests.put(url, json=data, headers=headers, timeout=10)
        return response.status_code < 400, response.status_code, response.text[:100]
    except:
        return False, 0, "Connection failed"

def main():
    print("=" * 80)
    print("COMPLETE 20 ENDPOINT VERIFICATION")
    print("Contact Management Microservice - AWS EC2 Deployment")
    print("=" * 80)
    print(f"Server: {BASE_URL}")
    print(f"Test Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    # All 20 endpoints to test
    endpoints = [
        # Health & Monitoring (6 endpoints)
        ("GET", "/health", None, "Basic Health Check"),
        ("GET", "/health/deep", None, "Deep Health Check"),
        ("GET", "/status", None, "Status Check"),
        ("GET", "/ready", None, "Readiness Check"),
        ("GET", "/alive", None, "Liveness Check"),
        ("GET", "/metrics", None, "Metrics Check"),
        
        # Public & Utility (2 endpoints)
        ("GET", "/api/v1/test", None, "Test Endpoint"),
        ("POST", "/api/v1/public/contact", {"name": "Test", "email": "test@test.com", "message": "Test"}, "Public Contact Form"),
        
        # Authentication (6 endpoints)
        ("POST", "/api/v1/auth/login", {"email": "admin@mejona.com", "password": "admin123"}, "User Login"),
        ("POST", "/api/v1/auth/refresh", {}, "Refresh Token"),
        ("GET", "/api/v1/auth/profile", None, "Get User Profile"),
        ("GET", "/api/v1/auth/validate", None, "Validate Token"),
        ("POST", "/api/v1/auth/logout", {}, "User Logout"),
        ("POST", "/api/v1/auth/change-password", {"old_password": "old", "new_password": "new"}, "Change Password"),
        
        # Dashboard Management (6 endpoints)
        ("GET", "/api/v1/dashboard/contacts", None, "List Contacts"),
        ("GET", "/api/v1/dashboard/contacts/stats", None, "Contact Statistics"),
        ("POST", "/api/v1/dashboard/contact", {"name": "Test", "email": "test@test.com"}, "Create Contact"),
        ("PUT", "/api/v1/dashboard/contacts/1/status", {"status": "resolved"}, "Update Contact Status"),
        ("GET", "/api/v1/dashboard/contacts/1", None, "Get Contact by ID"),
        ("GET", "/api/v1/dashboard/contacts/export", None, "Export Contacts"),
        ("POST", "/api/v1/dashboard/contacts/bulk-update", {"ids": [1,2], "status": "resolved"}, "Bulk Update"),
    ]
    
    total_tests = len(endpoints)
    passed_tests = 0
    
    for i, (method, endpoint, data, description) in enumerate(endpoints, 1):
        print(f"{i:2d}. Testing {description}...")
        success, status_code, response = test_endpoint(method, endpoint, data)
        
        if success:
            print(f"    ✅ SUCCESS - HTTP {status_code}")
            print(f"    Response: {response}...")
            passed_tests += 1
        else:
            print(f"    ❌ FAILED - HTTP {status_code}")
            print(f"    Error: {response}")
        print()
    
    # Summary
    success_rate = (passed_tests / total_tests) * 100
    print("=" * 80)
    print("FINAL VERIFICATION RESULTS")
    print("=" * 80)
    print(f"Total Endpoints: {total_tests}")
    print(f"Passed: {passed_tests}")
    print(f"Failed: {total_tests - passed_tests}")
    print(f"Success Rate: {success_rate:.1f}%")
    print()
    
    if success_rate == 100:
        print("🎉 DEPLOYMENT STATUS: FULLY OPERATIONAL")
        print("✅ All 20 endpoints are working correctly")
        print("🚀 Contact Management Microservice is production-ready")
    elif success_rate >= 80:
        print("⚠️  DEPLOYMENT STATUS: MOSTLY OPERATIONAL")
        print("💪 Service is functional with minor issues")
    else:
        print("❌ DEPLOYMENT STATUS: NEEDS ATTENTION")
        print("🔧 Service requires troubleshooting")
    
    print()
    print("Service Information:")
    print(f"  Main URL: {BASE_URL}")
    print(f"  Health: {BASE_URL}/health")
    print(f"  API Docs: Available in repository")
    print(f"  GitHub: https://github.com/MejonaTechnology/contact-management-microservice")

if __name__ == "__main__":
    main()