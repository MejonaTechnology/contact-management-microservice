#!/usr/bin/env python3
"""
Simple AWS EC2 Diagnostic for Contact Management Microservice
"""

import subprocess
import socket
import requests
from datetime import datetime

def test_port_connectivity(host, port, timeout=5):
    """Test if a port is open"""
    try:
        sock = socket.create_connection((host, port), timeout=timeout)
        sock.close()
        return True
    except:
        return False

def test_ssh_command(host, key_path, command, timeout=10):
    """Execute SSH command and return result"""
    try:
        ssh_cmd = [
            'ssh', '-i', key_path,
            '-o', f'ConnectTimeout={timeout}',
            '-o', 'StrictHostKeyChecking=no',
            f'ubuntu@{host}',
            command
        ]
        
        result = subprocess.run(ssh_cmd, capture_output=True, text=True, timeout=timeout+5)
        return result.returncode == 0, result.stdout.strip(), result.stderr.strip()
    except:
        return False, "", "Timeout or connection error"

def test_http_endpoint(url, timeout=5):
    """Test HTTP endpoint"""
    try:
        response = requests.get(url, timeout=timeout)
        return True, response.status_code, response.text[:100]
    except:
        return False, 0, "Connection failed"

def main():
    instance_ip = "65.1.94.25"
    ssh_key = "D:\\Mejona Workspace\\Mejona Cred\\AWS\\mejona.pem"
    
    print("AWS EC2 Instance Diagnostic Report")
    print("=" * 50)
    print(f"Instance: {instance_ip}")
    print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    # Test network connectivity
    print("NETWORK CONNECTIVITY:")
    print("-" * 20)
    
    # Test ping
    try:
        result = subprocess.run(['ping', '-n', '2', instance_ip], 
                              capture_output=True, text=True, timeout=10)
        if "TTL=" in result.stdout:
            print("PING: SUCCESS")
        else:
            print("PING: FAILED")
    except:
        print("PING: ERROR")
    
    # Test ports
    ports = [22, 80, 8081]
    for port in ports:
        if test_port_connectivity(instance_ip, port):
            print(f"PORT {port}: OPEN")
        else:
            print(f"PORT {port}: CLOSED/TIMEOUT")
    
    print()
    
    # Test SSH connectivity
    print("SSH CONNECTIVITY:")
    print("-" * 17)
    
    ssh_tests = [
        ("Basic SSH", "echo 'SSH_SUCCESS'"),
        ("Uptime", "uptime"),
        ("Memory", "free -h | head -2"),
        ("Disk", "df -h / | tail -1"),
        ("Service Status", "systemctl is-active contact-service 2>/dev/null || echo 'INACTIVE'"),
        ("Process Check", "ps aux | grep contact-service | grep -v grep || echo 'NO_PROCESS'"),
        ("Port 8081", "netstat -tlnp | grep 8081 || echo 'NOT_LISTENING'")
    ]
    
    for test_name, command in ssh_tests:
        success, stdout, stderr = test_ssh_command(instance_ip, ssh_key, command)
        if success:
            print(f"{test_name}: {stdout}")
        else:
            print(f"{test_name}: FAILED - {stderr}")
    
    print()
    
    # Test HTTP endpoints
    print("HTTP SERVICE:")
    print("-" * 13)
    
    base_url = f"http://{instance_ip}:8081"
    endpoints = ["/health", "/status", "/api/v1/test"]
    
    for endpoint in endpoints:
        success, status_code, content = test_http_endpoint(f"{base_url}{endpoint}")
        if success:
            print(f"{endpoint}: HTTP {status_code} - {content[:50]}...")
        else:
            print(f"{endpoint}: FAILED")
    
    print()
    
    # Check for common issues
    print("SYSTEM HEALTH:")
    print("-" * 14)
    
    health_checks = [
        ("Memory Usage", "free | awk 'NR==2{printf \"%.1f%%\", $3*100/$2}'"),
        ("Disk Usage", "df / | awk 'NR==2{print $5}'"),
        ("Load Average", "uptime | awk '{print $(NF-2) $(NF-1) $NF}'"),
        ("Recent Errors", "journalctl -p err --since '1 hour ago' --no-pager -q | wc -l"),
        ("OOM Events", "dmesg | grep -c 'Out of memory' || echo '0'")
    ]
    
    for check_name, command in health_checks:
        success, stdout, stderr = test_ssh_command(instance_ip, ssh_key, command)
        if success:
            print(f"{check_name}: {stdout}")
        else:
            print(f"{check_name}: UNABLE_TO_CHECK")
    
    print()
    print("DIAGNOSIS SUMMARY:")
    print("-" * 17)
    print("1. Check if ports 22 and 8081 are OPEN")
    print("2. Check if SSH commands work")
    print("3. Check if contact-service is ACTIVE")
    print("4. Check if port 8081 is LISTENING")
    print("5. Check HTTP endpoints respond")
    print()
    print("NEXT STEPS:")
    print("- If SSH fails: Instance may be stopped or overloaded")
    print("- If service INACTIVE: Run 'sudo systemctl start contact-service'")
    print("- If port not listening: Service crashed or not running")
    print("- If HTTP fails: Service needs restart or redeployment")

if __name__ == "__main__":
    main()