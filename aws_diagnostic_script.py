#!/usr/bin/env python3
"""
AWS EC2 Diagnostic Script for Contact Management Microservice
Helps identify common issues causing connectivity problems
"""

import subprocess
import time
import socket
import requests
from datetime import datetime

class AWSInstanceDiagnostic:
    def __init__(self, instance_ip="65.1.94.25", ssh_key_path="D:\\Mejona Workspace\\Mejona Cred\\AWS\\mejona.pem"):
        self.instance_ip = instance_ip
        self.ssh_key_path = ssh_key_path
        self.ssh_user = "ubuntu"
        
    def test_network_connectivity(self):
        """Test basic network connectivity to the instance"""
        print("=" * 60)
        print("NETWORK CONNECTIVITY TESTS")
        print("=" * 60)
        
        # Test ping
        print("1. Testing ICMP (ping)...")
        try:
            result = subprocess.run(['ping', '-n', '3', self.instance_ip], 
                                  capture_output=True, text=True, timeout=15)
            if "TTL=" in result.stdout:
                print("   ✅ ICMP: Responsive")
            else:
                print("   ❌ ICMP: No response")
                print(f"   Output: {result.stdout[:100]}")
        except Exception as e:
            print(f"   ❌ ICMP: Error - {e}")
        
        # Test port connectivity
        ports_to_test = [22, 80, 8081]
        for port in ports_to_test:
            print(f"2.{port}. Testing port {port}...")
            try:
                sock = socket.create_connection((self.instance_ip, port), timeout=5)
                sock.close()
                print(f"   ✅ Port {port}: Open")
            except socket.timeout:
                print(f"   ❌ Port {port}: Timeout")
            except ConnectionRefusedError:
                print(f"   ❌ Port {port}: Connection refused")
            except Exception as e:
                print(f"   ❌ Port {port}: Error - {e}")
        
        print()
    
    def test_ssh_connectivity(self):
        """Test SSH connectivity and basic commands"""
        print("=" * 60)
        print("SSH CONNECTIVITY TESTS")
        print("=" * 60)
        
        ssh_commands = [
            ("Basic SSH Test", "echo 'SSH Connection: SUCCESS'"),
            ("System Uptime", "uptime"),
            ("Memory Usage", "free -h"),
            ("Disk Usage", "df -h /"),
            ("Service Status", "systemctl is-active contact-service || echo 'Service not active'"),
            ("Process Check", "ps aux | grep contact-service | grep -v grep || echo 'No process found'"),
            ("Port Check", "netstat -tlnp | grep 8081 || echo 'Port 8081 not listening'"),
        ]
        
        for test_name, command in ssh_commands:
            print(f"Testing: {test_name}")
            try:
                ssh_cmd = [
                    'ssh', '-i', self.ssh_key_path,
                    '-o', 'ConnectTimeout=10',
                    '-o', 'StrictHostKeyChecking=no',
                    f'{self.ssh_user}@{self.instance_ip}',
                    command
                ]
                
                result = subprocess.run(ssh_cmd, capture_output=True, text=True, timeout=15)
                
                if result.returncode == 0:
                    print(f"   ✅ {test_name}: Success")
                    print(f"   Output: {result.stdout.strip()}")
                else:
                    print(f"   ❌ {test_name}: Failed (code: {result.returncode})")
                    print(f"   Error: {result.stderr.strip()}")
                    
            except subprocess.TimeoutExpired:
                print(f"   ⏱️ {test_name}: Timeout")
            except Exception as e:
                print(f"   ❌ {test_name}: Error - {e}")
            
            print()
    
    def test_http_endpoints(self):
        """Test HTTP service endpoints"""
        print("=" * 60)
        print("HTTP SERVICE TESTS")
        print("=" * 60)
        
        base_url = f"http://{self.instance_ip}:8081"
        endpoints = [
            "/health",
            "/status",
            "/ready",
            "/alive",
            "/api/v1/test"
        ]
        
        for endpoint in endpoints:
            print(f"Testing: {endpoint}")
            try:
                response = requests.get(f"{base_url}{endpoint}", timeout=5)
                print(f"   ✅ Status: {response.status_code}")
                print(f"   Response time: {response.elapsed.total_seconds():.2f}s")
                if response.text:
                    print(f"   Content: {response.text[:100]}...")
            except requests.exceptions.Timeout:
                print(f"   ⏱️ Timeout (5s)")
            except requests.exceptions.ConnectionError:
                print(f"   ❌ Connection Error")
            except Exception as e:
                print(f"   ❌ Error: {e}")
            print()
    
    def check_common_issues(self):
        """Check for common issues that cause service problems"""
        print("=" * 60)
        print("COMMON ISSUES CHECK")
        print("=" * 60)
        
        checks = [
            ("Memory Usage", "free | awk 'NR==2{printf \"Memory Usage: %s/%sMB (%.2f%%)\\n\", $3,$2,$3*100/$2 }'"),
            ("Disk Usage", "df -h / | awk 'NR==2{print \"Disk Usage: \" $5 \" of \" $2 \" used\"}'"),
            ("CPU Load", "uptime | awk '{print \"Load Average: \" $(NF-2) \" \" $(NF-1) \" \" $NF}'"),
            ("Service Logs", "journalctl -u contact-service --no-pager -n 5 || echo 'No service logs found'"),
            ("System Errors", "journalctl --no-pager -p err -n 5 || echo 'No recent errors'"),
            ("OOM Kills", "dmesg | grep -i 'killed process' | tail -3 || echo 'No OOM kills found'"),
        ]
        
        for check_name, command in checks:
            print(f"Checking: {check_name}")
            try:
                ssh_cmd = [
                    'ssh', '-i', self.ssh_key_path,
                    '-o', 'ConnectTimeout=10',
                    '-o', 'StrictHostKeyChecking=no',
                    f'{self.ssh_user}@{self.instance_ip}',
                    command
                ]
                
                result = subprocess.run(ssh_cmd, capture_output=True, text=True, timeout=15)
                
                if result.returncode == 0:
                    print(f"   ✅ {result.stdout.strip()}")
                else:
                    print(f"   ❌ Failed to check {check_name}")
                    
            except Exception as e:
                print(f"   ❌ Error checking {check_name}: {e}")
            print()
    
    def generate_diagnosis_report(self):
        """Generate a comprehensive diagnosis report"""
        print("\n" + "=" * 80)
        print("AWS EC2 INSTANCE DIAGNOSTIC REPORT")
        print("=" * 80)
        print(f"Instance IP: {self.instance_ip}")
        print(f"Timestamp: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"SSH Key: {self.ssh_key_path}")
        print()
        
        self.test_network_connectivity()
        self.test_ssh_connectivity()
        self.test_http_endpoints()
        self.check_common_issues()
        
        print("=" * 80)
        print("RECOMMENDED ACTIONS")
        print("=" * 80)
        print("Based on the diagnostic results above:")
        print()
        print("1. If SSH timeouts: Instance may be stopped/terminated or high CPU load")
        print("2. If HTTP timeouts: Service not running or crashed")
        print("3. If memory >90%: Restart instance or service")
        print("4. If disk >90%: Clean logs and temp files")
        print("5. If OOM kills found: Increase instance size or optimize service")
        print()
        print("Next steps:")
        print("- Check AWS Console for instance status")
        print("- Review CloudWatch metrics for resource usage")
        print("- Consider restarting instance or redeploying service")
        print()

def main():
    diagnostic = AWSInstanceDiagnostic()
    diagnostic.generate_diagnosis_report()

if __name__ == "__main__":
    main()