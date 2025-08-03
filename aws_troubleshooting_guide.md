# AWS EC2 Troubleshooting Guide
## Contact Management Microservice - Instance: 65.1.94.25

### 🔍 DIAGNOSTIC STEPS FOR AWS CONSOLE

#### Step 1: Instance Status Check
**Navigate to**: AWS Console → EC2 → Instances

**Check for**:
- Instance State: `running`, `stopped`, `terminated`, or `pending`
- Status Checks: System Status and Instance Status
- Instance Health: Any failed checks
- CPU Utilization: High usage indicating resource exhaustion
- Memory Usage: RAM consumption levels

**Expected Issues**:
```
❌ Instance State: stopped/terminated
❌ Status Checks: 1/2 checks failed (system/instance)
❌ CPU Utilization: >90% sustained
❌ Memory: >95% usage causing OOM kills
```

#### Step 2: System Log Analysis
**Navigate to**: Instance → Actions → Monitor and troubleshoot → Get system log

**Look for**:
```bash
# Memory issues
kernel: Out of memory: Kill process
kernel: Killed process [PID] (contact-service)

# Disk space issues  
kernel: No space left on device

# SSH daemon issues
sshd: fatal: Cannot bind any address

# Service crashes
systemd: contact-service.service: Main process exited
```

#### Step 3: Instance Metrics Review
**Navigate to**: Instance → Monitoring tab

**Critical Metrics**:
- **CPU Utilization**: Should be <80%
- **Memory Utilization**: Should be <85% 
- **Disk Space**: Should have >15% free
- **Network In/Out**: Check for unusual traffic

#### Step 4: Security Group Verification
**Navigate to**: Instance → Security tab → Security groups

**Required Rules**:
```
Inbound Rules:
- Type: SSH, Port: 22, Source: 0.0.0.0/0
- Type: Custom TCP, Port: 8081, Source: 0.0.0.0/0
- Type: HTTP, Port: 80, Source: 0.0.0.0/0 (optional)

Outbound Rules:
- Type: All traffic, Port: All, Destination: 0.0.0.0/0
```

### 🛠️ COMMON ISSUES & SOLUTIONS

#### Issue 1: Instance Stopped/Terminated
**Symptoms**: Cannot ping, SSH timeout, HTTP timeout
**Cause**: Instance was stopped or terminated (possibly auto-scaling/billing)
**Solution**:
```bash
# If stopped - restart
Instance → Actions → Instance State → Start

# If terminated - launch new instance
Launch new EC2 instance with same configuration
```

#### Issue 2: Memory Exhaustion (OOM Killer)
**Symptoms**: SSH hangs, services crash, high memory usage
**Cause**: Go service consuming too much RAM or memory leak
**Solution**:
```bash
# Reboot instance to clear memory
Instance → Actions → Instance State → Reboot

# Or connect via EC2 Instance Connect
sudo systemctl restart contact-service
sudo systemctl status contact-service
free -h  # Check memory usage
```

#### Issue 3: Disk Space Full
**Symptoms**: Service crashes, cannot write logs, SSH issues
**Cause**: Logs filling disk, large files, or database growth
**Solution**:
```bash
# Connect via EC2 Instance Connect
df -h  # Check disk usage
sudo journalctl --vacuum-time=7d  # Clean old logs
sudo apt clean  # Clean package cache
sudo find /tmp -type f -atime +7 -delete  # Clean temp files
```

#### Issue 4: High CPU Usage
**Symptoms**: Slow responses, timeouts, system lag
**Cause**: Infinite loops, CPU-intensive operations, or resource contention
**Solution**:
```bash
# Check processes
top -o %CPU
ps aux --sort=-%cpu | head -10

# Restart service
sudo systemctl restart contact-service
```

#### Issue 5: Network/Security Group Issues
**Symptoms**: Ports inaccessible, connection refused
**Cause**: Security group misconfiguration or network ACL issues
**Solution**:
- Verify security group rules (ports 22, 8081, 80)
- Check Network ACLs
- Verify VPC/subnet configuration

### 🚀 RECOVERY PROCEDURES

#### Quick Recovery (If Instance is Running)
```bash
# Via EC2 Instance Connect or System Manager Session Manager
sudo systemctl stop contact-service
sudo systemctl start contact-service
sudo systemctl status contact-service
sudo journalctl -u contact-service -f
```

#### Full Recovery (If Major Issues)
```bash
# Option 1: Reboot instance
Instance → Actions → Instance State → Reboot

# Option 2: Stop and start (not reboot)
Instance → Actions → Instance State → Stop
Wait for stopped state
Instance → Actions → Instance State → Start
```

#### Complete Redeployment (If All Else Fails)
```bash
# Use existing deployment script
./DEPLOY_NOW.bat

# Or manual deployment
ssh -i "mejona.pem" ubuntu@65.1.94.25
cd /opt/mejona/contact-management-microservice
git pull origin main
go build -o contact-service cmd/server/main_fixed.go
sudo systemctl restart contact-service
```

### 📊 MONITORING COMMANDS

#### Instance Health Check
```bash
# System resources
free -h          # Memory usage
df -h            # Disk usage  
top              # CPU and processes
netstat -tlnp    # Open ports
systemctl status contact-service

# Service verification
curl http://localhost:8081/health
curl http://localhost:8081/status
```

#### Log Analysis
```bash
# System logs
sudo journalctl -u contact-service --since "1 hour ago"
sudo journalctl -u contact-service -f  # Follow logs
sudo tail -f /var/log/syslog

# Application logs (if any)
sudo tail -f /opt/mejona/contact-management-microservice/*.log
```

### 🎯 ACTION PLAN

#### Immediate Steps:
1. **Login to AWS Console**: https://console.aws.amazon.com/ec2/
2. **Locate Instance**: Search for IP 65.1.94.25 or instance ID
3. **Check Instance State**: Verify running/stopped/terminated
4. **Review Status Checks**: Look for any failed checks
5. **Check Metrics**: CPU, Memory, Disk usage over last 24 hours

#### If Instance is Stopped:
1. **Start Instance**: Actions → Instance State → Start
2. **Wait 2-3 minutes**: For full startup
3. **Test Connectivity**: SSH and HTTP endpoints
4. **Restart Service**: If needed

#### If Instance is Terminated:
1. **Launch New Instance**: Use same AMI and configuration
2. **Configure Security Groups**: Ensure ports 22, 8081, 80 are open
3. **Run Deployment Script**: Use existing `DEPLOY_NOW.bat`
4. **Verify All Endpoints**: Test all 20 API endpoints

#### If Instance is Running but Unresponsive:
1. **Connect via EC2 Instance Connect**: Browser-based SSH
2. **Check Resource Usage**: Memory, CPU, disk
3. **Restart Services**: SSH daemon and contact-service
4. **Review Logs**: Look for error patterns

### 📋 CHECKLIST

- [ ] Instance state verified
- [ ] Status checks reviewed  
- [ ] Resource metrics analyzed
- [ ] Security groups confirmed
- [ ] Service status checked
- [ ] Logs reviewed for errors
- [ ] Connectivity restored
- [ ] All endpoints tested

---

**Next**: After checking AWS Console, update this document with findings and execute appropriate recovery procedure.