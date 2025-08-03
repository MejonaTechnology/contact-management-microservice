#!/usr/bin/env python3
"""
Simple AWS EC2 Contact Management Microservice Deployment Test
"""

import requests
import json
import time
from datetime import datetime

BASE_URL = "http://65.1.94.25:8081"
TIMEOUT = 10

def test_endpoint(method, endpoint, data=None, headers=None, auth_token=None):
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
        
        response_time = round((time.time() - start_time) * 1000, 2)
        
        return {
            'status': 'SUCCESS' if response.status_code < 400 else 'FAILED',
            'status_code': response.status_code,
            'response_time': f"{response_time}ms",
            'content': response.text[:200]
        }
        
    except requests.exceptions.Timeout:
        return {'status': 'TIMEOUT', 'message': f'Request timed out after {TIMEOUT}s'}
    except requests.exceptions.ConnectionError:
        return {'status': 'CONNECTION_ERROR', 'message': 'Unable to connect to server'}
    except Exception as e:
        return {'status': 'ERROR', 'message': str(e)}

def main():
    print("=" * 80)
    print("AWS EC2 DEPLOYMENT VERIFICATION")
    print("Contact Management Microservice - Systems Check")
    print("=" * 80)
    print(f"Server: {BASE_URL}")
    print(f"Test Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    total_tests = 0
    passed_tests = 0
    auth_token = None
    
    # Test 1: Basic Health Check
    print("1. Testing Basic Health Check...")
    result = test_endpoint('GET', '/health')
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    print(f"   Time: {result.get('response_time', 'N/A')}")
    if result.get('content'):
        print(f"   Response: {result['content'][:100]}...")
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 2: Deep Health Check
    print("2. Testing Deep Health Check...")
    result = test_endpoint('GET', '/health/deep')
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 3: Status Check
    print("3. Testing Status Endpoint...")
    result = test_endpoint('GET', '/status')
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 4: Test Endpoint
    print("4. Testing API Test Endpoint...")
    result = test_endpoint('GET', '/api/v1/test')
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 5: Public Contact Form
    print("5. Testing Public Contact Form...")
    contact_data = {
        'name': 'Test User',
        'email': 'test@example.com',
        'message': 'Test message from deployment verification'
    }
    result = test_endpoint('POST', '/api/v1/public/contact', contact_data)
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 6: Authentication
    print("6. Testing User Authentication...")
    auth_data = {
        'email': 'admin@mejona.com',
        'password': 'admin123'
    }
    result = test_endpoint('POST', '/api/v1/auth/login', auth_data)
    print(f"   Status: {result.get('status')}")
    print(f"   Code: {result.get('status_code', 'N/A')}")
    
    if result.get('status') == 'SUCCESS' and result.get('status_code') == 200:
        try:
            response_data = json.loads(result.get('content', '{}'))
            if 'data' in response_data and 'token' in response_data['data']:
                auth_token = response_data['data']['token']
                print("   Auth token obtained successfully")
        except:
            pass
    print()
    total_tests += 1
    if result.get('status') == 'SUCCESS':
        passed_tests += 1
    
    # Test 7: Dashboard Contacts (if authenticated)
    if auth_token:
        print("7. Testing Dashboard Contacts (Authenticated)...")
        result = test_endpoint('GET', '/api/v1/dashboard/contacts', auth_token=auth_token)
        print(f"   Status: {result.get('status')}")
        print(f"   Code: {result.get('status_code', 'N/A')}")
        print()
        total_tests += 1
        if result.get('status') == 'SUCCESS':
            passed_tests += 1
        
        # Test 8: Contact Statistics
        print("8. Testing Contact Statistics...")
        result = test_endpoint('GET', '/api/v1/dashboard/contacts/stats', auth_token=auth_token)
        print(f"   Status: {result.get('status')}")
        print(f"   Code: {result.get('status_code', 'N/A')}")
        print()
        total_tests += 1
        if result.get('status') == 'SUCCESS':
            passed_tests += 1
    
    # Summary
    print("=" * 80)
    print("DEPLOYMENT VERIFICATION SUMMARY")
    print("=" * 80)
    
    success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
    
    print(f"Total Tests: {total_tests}")
    print(f"Passed: {passed_tests}")
    print(f"Failed: {total_tests - passed_tests}")
    print(f"Success Rate: {success_rate:.1f}%")
    print()
    
    if success_rate >= 80:
        print("DEPLOYMENT STATUS: SUCCESSFUL")
        print("Contact Management Microservice is operational on AWS EC2")
    else:
        print("DEPLOYMENT STATUS: ISSUES DETECTED")
        print("Service may not be fully operational")
    
    print()
    print("Service URLs:")
    print(f"  Health Check: {BASE_URL}/health")
    print(f"  API Base: {BASE_URL}/api/v1/")
    print(f"  Dashboard API: {BASE_URL}/api/v1/dashboard/contacts")

if __name__ == "__main__":
    main()