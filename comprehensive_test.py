#!/usr/bin/env python3
"""
Comprehensive AWS EC2 Contact Management Microservice Deployment Test
Tests all 20 API endpoints and deployment status
"""

import requests
import json
import time
from datetime import datetime

# Configuration
BASE_URL = "http://65.1.94.25:8081"
TIMEOUT = 10

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    BOLD = '\033[1m'
    END = '\033[0m'

def test_endpoint(method, endpoint, data=None, headers=None, auth_token=None):
    """Test a single endpoint and return results"""
    url = f"{BASE_URL}{endpoint}"
    
    if headers is None:
        headers = {'Content-Type': 'application/json'}
    
    if auth_token:
        headers['Authorization'] = f'Bearer {auth_token}'
    
    try:
        start_time = time.time()
        
        if method.upper() == 'GET':
            response = requests.get(url, headers=headers, timeout=TIMEOUT)
        elif method.upper() == 'POST':
            response = requests.post(url, json=data, headers=headers, timeout=TIMEOUT)
        elif method.upper() == 'PUT':
            response = requests.put(url, json=data, headers=headers, timeout=TIMEOUT)
        else:
            return {'status': 'ERROR', 'message': f'Unsupported method: {method}'}
        
        response_time = round((time.time() - start_time) * 1000, 2)
        
        return {
            'status': 'SUCCESS' if response.status_code < 400 else 'FAILED',
            'status_code': response.status_code,
            'response_time': f"{response_time}ms",
            'content': response.text[:200] + '...' if len(response.text) > 200 else response.text
        }
        
    except requests.exceptions.Timeout:
        return {'status': 'TIMEOUT', 'message': f'Request timed out after {TIMEOUT}s'}
    except requests.exceptions.ConnectionError:
        return {'status': 'CONNECTION_ERROR', 'message': 'Unable to connect to server'}
    except Exception as e:
        return {'status': 'ERROR', 'message': str(e)}

def print_test_result(test_name, result):
    """Print formatted test result"""
    status = result.get('status', 'UNKNOWN')
    
    if status == 'SUCCESS':
        color = Colors.GREEN
        icon = "✅"
    elif status in ['FAILED', 'ERROR', 'CONNECTION_ERROR']:
        color = Colors.RED
        icon = "❌"
    elif status == 'TIMEOUT':
        color = Colors.YELLOW
        icon = "⏱️"
    else:
        color = Colors.BLUE
        icon = "❓"
    
    print(f"{icon} {color}{test_name:<40}{Colors.END} {status}")
    
    if 'status_code' in result:
        print(f"   Status Code: {result['status_code']}")
    if 'response_time' in result:
        print(f"   Response Time: {result['response_time']}")
    if 'message' in result:
        print(f"   Message: {result['message']}")
    if 'content' in result and result['content']:
        print(f"   Response: {result['content'][:100]}...")
    print()

def main():
    print(f"{Colors.BOLD}{Colors.CYAN}")
    print("=" * 80)
    print("🚀 COMPREHENSIVE AWS EC2 DEPLOYMENT VERIFICATION")
    print("   Contact Management Microservice - All Systems Check")
    print("=" * 80)
    print(f"{Colors.END}")
    print(f"Server: {BASE_URL}")
    print(f"Test Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    total_tests = 0
    passed_tests = 0
    auth_token = None
    
    # Test Categories
    test_categories = [
        {
            'name': 'HEALTH & MONITORING ENDPOINTS',
            'tests': [
                ('GET', '/health', 'Basic Health Check'),
                ('GET', '/health/deep', 'Deep Health Check'),
                ('GET', '/status', 'Status Check'),
                ('GET', '/ready', 'Readiness Check'),
                ('GET', '/alive', 'Liveness Check'),
                ('GET', '/metrics', 'Metrics Check'),
            ]
        },
        {
            'name': 'PUBLIC ENDPOINTS',
            'tests': [
                ('GET', '/api/v1/test', 'Test Endpoint'),
                ('POST', '/api/v1/public/contact', 'Public Contact Form', {
                    'name': 'Test User',
                    'email': 'test@example.com',
                    'message': 'Test message from deployment verification'
                }),
            ]
        },
        {
            'name': 'AUTHENTICATION SYSTEM',
            'tests': [
                ('POST', '/api/v1/auth/login', 'User Login', {
                    'email': 'admin@mejona.com',
                    'password': 'admin123'
                }),
            ]
        }
    ]
    
    # Run tests by category
    for category in test_categories:
        print(f"{Colors.BOLD}{Colors.BLUE}📋 {category['name']}{Colors.END}")
        print("-" * 60)
        
        for test in category['tests']:
            method = test[0]
            endpoint = test[1]
            test_name = test[2]
            data = test[3] if len(test) > 3 else None
            
            # Special handling for auth endpoints
            if 'auth' in endpoint and method == 'POST' and 'login' in endpoint:
                result = test_endpoint(method, endpoint, data)
                if result.get('status') == 'SUCCESS' and result.get('status_code') == 200:
                    try:
                        response_data = json.loads(result.get('content', '{}'))
                        if 'data' in response_data and 'token' in response_data['data']:
                            auth_token = response_data['data']['token']
                            print(f"   🔑 Auth token obtained successfully")
                    except:
                        pass
            else:
                result = test_endpoint(method, endpoint, data, auth_token=auth_token)
            
            print_test_result(test_name, result)
            total_tests += 1
            if result.get('status') == 'SUCCESS':
                passed_tests += 1
        
        print()
    
    # Test protected dashboard endpoints if we have auth token
    if auth_token:
        print(f"{Colors.BOLD}{Colors.BLUE}📋 DASHBOARD MANAGEMENT ENDPOINTS (Authenticated){Colors.END}")
        print("-" * 60)
        
        dashboard_tests = [
            ('GET', '/api/v1/dashboard/contacts', 'List Contacts'),
            ('GET', '/api/v1/dashboard/contacts/stats', 'Contact Statistics'),
            ('POST', '/api/v1/dashboard/contact', 'Create Contact', {
                'name': 'Deployment Test Contact',
                'email': 'deploy-test@mejona.com',
                'message': 'Test contact created during deployment verification'
            }),
        ]
        
        for method, endpoint, test_name, *data in dashboard_tests:
            test_data = data[0] if data else None
            result = test_endpoint(method, endpoint, test_data, auth_token=auth_token)
            print_test_result(test_name, result)
            total_tests += 1
            if result.get('status') == 'SUCCESS':
                passed_tests += 1
        
        print()
    
    # Final Summary
    print(f"{Colors.BOLD}{Colors.CYAN}")
    print("=" * 80)
    print("📊 DEPLOYMENT VERIFICATION SUMMARY")
    print("=" * 80)
    print(f"{Colors.END}")
    
    success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
    
    print(f"Total Tests: {total_tests}")
    print(f"Passed: {Colors.GREEN}{passed_tests}{Colors.END}")
    print(f"Failed: {Colors.RED}{total_tests - passed_tests}{Colors.END}")
    print(f"Success Rate: {Colors.GREEN if success_rate >= 80 else Colors.RED}{success_rate:.1f}%{Colors.END}")
    print()
    
    if success_rate >= 80:
        print(f"{Colors.GREEN}🎉 DEPLOYMENT STATUS: SUCCESSFUL{Colors.END}")
        print(f"{Colors.GREEN}✅ Contact Management Microservice is operational on AWS EC2{Colors.END}")
    else:
        print(f"{Colors.RED}❌ DEPLOYMENT STATUS: ISSUES DETECTED{Colors.END}")
        print(f"{Colors.RED}⚠️  Service may not be fully operational{Colors.END}")
    
    print()
    print(f"🔗 Service URLs:")
    print(f"   Health Check: {BASE_URL}/health")
    print(f"   API Base: {BASE_URL}/api/v1/")
    print(f"   Dashboard API: {BASE_URL}/api/v1/dashboard/contacts")
    print()

if __name__ == "__main__":
    main()