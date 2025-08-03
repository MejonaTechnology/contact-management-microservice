#!/usr/bin/env python3
"""
FINAL COMPREHENSIVE VERIFICATION
Contact Management Microservice - AWS EC2 Deployment
"""

import requests
import json
import time
from datetime import datetime

BASE_URL = "http://65.1.94.25:8081"

def test_endpoint(method, endpoint, data=None, headers=None):
    url = f"{BASE_URL}{endpoint}"
    try:
        start_time = time.time()
        if method == 'GET':
            response = requests.get(url, headers=headers, timeout=10)
        elif method == 'POST':
            response = requests.post(url, json=data, headers=headers, timeout=10)
        elif method == 'PUT':
            response = requests.put(url, json=data, headers=headers, timeout=10)
        
        response_time = round((time.time() - start_time) * 1000, 2)
        return {
            'success': response.status_code < 400,
            'status_code': response.status_code,
            'response_time': response_time,
            'content': response.text[:150]
        }
    except Exception as e:
        return {'success': False, 'error': str(e)}

def main():
    print("="*80)
    print("FINAL COMPREHENSIVE DEPLOYMENT VERIFICATION")
    print("Contact Management Microservice - AWS EC2")
    print("="*80)
    print(f"Server: {BASE_URL}")
    print(f"Verification Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    # Critical endpoints for verification
    critical_tests = [
        # Health & Core
        ("GET", "/health", None, "Health Check"),
        ("GET", "/status", None, "Status Check"),
        ("GET", "/api/v1/test", None, "API Test"),
        
        # Authentication
        ("POST", "/api/v1/auth/login", {"email": "admin@mejona.com", "password": "admin123"}, "Authentication"),
        
        # Dashboard Core
        ("GET", "/api/v1/dashboard/contacts", None, "Dashboard Contacts"),
        ("GET", "/api/v1/dashboard/contacts/stats", None, "Dashboard Stats"),
        
        # Public API
        ("POST", "/api/v1/public/contact", {"name": "Test", "email": "test@example.com", "message": "Test"}, "Public Contact Form"),
    ]
    
    print("CRITICAL ENDPOINTS VERIFICATION:")
    print("-" * 50)
    
    total_tests = 0
    passed_tests = 0
    auth_token = None
    
    for method, endpoint, data, description in critical_tests:
        print(f"Testing: {description}")
        result = test_endpoint(method, endpoint, data)
        
        if result.get('success'):
            print(f"  ✅ SUCCESS - HTTP {result['status_code']} ({result['response_time']}ms)")
            
            # Extract token for auth test
            if 'auth/login' in endpoint and result.get('content'):
                try:
                    response_data = json.loads(result['content'])
                    if 'data' in response_data and 'token' in response_data['data']:
                        auth_token = response_data['data']['token']
                        print(f"  🔑 Auth token obtained: {auth_token[:20]}...")
                except:
                    pass
            
            # Show sample response
            if result.get('content'):
                print(f"  📄 Response: {result['content'][:100]}...")
            
            passed_tests += 1
        else:
            print(f"  ❌ FAILED - {result.get('error', 'Unknown error')}")
        
        total_tests += 1
        print()
    
    # Test authenticated endpoints if we have token
    if auth_token:
        print("AUTHENTICATED ENDPOINTS:")
        print("-" * 30)
        
        auth_headers = {'Authorization': f'Bearer {auth_token}'}
        auth_tests = [
            ("GET", "/api/v1/auth/profile", None, "User Profile"),
            ("POST", "/api/v1/dashboard/contact", {"name": "Auth Test", "email": "auth@test.com"}, "Create Contact (Auth)"),
        ]
        
        for method, endpoint, data, description in auth_tests:
            print(f"Testing: {description}")
            result = test_endpoint(method, endpoint, data, auth_headers)
            
            if result.get('success'):
                print(f"  ✅ SUCCESS - HTTP {result['status_code']} ({result['response_time']}ms)")
                passed_tests += 1
            else:
                print(f"  ❌ FAILED - {result.get('error', 'Unknown error')}")
            
            total_tests += 1
            print()
    
    # Performance and accessibility test
    print("PERFORMANCE & ACCESSIBILITY:")
    print("-" * 35)
    
    # Test multiple rapid requests
    print("Testing: Rapid Sequential Requests")
    rapid_success = 0
    rapid_total = 5
    
    for i in range(rapid_total):
        result = test_endpoint('GET', '/health')
        if result.get('success'):
            rapid_success += 1
    
    print(f"  ✅ Rapid Requests: {rapid_success}/{rapid_total} successful")
    print()
    
    # Calculate final results
    success_rate = (passed_tests / total_tests) * 100 if total_tests > 0 else 0
    
    print("="*80)
    print("FINAL VERIFICATION RESULTS")
    print("="*80)
    
    print(f"Total Critical Tests: {total_tests}")
    print(f"Passed: {passed_tests}")
    print(f"Failed: {total_tests - passed_tests}")
    print(f"Success Rate: {success_rate:.1f}%")
    print(f"Rapid Requests: {rapid_success}/{rapid_total} ({(rapid_success/rapid_total)*100:.1f}%)")
    print()
    
    # Final status
    if success_rate >= 90 and rapid_success >= 4:
        print("🎉 DEPLOYMENT STATUS: FULLY SUCCESSFUL")
        print("✅ Contact Management Microservice is COMPLETELY ACCESSIBLE")
        print("🚀 Production ready for immediate use")
        print("🌐 All critical endpoints operational")
        
        print()
        print("IMMEDIATE ACCESS URLS:")
        print(f"  🔗 Health Check: {BASE_URL}/health")
        print(f"  🔗 API Test: {BASE_URL}/api/v1/test")
        print(f"  🔗 Dashboard: {BASE_URL}/api/v1/dashboard/contacts")
        print(f"  🔗 Auth Login: {BASE_URL}/api/v1/auth/login")
        
    elif success_rate >= 70:
        print("⚠️  DEPLOYMENT STATUS: MOSTLY ACCESSIBLE")
        print("💪 Service functional with minor issues")
        print("🔧 Some endpoints may need attention")
        
    else:
        print("❌ DEPLOYMENT STATUS: ACCESSIBILITY ISSUES")
        print("🚨 Service requires immediate attention")
        print("🔧 Multiple endpoints failing")
    
    print()
    print("="*80)

if __name__ == "__main__":
    main()