# AWS Console Investigation Checklist
## Contact Management Microservice - Instance: 65.1.94.25

### 🔐 LOGIN TO AWS CONSOLE
**URL**: https://console.aws.amazon.com/
**Account**: Your AWS account credentials
**Region**: Check the correct region where instance was deployed

---

### 📍 STEP 1: LOCATE THE INSTANCE

1. **Navigate to EC2 Dashboard**
   - Services → EC2 → Instances (Running)
   
2. **Find Your Instance**
   - Search by IP: `65.1.94.25`
   - Or search by Name tag: `contact-management-microservice`
   - Or filter by instance type/launch time

3. **Record Instance Details**
   - Instance ID: `i-xxxxxxxxx`
   - Instance Type: (e.g., t2.micro, t3.small)
   - Launch Time: When it was created
   - Current State: **[CRITICAL - CHECK THIS]**

---

### 🚨 STEP 2: INSTANCE STATE CHECK

**Look for the "Instance state" column:**

#### ✅ If State = "Running"
- Instance is supposed to be operational
- Continue to Step 3 (Status Checks)

#### ❌ If State = "Stopped" 
- **CAUSE FOUND**: Instance was stopped
- **ACTION**: Select instance → Actions → Instance State → Start
- **WAIT**: 2-3 minutes for startup
- **THEN**: Test connectivity again

#### ❌ If State = "Stopping" or "Pending"
- Instance is in transition
- **ACTION**: Wait 5 minutes and refresh
- **IF STUCK**: Force stop and restart

#### ❌ If State = "Terminated"
- **CRITICAL**: Instance was permanently deleted
- **ACTION**: Need to launch new instance and redeploy
- **USE**: Existing deployment scripts

#### ❌ If State = "Shutting-down"
- Instance is being terminated
- **ACTION**: Launch new instance immediately

---

### 🔍 STEP 3: STATUS CHECKS (If Running)

**Click on the instance → Status checks tab**

#### System Status Check
- ✅ **Passed**: AWS infrastructure is healthy
- ❌ **Failed**: AWS hardware/network issues
  - **ACTION**: Stop and start instance (not reboot)

#### Instance Status Check  
- ✅ **Passed**: Instance OS is responding
- ❌ **Failed**: Instance is frozen, out of memory, or crashed
  - **ACTION**: Reboot instance or investigate further

---

### 📊 STEP 4: MONITORING METRICS

**Click on instance → Monitoring tab**

#### Check Last 24 Hours:

1. **CPU Utilization**
   - Normal: <80%
   - High: >90% (indicates overload/infinite loop)
   - Flat 0%: Instance may be crashed

2. **Memory Utilization** (if CloudWatch agent installed)
   - Normal: <85%
   - High: >95% (OOM kills likely)

3. **Disk Space**
   - Normal: <85% used
   - High: >95% (can cause crashes)

4. **Network In/Out**
   - Unusual spikes may indicate issues

---

### 🔧 STEP 5: COMMON ISSUE PATTERNS

#### Pattern 1: High Memory → OOM Kill
**Symptoms**: Instance running, status checks failed
**Metrics**: Memory >95%, CPU may spike then drop to 0%
**Solution**: Restart instance, optimize service memory usage

#### Pattern 2: Disk Full
**Symptoms**: Services crash, can't write logs
**Metrics**: Disk usage >95%
**Solution**: Connect via Session Manager, clean logs/temp files

#### Pattern 3: Infinite Loop
**Symptoms**: High CPU 100%, unresponsive
**Metrics**: CPU consistently 100%
**Solution**: Reboot instance, fix code issue

#### Pattern 4: Instance Stopped by AWS
**Symptoms**: Instance state = stopped
**Causes**: 
- Billing issues (payment failed)
- Spot instance interruption
- Scheduled maintenance
- Manual stop by user
**Solution**: Start instance

---

### 🛠️ STEP 6: IMMEDIATE ACTIONS

#### If Instance is Stopped:
```
1. Select instance
2. Actions → Instance State → Start
3. Wait 2-3 minutes
4. Test: ssh -i "mejona.pem" ubuntu@65.1.94.25 "echo success"
```

#### If Instance is Running but Unresponsive:
```
1. Try EC2 Instance Connect (browser SSH)
2. Or use Session Manager (if configured)
3. Actions → Instance State → Reboot
4. Wait 3-5 minutes
5. Test connectivity
```

#### If Status Checks Failed:
```
1. Actions → Instance State → Stop
2. Wait until fully stopped
3. Actions → Instance State → Start
4. Monitor status checks
```

---

### 💻 STEP 7: ALTERNATIVE ACCESS METHODS

#### EC2 Instance Connect (if SSH fails)
1. Select instance
2. Actions → Connect
3. EC2 Instance Connect tab
4. Connect (opens browser terminal)

#### Session Manager (if configured)
1. Services → Systems Manager
2. Session Manager
3. Start session
4. Select your instance

---

### 📋 STEP 8: RECOVERY COMMANDS

**Once you can access the instance:**

```bash
# Check service status
sudo systemctl status contact-service

# Check if service exists
ls -la /opt/mejona/contact-management-microservice/

# Check system resources
free -h
df -h
top

# Check recent errors
sudo journalctl -u contact-service -n 50

# Restart service
sudo systemctl restart contact-service

# Check if it's listening
netstat -tlnp | grep 8081

# Test locally
curl http://localhost:8081/health
```

---

### 🚀 STEP 9: FULL REDEPLOYMENT (If Needed)

**If service is completely broken:**

1. **Use existing deployment script**: 
   - Run `DEPLOY_NOW.bat` from local machine
   
2. **Manual redeployment**:
   ```bash
   cd /opt/mejona
   sudo rm -rf contact-management-microservice
   git clone https://github.com/MejonaTechnology/contact-management-microservice.git
   cd contact-management-microservice
   go build -o contact-service cmd/server/main_fixed.go
   sudo systemctl restart contact-service
   ```

---

### 📝 REPORT YOUR FINDINGS

**After checking, report back:**

1. **Instance State**: Running/Stopped/Terminated
2. **Status Checks**: Passed/Failed (both system and instance)
3. **Metrics**: Any unusual CPU/Memory/Disk patterns
4. **Access**: Can you connect via SSH/Instance Connect?
5. **Service Status**: Is contact-service running?

**This information will determine the next steps for resolution.**