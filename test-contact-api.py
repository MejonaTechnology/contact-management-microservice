#!/usr/bin/env python3
"""
Test script for Contact Microservice API
Tests all endpoints and verifies functionality
"""

import requests
import json
import time
import sys
from datetime import datetime

# Configuration
BASE_URL = "http://localhost:8081"
HEALTH_URL = f"{BASE_URL}/health"
CONTACT_URL = f"{BASE_URL}/api/v1/public/contact"
TEST_URL = f"{BASE_URL}/api/v1/test"

def test_health_check():
    """Test health check endpoint"""
    print("🏥 Testing health check endpoint...")
    try:
        response = requests.get(HEALTH_URL, timeout=10)
        if response.status_code == 200:
            data = response.json()
            print(f"✅ Health check passed: {data.get('message', 'OK')}")
            return True
        else:
            print(f"❌ Health check failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Health check error: {e}")
        return False

def test_test_endpoint():
    """Test the test endpoint"""
    print("🧪 Testing test endpoint...")
    try:
        response = requests.get(TEST_URL, timeout=10)
        if response.status_code == 200:
            data = response.json()
            print(f"✅ Test endpoint passed: {data.get('message', 'OK')}")
            return True
        else:
            print(f"❌ Test endpoint failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Test endpoint error: {e}")
        return False

def test_contact_submission():
    """Test contact form submission"""
    print("📧 Testing contact submission...")
    
    # Test data in the format expected by the microservice
    test_data = {
        "name": "Test User",
        "email": f"test.user.{int(time.time())}@example.com",
        "phone": "+1234567890",
        "subject": "API Test Submission",
        "message": "This is a test contact submission from the API test script.",
        "source": "api_test",
        "website": ""  # Honeypot field - should be empty
    }
    
    print(f"📤 Submitting test contact: {test_data['email']}")
    
    try:
        response = requests.post(
            CONTACT_URL,
            json=test_data,
            headers={"Content-Type": "application/json"},
            timeout=30
        )
        
        print(f"📬 Response status: {response.status_code}")
        
        if response.status_code in [200, 201]:
            try:
                data = response.json()
                print(f"✅ Contact submission successful!")
                print(f"📋 Response: {json.dumps(data, indent=2)}")
                
                # Check if we got a contact ID
                if data.get('data') and data['data'].get('contact_id'):
                    contact_id = data['data']['contact_id']
                    print(f"🆔 Contact ID: {contact_id}")
                
                return True
            except json.JSONDecodeError:
                print(f"✅ Contact submission successful (non-JSON response)")
                print(f"📋 Response text: {response.text}")
                return True
        else:
            print(f"❌ Contact submission failed: {response.status_code}")
            try:
                error_data = response.json()
                print(f"📋 Error response: {json.dumps(error_data, indent=2)}")
            except:
                print(f"📋 Error text: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ Contact submission error: {e}")
        return False

def test_invalid_contact_submission():
    """Test contact submission with invalid data"""
    print("🚫 Testing invalid contact submission...")
    
    # Invalid data (missing required fields)
    invalid_data = {
        "email": "invalid-email",  # Invalid email format
        "message": ""  # Empty message
    }
    
    try:
        response = requests.post(
            CONTACT_URL,
            json=invalid_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if response.status_code == 400:
            print("✅ Invalid data correctly rejected")
            return True
        else:
            print(f"❌ Invalid data not rejected properly: {response.status_code}")
            return False
            
    except Exception as e:
        print(f"❌ Invalid submission test error: {e}")
        return False

def test_honeypot_detection():
    """Test spam detection via honeypot field"""
    print("🍯 Testing honeypot spam detection...")
    
    # Spam data (honeypot field filled)
    spam_data = {
        "name": "Spam Bot",
        "email": "spam@example.com",
        "phone": "+1234567890",
        "subject": "Spam Message",
        "message": "This is spam content",
        "source": "spam_test",
        "website": "http://spam-site.com"  # Honeypot field filled (indicates spam)
    }
    
    try:
        response = requests.post(
            CONTACT_URL,
            json=spam_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if response.status_code == 400:
            print("✅ Spam correctly detected and rejected")
            return True
        else:
            print(f"❌ Spam not detected properly: {response.status_code}")
            return False
            
    except Exception as e:
        print(f"❌ Honeypot test error: {e}")
        return False

def run_all_tests():
    """Run all tests"""
    print("🚀 Starting Contact Microservice API Tests")
    print("=" * 50)
    
    tests = [
        ("Health Check", test_health_check),
        ("Test Endpoint", test_test_endpoint),
        ("Valid Contact Submission", test_contact_submission),
        ("Invalid Contact Submission", test_invalid_contact_submission),
        ("Honeypot Detection", test_honeypot_detection),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n📋 Running: {test_name}")
        print("-" * 30)
        
        if test_func():
            passed += 1
            print(f"✅ {test_name}: PASSED")
        else:
            print(f"❌ {test_name}: FAILED")
    
    print("\n" + "=" * 50)
    print(f"🎯 Test Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 ALL TESTS PASSED! Contact microservice is working correctly!")
        return True
    else:
        print(f"⚠️  {total - passed} tests failed. Please check the microservice.")
        return False

if __name__ == "__main__":
    print(f"Contact Microservice API Test Suite")
    print(f"Testing against: {BASE_URL}")
    print(f"Timestamp: {datetime.now()}")
    print()
    
    # Wait for service to be ready
    print("⏳ Waiting for service to be ready...")
    time.sleep(2)
    
    success = run_all_tests()
    
    if success:
        sys.exit(0)
    else:
        sys.exit(1)